package store

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"time"
)

type Item struct {
	value     string
	expiresAt time.Time
	elem      *list.Element
}

type Cache struct {
	mu      sync.RWMutex
	items   map[string]*Item
	lru     *list.List
	maxSize int
	persist *Persistence
}

func NewCache(maxSize int, persist *Persistence) *Cache {
	return &Cache{
		maxSize: maxSize,
		persist: persist,
		items:   map[string]*Item{},
		lru:     list.New(),
	}
}

func (c *Cache) Set(key string, value string, ttl time.Duration, replaying bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var expiresAt time.Time
	if ttl != 0 {
		expiresAt = time.Now().Add(ttl)
	}

	if item, exists := c.items[key]; exists {
		item.value = value
		item.expiresAt = expiresAt
		c.lru.MoveToFront(item.elem)
	} else {
		elem := c.lru.PushFront(key)
		c.items[key] = &Item{
			value:     value,
			expiresAt: expiresAt,
			elem:      elem,
		}
	}

	if c.lru.Len() > c.maxSize {
		oldest := c.lru.Back()
		if oldest != nil {
			delete(c.items, oldest.Value.(string))
			c.lru.Remove(oldest)
		}
	}

	if c.persist != nil && !replaying {
		cmd := fmt.Sprintf("SET %s %s", key, value)
		err := c.persist.Append(cmd)
		if err != nil {
			log.Println("Failed to append to AOF file:", err)
		}
	}
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[key]
	if !exists {
		return "", exists
	}
	if !item.expiresAt.IsZero() && item.expiresAt.Before(time.Now()) {
		c.lru.Remove(item.elem)
		delete(c.items, key)
		return "", exists
	}
	c.lru.MoveToFront(item.elem)
	return item.value, true
}

func (c *Cache) StartCleaningServer() {
	ticker := time.NewTicker(time.Minute * 1)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Running cache cleanup")
		c.cleanExpiredItems()
	}
}

func (c *Cache) cleanExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for e := c.lru.Back(); e != nil; e = e.Prev() {
		item := c.items[e.Value.(string)]
		if item.expiresAt.IsZero() {
			continue
		}

		if time.Now().After(item.expiresAt) {
			c.lru.Remove(e)
			delete(c.items, e.Value.(string))
			log.Printf("Removed expired item: %s\n", e.Value)
		}
	}

	// If the cache exceeds the maximum size, evict the least recently used item
	if c.lru.Len() > c.maxSize {
		oldest := c.lru.Back()
		if oldest != nil {
			delete(c.items, oldest.Value.(string))
			c.lru.Remove(oldest)
		}
	}
}
