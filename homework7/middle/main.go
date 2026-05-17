package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	mutex  sync.RWMutex
	values map[string]int
}

func NewCache() *Cache {
	return &Cache{
		values: make(map[string]int),
	}
}

func (c *Cache) Get(key string) int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, ok := c.values[key]
	if !ok {
		return -1
	}
	return value
}

func (c *Cache) Set(key string, value int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.values[key] = value
}

func main() {
	wg := sync.WaitGroup{}

	cache := NewCache()

	for i := range 10 {
		wg.Go(func() {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(key, i)
		})
	}

	for i := range 10 {
		wg.Go(func() {
			key := fmt.Sprintf("key_%d", i)
			fmt.Printf("%#v\n", cache.Get(key))
		})
	}

	wg.Wait()

	fmt.Println(cache.values)
}
