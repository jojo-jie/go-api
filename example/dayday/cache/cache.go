package cache

import "dayday/blog_demo"

type cacheable interface {
	blog_demo.Category | blog_demo.Post
}

type cache[T cacheable] struct {
	data map[string]T
}

func (c *cache[T]) Set(k string, v T) {
	c.data[k] = v
}

func (c *cache[T]) Get(k string) (v T) {
	if v, ok := c.data[k]; ok {
		return v
	}
	return
}

func New[T cacheable]() *cache[T] {
	c := cache[T]{}
	c.data = make(map[string]T)
	return &c
}
