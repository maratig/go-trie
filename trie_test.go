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
	for i := 0; i < 100000; i++ {
		for j := 0; j < 8; j++ {
			word[j] = rand.Int31n(25) + 97
		}
		tr.Set(string(word), true)
	}
	tr.Set("mother", true)
	tr.Set("father", true)

	hasMother, _ := tr.Get("mother")
	hasFather, _ := tr.Get("father")
	hasSister, _ := tr.Get("sister")

	if hasMother == nil {
		t.Fatalf(`error while searching for "mother" word`)
	}

	if hasFather == nil {
		t.Fatalf(`error while searching for "father" word`)
	}

	if hasSister == nil {
		t.Fatalf(`error while searching for "sister" word`)
	}
}
