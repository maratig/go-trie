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
	var subTrieId int
	for _, char := range key {
		bitNum, ok := calcBitNum(char)

		if !ok {
			return errors.New("given string contains not only a-z, A-Z, 0-9 and/or space")
		}

		subTrieId = tr.getSubTreeIndex(bitNum)
		// If trie does not contain given character
		if tr.characters&(1<<bitNum) == 0 {
			tr.characters = tr.characters | (1 << bitNum)
			tr.subTries = append(tr.subTries, nil)
			copy(tr.subTries[subTrieId+1:], tr.subTries[subTrieId:])
			tr.subTries[subTrieId] = &Trie{}
		}

		tr = tr.subTries[subTrieId]
	}
	tr.data = value

	return nil
}

func (trie *Trie) Get(key string) interface{} {
	tr := trie
	for _, char := range key {
		bitNum, ok := calcBitNum(char)

		if !ok {
			return nil
		}

		subTrieId := tr.getSubTreeIndex(bitNum)
		if tr.characters & (1<<bitNum) == 0 {
			return nil
		}

		tr = tr.subTries[subTrieId]
	}

	return tr.data
}

func (trie *Trie) Delete(key string) {
	
}

// subTreeIndex is an amount of bits before bitNum
func (trie *Trie) getSubTreeIndex(bitNum uint64) int {
	shifted := uint(trie.characters << (64 - bitNum))
	return bits.OnesCount(shifted)
}

// Digits start from the first bit, "a" character starts from the 11th one etc.
func calcBitNum(char rune) (uint64, bool) {
	// a-z characters use 10-35 bit positions
	if char > 96 && char < 123 {
		return uint64(char - 87), true
	}

	// A-Z characters use 36-61 bit positions
	if char > 64 && char < 91 {
		return uint64(char - 29), true
	}

	// 0-9 digits use 0-9 bit positions
	if char > 47 && char < 58 {
		return uint64(char - 48), true
	}

	// space uses 62 bit position
	if char == 32 {
		return 62, true
	}

	return 0, false
}
