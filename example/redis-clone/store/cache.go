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
