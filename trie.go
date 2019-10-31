package trie

import (
	"errors"
	"math/bits"
	"sync"
	"unicode/utf8"
)

// Trie works with latin characters, space and digits
type Trie struct {
	sync.RWMutex
	characters uint64
	subTries   []*Trie
	data       interface{}
}

func (trie *Trie) Set(key string, value interface{}) error {
	if err := trie.checkKey(key); err != nil {
		return err
	}

	if value == nil {
		return errors.New("value cannot be nil")
	}

	tr := trie
	for _, char := range key {
		tr.Lock()
		bitNum, subTrieId, subTrie := tr.getSubTrie(char)

		if subTrie != nil {
			tr.Unlock()
			tr = subTrie
			continue
		}

		subTrie = &Trie{}
		tr.characters |= 1 << bitNum
		tr.subTries = append(tr.subTries, nil)
		copy(tr.subTries[subTrieId+1:], tr.subTries[subTrieId:])
		tr.subTries[subTrieId] = subTrie
		tr.Unlock()
		tr = subTrie
	}

	tr.Lock()
	tr.data = value
	tr.Unlock()

	return nil
}

// Find an item having exact given key
func (trie *Trie) Get(key string) (interface{}, error) {
	if err := trie.checkKey(key); err != nil {
		return nil, err
	}

	tr := trie
	for _, char := range key {
		tr.RLock()
		_, _, subTrie := tr.getSubTrie(char)
		tr.RUnlock()

		if subTrie == nil {
			return nil, nil
		}

		tr = subTrie
	}

	tr.RLock()
	ret := tr.data
	tr.RUnlock()

	return ret, nil
}

type TraverseNode struct {
	key  string
	trie *Trie
}

func (tn *TraverseNode) getChildren() []*TraverseNode {
	trie := tn.trie
	trie.RLock()
	ret, i := make([]*TraverseNode, len(trie.subTries)), 0
	for bitNum := int32(0); bitNum < 64; bitNum++ {
		if trie.characters&(1<<bitNum) == 0 {
			continue
		}

		char := calcRune(bitNum)
		ret[i] = &TraverseNode{key: tn.key + string(char), trie: trie.subTries[i]}
		i++

		if i == len(trie.subTries) {
			break
		}
	}
	trie.RUnlock()

	return ret
}

type DataForKey struct {
	Key  string
	Data interface{}
}

// Find items having given prefix
func (trie *Trie) GetByPrefix(prefix string, limit int) ([]DataForKey, error) {
	if err := trie.checkKey(prefix); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = -1
	}

	tr := trie
	var ret []DataForKey
	for _, char := range prefix {
		tr.RLock()
		_, _, subTrie := tr.getSubTrie(char)
		tr.RUnlock()

		if subTrie == nil {
			return nil, nil
		}

		tr = subTrie
	}

	tr.RLock()
	if tr.data != nil {
		ret = append(ret, DataForKey{Key: prefix, Data: tr.data})

		if limit > 0 {
			limit--
		}
	}
	tr.RUnlock()

	var toTraverse []*TraverseNode
	toTraverse = append(toTraverse, &TraverseNode{key: prefix, trie: tr})
	for limit > 0 && len(toTraverse) > 0 {
		toTr := toTraverse[0]
		items, added := toTr.getChildren(), 0
		for _, item := range items {
			item.trie.RLock()
			if item.trie.data != nil {
				ret = append(ret, DataForKey{Key: item.key, Data: item.trie.data})
				added++
				limit--
			}
			item.trie.RUnlock()
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

func (trie *Trie) Remove(key string) error {
	if err := trie.checkKey(key); err != nil {
		return err
	}

	toRemove, tr := make([]*Trie, 0, utf8.RuneCountInString(key)), trie
	lastIndex := utf8.RuneCountInString(key) - 1
	var rootChar rune
	for index, char := range key {
		if len(toRemove) == 0 {
			tr.Lock()
			rootChar = char
		}

		toRemove = append(toRemove, tr)
		_, _, subTrie := tr.getSubTrie(char)

		if subTrie == nil {
			toRemove[0].Unlock()
			return nil
		}

		subTrie.RLock()
		if len(subTrie.subTries) > 1 || subTrie.data != nil || index == lastIndex && len(subTrie.subTries) > 0 {
			toRemove[0].Unlock()
			toRemove = toRemove[:0]
		}
		subTrie.RUnlock()
		tr = subTrie
	}

	tr.Lock()
	tr.data = nil
	tr.Unlock()

	if len(toRemove) == 0 {
		return nil
	}

	for i := len(toRemove) - 1; i > 0; i-- {
		toRemove[i].characters, toRemove[i].subTries = 0, nil
	}

	tr = toRemove[0]
	bitNum, subTrieId, _ := tr.getSubTrie(rootChar)
	if len(tr.subTries) == 1 {
		tr.subTries = nil
		tr.characters = 0
	} else {
		tr.subTries = append(tr.subTries[:subTrieId], tr.subTries[subTrieId+1:]...)
		mask := uint64(^(1<<bitNum))
		tr.characters &= mask
	}
	tr.Unlock()

	return nil
}

func (trie *Trie) getSubTrie(char rune) (uint64, int, *Trie) {
	bitNum, _ := calcBitNum(char)
	subTrieId := trie.getSubTreeIndex(bitNum)

	// There is no subTrie under given character since the corresponding bit is zero
	if trie.characters&(1<<bitNum) == 0 {
		return bitNum, subTrieId, nil
	}

	return bitNum, subTrieId, trie.subTries[subTrieId]
}

// subTreeIndex is an amount of bits before bitNum
func (trie *Trie) getSubTreeIndex(bitNum uint64) int {
	shifted := trie.characters << (64 - bitNum)
	return bits.OnesCount64(shifted)
}

// Digits start from the first bit, "a" character starts from the 11th one etc.
func calcBitNum(char rune) (uint64, bool) {
	// Characters a-z use bit positions 10-35
	if char > 96 && char < 123 {
		return uint64(char - 87), true
	}

	// Characters A-Z use bit positions 36-61
	if char > 64 && char < 91 {
		return uint64(char - 29), true
	}

	// digits 0-9 use bit positions 0-9
	if char > 47 && char < 58 {
		return uint64(char - 48), true
	}

	// space uses 62 bit position
	if char == 32 {
		return 62, true
	}

	return 0, false
}

func calcRune(bitNum int32) rune {
	if bitNum > 9 && bitNum < 36 {
		return bitNum + 87
	}

	if bitNum > 35 && bitNum < 62 {
		return bitNum + 29
	}

	if bitNum > 0 && bitNum < 10 {
		return bitNum + 48
	}

	return 32
}

func (trie *Trie) checkKey(key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	for _, char := range key {
		if _, ok := calcBitNum(char); !ok {
			return errors.New("key can contain only a-z, A-Z characters, digits and space")
		}
	}

	return nil
}
