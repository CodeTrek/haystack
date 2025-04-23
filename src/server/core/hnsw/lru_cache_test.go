package hnsw

import (
	"fmt"
	"testing"
)

func TestLRUCache(t *testing.T) {
	// 测试 Put 和 Get
	t.Run("Put and Get", func(t *testing.T) {
		cache := NewLRUCache(3)
		// 添加三个节点
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("2"), nodeData{value: []float32{2.0}})
		cache.Put([]byte("3"), nodeData{value: []float32{3.0}})

		// 测试获取存在的节点
		if data, ok := cache.Get([]byte("1")); !ok || data.value[0] != 1.0 {
			t.Errorf("Get(1) = %v, %v; want Value=[1.0], true", data, ok)
		}

		// 测试获取不存在的节点
		if _, ok := cache.Get([]byte("4")); ok {
			t.Error("Get(4) should return false")
		}
	})

	// 测试 LRU 淘汰策略
	t.Run("LRU Eviction", func(t *testing.T) {
		cache := NewLRUCache(3)
		// 添加三个节点
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("2"), nodeData{value: []float32{2.0}})
		cache.Put([]byte("3"), nodeData{value: []float32{3.0}})

		// 添加第四个节点，应该淘汰第一个节点
		cache.Put([]byte("4"), nodeData{value: []float32{4.0}})

		// 第一个节点应该被淘汰
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Node 1 should have been evicted")
		}

		// 其他节点应该还在
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

	// 测试访问顺序影响
	t.Run("Access Order", func(t *testing.T) {
		cache := NewLRUCache(3)

		// 添加三个节点
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("2"), nodeData{value: []float32{2.0}})
		cache.Put([]byte("3"), nodeData{value: []float32{3.0}})

		// 访问节点1和节点3，使节点2成为最久未访问的
		if _, ok := cache.Get([]byte("1")); !ok {
			t.Error("Failed to get node 1")
		}
		if _, ok := cache.Get([]byte("3")); !ok {
			t.Error("Failed to get node 3")
		}

		// 添加新节点，应该淘汰节点2（最久未访问）
		cache.Put([]byte("4"), nodeData{value: []float32{4.0}})

		// 节点2应该被淘汰
		if _, ok := cache.Get([]byte("2")); ok {
			t.Error("Node 2 should have been evicted")
		}

		// 节点1和3应该还在
		if _, ok := cache.Get([]byte("1")); !ok {
			t.Error("Node 1 should still be in cache")
		}
		if _, ok := cache.Get([]byte("3")); !ok {
			t.Error("Node 3 should still be in cache")
		}
	})

	// 测试 Delete
	t.Run("Delete", func(t *testing.T) {
		cache := NewLRUCache(3)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Delete([]byte("1"))
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Node 1 should have been deleted")
		}
	})

	// 测试 Clear
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

	// 测试并发访问
	t.Run("Concurrent Access", func(t *testing.T) {
		cache := NewLRUCache(100)
		done := make(chan bool)

		// 启动多个goroutine同时访问缓存
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

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证缓存状态
		if cache.list.Len() > cache.capacity {
			t.Errorf("Cache size = %d; want <= %d", cache.list.Len(), cache.capacity)
		}
	})
}

func TestLRUCacheEdgeCases(t *testing.T) {
	// 测试零容量缓存
	t.Run("Zero Capacity", func(t *testing.T) {
		cache := NewLRUCache(0)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Cache with zero capacity should not store any items")
		}
	})

	// 测试负容量缓存
	t.Run("Negative Capacity", func(t *testing.T) {
		cache := NewLRUCache(-1)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		if _, ok := cache.Get([]byte("1")); ok {
			t.Error("Cache with negative capacity should not store any items")
		}
	})

	// 测试重复Put
	t.Run("Duplicate Put", func(t *testing.T) {
		cache := NewLRUCache(2)
		cache.Put([]byte("1"), nodeData{value: []float32{1.0}})
		cache.Put([]byte("1"), nodeData{value: []float32{2.0}})

		if data, ok := cache.Get([]byte("1")); !ok || data.value[0] != 2.0 {
			t.Errorf("Get(1) after duplicate Put = %v, %v; want Value=[2.0], true", data, ok)
		}
	})
}
