package trie

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestTrie(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var wg sync.WaitGroup
	wg.Add(5)
	tr := Trie{}

	go func() {
		for i := 0; i < 300000; i++ {
			if err := tr.Set("some"+strings.ToUpper(strconv.Itoa(i)), 5); err != nil {
				t.Fatal("error while setting bulk of items: ", err.Error())
			}
		}
		wg.Done()
	}()

	go func() {
		for i := 150000; i < 450000; i++ {
			if err := tr.Set("some"+strings.ToUpper(strconv.Itoa(i)), 5); err != nil {
				t.Fatal("error while setting another bulk of items: ", err.Error())
			}
		}
		wg.Done()
	}()

	go func() {
		for i := 300000; i <= 480000; i++ {
			if _, err := tr.Get(strings.ToUpper(fmt.Sprintf("some%d", i))); err != nil {
				t.Fatalf("error while getting some%d trie item: %s", i, err.Error())
			}
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 150000; i++ {
			key := strings.ToUpper("some" + strconv.Itoa(i))
			if err := tr.Remove(key); err != nil {
				t.Fatalf("error while removing %s trie item: %s", key, err.Error())
			}
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < 3000; i++ {
			if _, err := tr.GetByPrefix("some", 10); err != nil {
				t.Fatalf(`error while getting items by prefix "some": %s`, err.Error())
			}
		}
		wg.Done()
	}()

	wg.Wait()

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
		t.Fatal("result must not be nil")
	}

	if err := tr.Remove("my first example"); err != nil {
		t.Fatalf(`error while removing an item: %s`, err.Error())
	}

	result, err = tr.Get("my first example")

	if err != nil {
		t.Fatal("error while getting an item: ", err.Error())
	}

	if result != nil {
		t.Fatal("item was removed earlier but still exists in the trie")
	}

	items, err := tr.GetByPrefix("my", 10)
	if err != nil {
		t.Fatal("error while getting items by prefix: ", err.Error())
	}

	if len(items) != 2 {
		t.Fatal(`the amount of items with prefix "my" must be 2 but not `, len(items))
	}
}
