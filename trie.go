package trie

import (
	"errors"
	"math/bits"
	"sync"
	"unicode/utf8"
)

// Trie works with latin characters, space and digits
type Trie struct {
	rw         sync.RWMutex
	characters uint64
	subTries   []*Trie
	data       interface{}
}

func (trieInst *Trie) Set(key string, value interface{}) error {
	if err := trieInst.checkKey(key); err != nil {
		return err
	}

	if value == nil {
		return errors.New("value cannot be nil")
	}

	tr := trieInst
	for _, char := range key {
		tr.rw.Lock()
		bitNum, subTrieId, subTrie := tr.getSubTrie(char)

		if subTrie != nil {
			tr.rw.Unlock()
			tr = subTrie
			continue
		}

		subTrie = &Trie{}
		tr.characters |= 1 << bitNum
		tr.subTries = append(tr.subTries, nil)
		copy(tr.subTries[subTrieId+1:], tr.subTries[subTrieId:])
		tr.subTries[subTrieId] = subTrie
		tr.rw.Unlock()
		tr = subTrie
	}

	tr.rw.Lock()
	tr.data = value
	tr.rw.Unlock()

	return nil
}

// Find an item having exact given key
func (trieInst *Trie) Get(key string) (interface{}, error) {
	if err := trieInst.checkKey(key); err != nil {
		return nil, err
	}

	tr := trieInst
	for _, char := range key {
		tr.rw.RLock()
		_, _, subTrie := tr.getSubTrie(char)
		tr.rw.RUnlock()

		if subTrie == nil {
			return nil, nil
		}

		tr = subTrie
	}

	tr.rw.RLock()
	ret := tr.data
	tr.rw.RUnlock()

	return ret, nil
}

type TraverseNode struct {
	key  string
	trie *Trie
}

func (tn *TraverseNode) getChildren() []*TraverseNode {
	trie := tn.trie
	trie.rw.RLock()
	ret, i := make([]*TraverseNode, 0, len(trie.subTries)), 0
	for bitNum := 0; bitNum < 63; bitNum++ {
		if trie.characters&(1<<bitNum) == 0 {
			continue
		}

		char := calcRune(bitNum)
		ret = append(ret, &TraverseNode{key: tn.key + char, trie: trie.subTries[i]})
		i++

		if i == len(trie.subTries) {
			break
		}
	}
	trie.rw.RUnlock()

	return ret
}

type DataForKey struct {
	Key  string
	Data interface{}
}

// Find items having given prefix
func (trieInst *Trie) GetByPrefix(prefix string, limit int) ([]DataForKey, error) {
	if err := trieInst.checkKey(prefix); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = -1
	}

	tr := trieInst
	for _, char := range prefix {
		tr.rw.RLock()
		_, _, subTrie := tr.getSubTrie(char)
		tr.rw.RUnlock()

		if subTrie == nil {
			return nil, nil
		}

		tr = subTrie
	}

	var ret []DataForKey
	tr.rw.RLock()
	if tr.data != nil {
		ret = append(ret, DataForKey{Key: prefix, Data: tr.data})

		if limit > 0 {
			limit--
		}
	}
	tr.rw.RUnlock()

	var toTraverse []*TraverseNode
	toTraverse = append(toTraverse, &TraverseNode{key: prefix, trie: tr})
	for limit != 0 && len(toTraverse) > 0 {
		toTr := toTraverse[0]

		items := toTr.getChildren()

		for _, item := range items {
			item.trie.rw.RLock()
			if item.trie.data != nil {
				ret = append(ret, DataForKey{Key: item.key, Data: item.trie.data})

				if limit > 0 {
					limit--
				}
			}
			item.trie.rw.RUnlock()

			if limit == 0 {
				return ret, nil
			}
		}

		toTraverse = append(toTraverse, items...)

		if len(toTraverse) > 0 {
			toTraverse = toTraverse[1:]
		} else {
			break
		}
	}

	return ret, nil
}

func (trieInst *Trie) Remove(key string) error {
	if err := trieInst.checkKey(key); err != nil {
		return err
	}

	runesAmount := utf8.RuneCountInString(key)
	toRemove, tr, lastIndex := make([]*Trie, 0, runesAmount), trieInst, runesAmount - 1
	var rootChar rune
	for index, char := range key {
		if len(toRemove) == 0 {
			rootChar = char
		}

		toRemove = append(toRemove, tr)

		tr.rw.RLock()
		_, _, subTrie := tr.getSubTrie(char)
		tr.rw.RUnlock()

		if subTrie == nil {
			return nil
		}

		subTrie.rw.RLock()
		if len(subTrie.subTries) > 1 || index != lastIndex && subTrie.data != nil || index == lastIndex && len(subTrie.subTries) > 0 {
			toRemove = toRemove[:0]
		}
		subTrie.rw.RUnlock()
		tr = subTrie
	}

	tr.rw.Lock()
	tr.data = nil
	tr.rw.Unlock()

	if len(toRemove) == 0 {
		return nil
	}

	for i := len(toRemove) - 1; i > 0; i-- {
		toRemove[i].rw.Lock()
		toRemove[i].characters, toRemove[i].subTries = 0, nil
		toRemove[i].rw.Unlock()
	}

	tr = toRemove[0]
	tr.rw.Lock()
	bitNum, subTrieId, _ := tr.getSubTrie(rootChar)
	if len(tr.subTries) == 1 {
		tr.subTries = nil
		tr.characters = 0
	} else {
		tr.subTries = append(tr.subTries[:subTrieId], tr.subTries[subTrieId+1:]...)
		mask := uint64(^(1 << bitNum))
		tr.characters &= mask
	}
	tr.rw.Unlock()

	return nil
}

func (trieInst *Trie) getSubTrie(char rune) (uint64, int, *Trie) {
	bitNum, _ := calcBitNum(char)
	subTrieId := trieInst.getSubTreeIndex(bitNum)

	// There is no subTrie under given character since the corresponding bit is zero
	if trieInst.characters&(1<<bitNum) == 0 {
		return bitNum, subTrieId, nil
	}

	return bitNum, subTrieId, trieInst.subTries[subTrieId]
}

// subTreeIndex is an amount of bits before bitNum
func (trieInst *Trie) getSubTreeIndex(bitNum uint64) int {
	shifted := trieInst.characters << (64 - bitNum)
	return bits.OnesCount64(shifted)
}

var predefinedChars = [123]uint64{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	62, // space has index 32 (same as its ASCII code) and uses 62nd bit
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, // digits 0-9 have indexes from 48 to 57 and use 0-9 bits
	0, 0, 0, 0, 0, 0, 0,
	// A-Z characters have indexes from 65 to 90 and use 10-35 bits
	10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35,
	0, 0, 0, 0, 0, 0,
	// a-z characters have indexes from 97 to 122 and use 36-61 bits
	36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61,
}

func calcBitNum(char rune) (uint64, bool) {
	return predefinedChars[char], true
}

var predefinedRunes = [63]string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j",
	"k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	"u", "v", "w", "x", "y", "z", "A", "B", "C", "D",
	"E", "F", "G", "H", "I", "J", "K", "L", "M", "N",
	"O", "P", "Q", "R", "S", "T", "U", "V", "W", "X",
	"Y", "Z", " ",
}

func calcRune(bitNum int) string {
	return predefinedRunes[bitNum]
}

func (trieInst *Trie) checkKey(key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	for _, char := range key {
		if int(char) >= len(predefinedChars) {
			return errors.New("key can contain only a-z, A-Z characters, digits and space")
		}
	}

	return nil
}