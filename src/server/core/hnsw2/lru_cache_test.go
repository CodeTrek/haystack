package hnsw2

import (
	"fmt"
	"testing"
)

func TestLRUCache(t *testing.T) {
	// Test Put and Get
	t.Run("Put and Get", func(t *testing.T) {
		cache := NewLRUCache(3)
		// Add three nodes
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("2"), nodeData{value: []float32{2.0}})
		cache.Put([]byte("3"), nodeData{value: []float32{3.0}})

		// Test getting existing node
		if data, ok := cache.Get([]byte("1")); !ok || data.value[0] != 1.0 {
			t.Errorf("Get(1) = %v, %v; want Value=[1.0], true", data, ok)
		}

		// Test getting non-existent node
		if _, ok := cache.Get([]byte("4")); ok {
			t.Error("Get(4) should return false")
		}
	})

	// Test LRU eviction policy
	t.Run("LRU Eviction", func(t *testing.T) {
		cache := NewLRUCache(3)
		// Add three nodes
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("2"), nodeData{value: []float32{2.0}})
		cache.Put([]byte("3"), nodeData{value: []float32{3.0}})

		// Add fourth node, should evict first node
		cache.Put([]byte("4"), nodeData{value: []float32{4.0}})

		// First node should be evicted
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Node 1 should have been evicted")
		}

		// Other nodes should still be present
		if _, ok := cache.Get([]byte("2")); !ok {
			t.Error("Node 2 should still be in cache")
		}
		if _, ok := cache.Get([]byte("3")); !ok {
			t.Error("Node 3 should still be in cache")
		}
		if _, ok := cache.Get([]byte("4")); !ok {
			t.Error("Node 4 should still be in cache")
		}
	})

	// Test access order impact
	t.Run("Access Order", func(t *testing.T) {
		cache := NewLRUCache(3)

		// Add three nodes
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("2"), nodeData{value: []float32{2.0}})
		cache.Put([]byte("3"), nodeData{value: []float32{3.0}})

		// Access nodes 1 and 3, making node 2 the least recently used
		if _, ok := cache.Get([]byte("1")); !ok {
			t.Error("Failed to get node 1")
		}
		if _, ok := cache.Get([]byte("3")); !ok {
			t.Error("Failed to get node 3")
		}

		// Add new node, should evict node 2 (least recently used)
		cache.Put([]byte("4"), nodeData{value: []float32{4.0}})

		// Node 2 should be evicted
		if _, ok := cache.Get([]byte("2")); ok {
			t.Error("Node 2 should have been evicted")
		}

		// Nodes 1 and 3 should still be present
		if _, ok := cache.Get([]byte("1")); !ok {
			t.Error("Node 1 should still be in cache")
		}
		if _, ok := cache.Get([]byte("3")); !ok {
			t.Error("Node 3 should still be in cache")
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Delete([]byte("1"))
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Node 1 should have been deleted")
		}
	})

	// Test Clear
	t.Run("Clear", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Clear()
		if cache.list.Len() != 0 {
			t.Errorf("Cache length after Clear = %d; want 0", cache.list.Len())
		}
		if len(cache.cache) != 0 {
			t.Errorf("Cache map size after Clear = %d; want 0", len(cache.cache))
		}
	})

	// Test concurrent access
	t.Run("Concurrent Access", func(t *testing.T) {
		cache := NewLRUCache(100)
		done := make(chan bool)

		// Start multiple goroutines to access cache concurrently
		for i := 0; i < 10; i++ {
			go func(id int) {
				for j := 0; j < 100; j++ {
					key := int64(id*100 + j)
					cache.Put([]byte(fmt.Sprintf("%d", key)), nodeData{value: []float32{float32(key)}})
					cache.Get([]byte(fmt.Sprintf("%d", key)))
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify cache state
		if cache.list.Len() > cache.capacity {
			t.Errorf("Cache size = %d; want <= %d", cache.list.Len(), cache.capacity)
		}
	})
}

func TestLRUCacheEdgeCases(t *testing.T) {
	// Test zero capacity cache
	t.Run("Zero Capacity", func(t *testing.T) {
		cache := NewLRUCache(0)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Cache with zero capacity should not store any items")
		}
	})

	// Test negative capacity cache
	t.Run("Negative Capacity", func(t *testing.T) {
		cache := NewLRUCache(-1)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Cache with negative capacity should not store any items")
		}
	})

	// Test duplicate Put
	t.Run("Duplicate Put", func(t *testing.T) {
		cache := NewLRUCache(2)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("1"), nodeData{value: []float32{2.0}})

		if data, ok := cache.Get([]byte("1")); !ok || data.value[0] != 2.0 {
			t.Errorf("Get(1) after duplicate Put = %v, %v; want Value=[2.0], true", data, ok)
		}
	})
}
