package hnsw

import (
	"container/list"
	"sync"
)

// LRUCache implements a thread-safe LRU cache for nodeData
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

// cacheEntry represents an entry in the LRU cache
type cacheEntry struct {
	key   []byte
	value nodeData
}

// NewLRUCache creates a new LRU cache with the specified capacity
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 0
	}
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key []byte) (nodeData, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return nodeData{}, false
	}

	keyStr := string(key)
	if elem, ok := c.cache[keyStr]; ok {
		c.list.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nodeData{}, false
}

// Put adds or updates a value in the cache
func (c *LRUCache) Put(key []byte, value nodeData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity == 0 {
		return
	}

	keyStr := string(key)
	if elem, ok := c.cache[keyStr]; ok {
		c.list.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}

	if c.list.Len() >= c.capacity {
		// Remove the least recently used item
		last := c.list.Back()
		if last != nil {
			delete(c.cache, string(last.Value.(*cacheEntry).key))
			c.list.Remove(last)
		}
	}

	entry := &cacheEntry{key: key, value: value}
	elem := c.list.PushFront(entry)
	c.cache[keyStr] = elem
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(key []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	keyStr := string(key)
	if elem, ok := c.cache[keyStr]; ok {
		delete(c.cache, keyStr)
		c.list.Remove(elem)
	}
}

// Clear removes all values from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.list.Init()
}
