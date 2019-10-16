package go_trie

import (
	"errors"
	"math/bits"
)

// Trie works with latin characters, space and digits
type Trie struct {
	characters uint64
	subTries   []*Trie
	data       interface{}
}

func (trie *Trie) Set(key string, value interface{}) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	if value == nil {
		return errors.New("value cannot be nil")
	}

	tr := trie
	for _, char := range key {
		bitNum, subTrieId, subTrie, err := tr.getSubTrie(char)

		if err != nil {
			return err
		}

		if subTrie == nil {
			subTrie = &Trie{}
			tr.characters = tr.characters | (1 << bitNum)
			tr.subTries = append(tr.subTries, nil)
			copy(tr.subTries[subTrieId+1:], tr.subTries[subTrieId:])
			tr.subTries[subTrieId] = subTrie
		}

		tr = subTrie
	}
	tr.data = value

	return nil
}

// Find an item having exact given key
func (trie *Trie) Get(key string) (interface{}, error) {
	tr := trie
	for _, char := range key {
		_, _, subTrie, err := tr.getSubTrie(char)

		if err != nil {
			return nil, err
		}

		if subTrie == nil {
			return nil, nil
		}
		tr = subTrie
	}

	return tr.data, nil
}

type TraverseNode struct {
	key string
	trie *Trie
}

func (tn *TraverseNode) getChildren() []*TraverseNode {
	trie := tn.trie
	ret, i := make([]*TraverseNode, len(trie.subTries)), 0
	for bitNum := int32(0); bitNum < 64; bitNum++ {
		if trie.characters & (1<<bitNum) == 0 {
			continue
		}

		char := calcRune(bitNum)
		ret[i] = &TraverseNode{key: tn.key + string(char), trie: trie.subTries[i]}
		i++

		if i == len(trie.subTries) {
			break
		}
	}

	return ret
}

type DataForKey struct {
	Key string
	Data interface{}
}

// Find items having given prefix
func (trie *Trie) GetByPrefix(prefix string, limit int) ([]DataForKey, error) {
	if prefix == "" {
		return nil, nil
	}

	if limit <= 0 {
		limit = -1
	}

	tr := trie
	var ret []DataForKey
	for _, char := range prefix {
		_, _, subTrie, err := tr.getSubTrie(char)

		if err != nil {
			return nil, err
		}

		if subTrie == nil {
			return nil, nil
		}

		tr = subTrie
	}

	if tr.data != nil {
		ret = append(ret, DataForKey{Key: prefix, Data: tr.data})

		if limit > 0 {
			limit--
		}
	}

	var toTraverse []*TraverseNode
	toTraverse = append(toTraverse, &TraverseNode{key: prefix, trie: tr})
	for limit > 0 && len(toTraverse) > 0 {
		toTr := toTraverse[0]
		items, added := toTr.getChildren(), 0
		for _, item := range items {
			if item.trie.data != nil {
				ret = append(ret, DataForKey{Key: item.key, Data: item.trie.data})
				added++
			}
		}
		limit -= added

		toTraverse = append(toTraverse, items...)
		toTraverse = toTraverse[1:]
	}

	return ret, nil
}

func (trie *Trie) Remove(key string) error {
	if key == "" {
		return nil
	}

	type ToRemove struct {
		root *Trie
		bitNum uint64
		subTrueId int
	}

	var removeList []ToRemove
	tr := trie
	for _, char := range key {
		bitNum, subTrieId, subTrie, err := tr.getSubTrie(char)

		if err != nil {
			return err
		}

		if bits.OnesCount64(subTrie.characters) <= 1 {
			removeList = append(removeList, ToRemove{root: tr, bitNum: bitNum, subTrueId: subTrieId})
		} else {
			removeList = removeList[:0]
		}
		tr = subTrie
	}

	tr.data = nil

	if bits.OnesCount64(tr.characters) > 0 {
		return nil
	}

	for i := len(removeList)-1; i >= 0; i-- {
		toRem := removeList[i]
		toRem.root.characters = toRem.root.characters & (0<<toRem.bitNum)
		subCount := len(toRem.root.subTries)

		if subCount == 1 {
			toRem.root.subTries = toRem.root.subTries[:0]
			continue
		}

		copy(toRem.root.subTries[:toRem.subTrueId], toRem.root.subTries[:toRem.subTrueId+1])
		toRem.root.subTries = toRem.root.subTries[:subCount-1]
	}

	return nil
}

func (trie *Trie) getSubTrie(char rune) (uint64, int, *Trie, error) {
	bitNum, ok := calcBitNum(char)

	if !ok {
		return 0, 0, nil, errors.New("trie key can contain only a-z, A-Z, 0-9 characters and space")
	}

	// There is no subTrie under given character since the corresponding bit is zero
	if trie.characters & (1<<bitNum) == 0 {
		return 0, 0, nil, nil
	}

	subTrieId := trie.getSubTreeIndex(bitNum)
	return bitNum, subTrieId, trie.subTries[subTrieId], nil
}

// subTreeIndex is an amount of bits before bitNum
func (trie *Trie) getSubTreeIndex(bitNum uint64) int {
	shifted := uint64(trie.characters << (64 - bitNum))
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