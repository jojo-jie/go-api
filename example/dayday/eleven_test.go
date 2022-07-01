package dayday

import (
	"github.com/bluele/gcache"
	"testing"
)

func TestCache(t *testing.T) {
	gc := gcache.New(20).
		LRU().
		Build()
	gc.Set("key", "ok")
	value, err := gc.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Get:", value)
}