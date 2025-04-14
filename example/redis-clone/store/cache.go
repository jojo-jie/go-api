package store

import (
	"container/list"
	"sync"
	"time"
)

type Item struct {
	value     string
	expiresAt time.Time
	elem      *list.Element
}

type Cache struct {
	mu      sync.RWMutex     // Mutex to ensure thread safety for concurrent access to the cache
	items   map[string]*Item // Map that stores cache items by their key
	lru     *list.List       // Doubly linked list used for tracking the access order of cache items (for LRU eviction)
	maxSize int              // Maximum number of items the cache can hold before eviction occurs
	persist *Persistence     // Optional persistence mechanism for appending commands to a file
}

func NewCache(maxSize int, persist *Persistence) *Cache {
	return &Cache{
		maxSize: maxSize,
		persist: persist,
		items:   map[string]*Item{},
		lru:     list.New(),
	}
}
