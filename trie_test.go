package trie

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
)

func TestTrie(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	tr, done := Trie{}, make(chan bool)

	go func() {
		for i := 0; i < 100000; i++ {
			if err := tr.Set("some"+strconv.Itoa(i), 5); err != nil {
				t.Fatal("error while setting bulk of items: ", err.Error())
			}
		}
		done <- true
	}()

	go func() {
		for i := 50000; i < 150000; i++ {
			if err := tr.Set("some"+strconv.Itoa(i), 5); err != nil {
				t.Fatal("error while setting another bulk of items: ", err.Error())
			}
		}
		done <- true
	}()

	go func() {
		for i := 100000; i <= 160000; i++ {
			if _, err := tr.Get(fmt.Sprintf("some%d", i)); err != nil {
				t.Fatalf("error while getting some%d trie item: %s", i, err.Error())
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50000; i++ {
			key := "some" + strconv.Itoa(i)
			if err := tr.Remove(key); err != nil {
				t.Fatalf("error while removing %s trie item: %s", key, err.Error())
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10000; i++ {
			if _, err := tr.GetByPrefix("some", 100); err != nil {
				t.Fatalf(`error while getting items by prefix "some": %s`, err.Error())
			}
		}
		done <- true
	}()

	left := 5
	for {
		if left == 0 {
			break
		}

		if <-done {
			left--
		}
	}

	if err := tr.Set("my first example", true); err != nil {
		t.Fatalf(`error while setting an item: %s`, err.Error())
	}

	if err := tr.Set("my second example", true); err != nil {
		t.Fatalf(`error while setting an item: %s`, err.Error())
	}

	if err := tr.Set("my third example", true); err != nil {
		t.Fatalf(`error while setting an item: %s`, err.Error())
	}

	result, err := tr.Get("my first example")

	if err != nil {
		t.Fatalf(`error while getting an item: %s`, err.Error())
	}

	if result == nil {
		t.Fatal("result must not be nill")
	}

	if err := tr.Remove("my first example"); err != nil {
		t.Fatalf(`error while removing an item: %s`, err.Error())
	}

	result, err = tr.Get("my first example")

	if err != nil {
		t.Fatal("error while getting an item: ", err.Error())
	}

	if result != nil {
		t.Fatal("item was removed earlier by still exists in the trie")
	}

	items, err := tr.GetByPrefix("my", 10);
	if err != nil {
		t.Fatal("error while getting items by prefix: ", err.Error())
	}

	if len(items) != 2 {
		t.Fatal(`the amount of items with prefix "my" must be 2 instead of `, len(items))
	}
}
