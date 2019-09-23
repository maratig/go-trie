package go_trie

import (
	"math/bits"
)

type Trie struct {
	lettersIndex      uint32
	lettersTerminated uint32
	subTries          []*Trie
}

// For letter "a" bitNum is 0 (ASCII code for "a" is 97), for "b" bitNum is 1 etc.
func calcBitNum(letter rune) uint32 {
	// code 32 is for space
	if letter == 32 {
		return uint32(31)
	}

	return uint32(letter - 97)
}

func (trie *Trie) Add(words []rune) {
	if len(words) == 0 {
		return
	}

	bitNum := calcBitNum(words[0])

	// If there are no more letters then set "terminated" bit
	if len(words) == 1 {
		trie.lettersTerminated = trie.lettersTerminated | (1 << bitNum)
		return
	}

	var subTrieIndex int
	// Letter "a" uses the first bit, "b" uses the second etc.
	// If letter's bit is not set then set it and move some subtrees to the right
	if trie.lettersIndex&(1<<bitNum) == 0 {
		trie.lettersIndex = trie.lettersIndex | (1 << bitNum)
		subTrieIndex = trie.getSubTreeIndex(bitNum)

		if len(trie.subTries) == 16 {
			newSubTries := make([]*Trie, 16, 26)
			copy(newSubTries, trie.subTries)
			trie.subTries = newSubTries
		}

		trie.subTries = append(trie.subTries, nil)
		copy(trie.subTries[subTrieIndex+1:], trie.subTries[subTrieIndex:])
		trie.subTries[subTrieIndex] = &Trie{}
	} else {
		// Letter's bit was set so just get the subTree index
		subTrieIndex = trie.getSubTreeIndex(bitNum)
	}

	trie.subTries[subTrieIndex].Add(words[1:])
}

// subTreeIndex is an amount of bits before bitNum
func (trie *Trie) getSubTreeIndex(bitNum uint32) int {
	shifted := uint(trie.lettersIndex << (32 - bitNum))
	return bits.OnesCount(shifted)
}

func (trie *Trie) Has(words []rune) bool {
	if len(words) == 0 {
		return false
	}

	var letter rune
	tr, wordLastIdx := trie, len(words) - 1
	for i := 0; i < wordLastIdx; i++ {
		letter = words[i]
		if !tr.hasLetter(letter) {
			return false
		}

		subTreeIndex := tr.getSubTreeIndex(calcBitNum(letter))
		tr = tr.subTries[subTreeIndex]
	}

	return tr.canTerminate(words[wordLastIdx])
}

func (trie *Trie) hasLetter(letter rune) bool {
	bitNum := calcBitNum(letter)

	return trie.lettersIndex & (1 << bitNum) != 0
}

func (trie *Trie) canTerminate(letter rune) bool {
	bitNum := calcBitNum(letter)

	return trie.lettersTerminated & (1 << bitNum) != 0
}