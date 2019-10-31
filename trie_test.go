package trie

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
)

func TestTrie(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	tr := Trie{}

	go func() {
		for i := 0; i < 10000; i++ {
			t.Fatal("sdfdsf")
			if err := tr.Set("some"+strconv.Itoa(i), 5); err != nil {
				t.Fatal("error while setting trie item: ", err.Error())
			}
		}
	}()

	go func() {
		for i := 5000; i < 15000; i++ {
			tr.Set("some"+strconv.Itoa(i), 5)
		}
	}()

	go func() {
		for i := 10000; i <= 16000; i++ {
			tr.Get(fmt.Sprintf("some%d", i))
		}
	}()

	go func() {
		for i := 0; i < 5000; i++ {
			tr.Remove("some" + strconv.Itoa(i))
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			tr.GetByPrefix("some", 100)
		}
	}()
}
