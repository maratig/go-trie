package go_trie

import (
	"math/rand"
	"testing"
)

func TestTrie(t *testing.T) {
	tr := Tree{}
	// Generate pseudo words
	word := make([]rune, 8, 8)
	for i := 0; i < 1000000; i++ {
		for j := 0; j < 8; j++ {
			word[j] = rand.Int31n(25) + 97
		}
		tr.AddWord(word)
	}
	tr.AddWord([]rune("mother"))
	tr.AddWord([]rune("father"))

	hasMother := tr.HasWord([]rune("mother"))
	hasFather := tr.HasWord([]rune("father"))
	hasSister := tr.HasWord([]rune("sister"))

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
