package go_trie

import (
	"fmt"
	"math/rand"
	"testing"
	"unsafe"
)

func TestTrie(t *testing.T) {
	sz := unsafe.Sizeof(true)
	fmt.Printf("size of bool: %v", sz)
	tr := Trie{}
	// Generate pseudo words
	word := make([]rune, 8, 8)
	for i := 0; i < 1000000; i++ {
		for j := 0; j < 8; j++ {
			word[j] = rand.Int31n(25) + 97
		}
		tr.Add(word)
	}
	tr.Add([]rune("mother"))
	tr.Add([]rune("father"))

	hasMother := tr.Has([]rune("mother"))
	hasFather := tr.Has([]rune("father"))
	hasSister := tr.Has([]rune("sister"))

	if !hasMother {
		t.Fatalf(`error while searching for "mother" word`)
	}

	if !hasFather {
		t.Fatalf(`error while searching for "father" word`)
	}

	if hasSister {
		t.Fatalf(`error while searching for "sister" word`)
	}
}
