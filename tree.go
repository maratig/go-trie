package go_trie

import (
	"math/bits"
)

type Tree struct {
	lettersIndex      uint32
	lettersTerminated uint32
	subTrees          []*Tree
}

// For letter "a" bitNum is 0 (ASCII code for "a" is 97), for "b" bitNum is 1 etc.
func calcBitNum(letter rune) uint32 {
	return uint32(letter - 97)
}

func (tree *Tree) AddWord(word []rune) {
	if len(word) == 0 {
		return
	}

	bitNum := calcBitNum(word[0])

	// If there are no more letters then set "terminated" bit
	if len(word) == 1 {
		tree.lettersTerminated = tree.lettersTerminated | (1 << bitNum)
		return
	}

	var subTreeIndex int
	// Letter "a" uses the first bit, "b" uses the second etc.
	// If letter's bit is not set then set it and move some subtrees to the right
	if tree.lettersIndex&(1<<bitNum) == 0 {
		tree.lettersIndex = tree.lettersIndex | (1 << bitNum)
		subTreeIndex = tree.getSubTreeIndex(bitNum)

		if len(tree.subTrees) == 16 {
			newSubTrees := make([]*Tree, 16, 26)
			copy(newSubTrees, tree.subTrees)
			tree.subTrees = newSubTrees
		}

		tree.subTrees = append(tree.subTrees, nil)
		copy(tree.subTrees[subTreeIndex+1:], tree.subTrees[subTreeIndex:])
		tree.subTrees[subTreeIndex] = &Tree{}
	} else {
		// Letter's bit was set so just get the subTree index
		subTreeIndex = tree.getSubTreeIndex(bitNum)
	}

	tree.subTrees[subTreeIndex].AddWord(word[1:])
}

// subTreeIndex is an amount of bits before bitNum
func (tree *Tree) getSubTreeIndex(bitNum uint32) int {
	shifted := uint(tree.lettersIndex << (32 - bitNum))
	return bits.OnesCount(shifted)
}

func (tree *Tree) HasWord(word []rune) bool {
	if len(word) == 0 {
		return false
	}

	var letter rune
	tr, wordLastIdx := tree, len(word) - 1
	for i := 0; i < wordLastIdx; i++ {
		letter = word[i]
		if !tr.hasLetter(letter) {
			return false
		}

		subTreeIndex := tr.getSubTreeIndex(calcBitNum(letter))
		tr = tr.subTrees[subTreeIndex]
	}

	return tr.canTerminate(word[wordLastIdx])
}

func (tree *Tree) hasLetter(letter rune) bool {
	bitNum := calcBitNum(letter)

	return tree.lettersIndex & (1 << bitNum) != 0
}

func (tree *Tree) canTerminate(letter rune) bool {
	bitNum := calcBitNum(letter)

	return tree.lettersTerminated & (1 << bitNum) != 0
}