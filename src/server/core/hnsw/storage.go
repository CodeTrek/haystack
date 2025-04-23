package hnsw

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
)

// Key type constants
const (
	KeyTypeNode  byte = 'n' // Node data
	KeyTypeEntry byte = 'e' // Entry point
)

// Storage handles all database operations for the HNSW graph
type Storage struct {
	db    *pebble.DB
	path  string
	mu    sync.RWMutex
	cache *LRUCache
}

var ErrNotFound = pebble.ErrNotFound

// NewStorage creates a new storage instance
func NewStorage(path string) (*Storage, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Configure Pebble options optimized for HNSW
	opts := &pebble.Options{
		// Cache configuration
		Cache: pebble.NewCache(64 << 20), // 64MB cache

		// WAL settings for durability and performance
		WALMinSyncInterval: func() time.Duration {
			return 500 * time.Microsecond
		},

		// File handling
		MaxOpenFiles: 8192,

		// MemTable settings
		MemTableSize:                4 * 1024 * 1024,
		MemTableStopWritesThreshold: 2,

		// L0 compaction settings optimized for read-heavy workload
		L0CompactionFileThreshold: 256, // Lowered from 1024 for faster compaction
		L0CompactionThreshold:     8,
		L0StopWritesThreshold:     12,

		// Level settings with bloom filter
		Levels: []pebble.LevelOptions{
			{
				BlockSize:    32 * 1024,
				FilterPolicy: bloom.FilterPolicy(10),
			},
		},
	}
	opts.EnsureDefaults()

	db, err := pebble.Open(filepath.Join(path, "hnsw"), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open pebble db: %w", err)
	}

	return &Storage{
		db:    db,
		path:  path,
		cache: NewLRUCache(10000), // 设置缓存容量为10000个节点
	}, nil
}

// int64ToBytes converts an int64 to a byte slice
func int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

// bytesToInt64 converts a byte slice to an int64
func bytesToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

// makeKey creates a key with workspace prefix
func makeKey(workspace int64, keyType byte, id int64) []byte {
	// Format: <keytype><workspace><key>
	buf := make([]byte, 0, 17) // 1 byte for keyType + 8 bytes for workspace + 8 bytes for id
	buf = append(buf, keyType)
	buf = append(buf, int64ToBytes(workspace)...)
	buf = append(buf, int64ToBytes(id)...)
	return buf
}

// SaveNode saves a node to the database
func (s *Storage) SaveNode(workspace int64, id int64, data nodeData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("failed to encode node data: %w", err)
	}

	key := makeKey(workspace, KeyTypeNode, id)
	if err := s.db.Set(key, buf.Bytes(), pebble.Sync); err != nil {
		return fmt.Errorf("failed to save node: %w", err)
	}

	// 更新缓存
	s.cache.Put(key, data)

	return nil
}

// LoadNode loads a node from the database
func (s *Storage) LoadNode(workspace int64, id int64) (nodeData, error) {
	// 首先尝试从缓存中获取
	key := makeKey(workspace, KeyTypeNode, id)
	if data, ok := s.cache.Get(key); ok {
		return data, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	value, closer, err := s.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nodeData{}, ErrNotFound
		}
		return nodeData{}, err
	}
	defer closer.Close()

	var data nodeData
	dec := gob.NewDecoder(bytes.NewReader(value))
	if err := dec.Decode(&data); err != nil {
		return nodeData{}, fmt.Errorf("failed to decode node data: %w", err)
	}

	// 将数据存入缓存
	s.cache.Put(key, data)

	return data, nil
}

// DeleteNode deletes a node from the database
func (s *Storage) DeleteNode(workspace int64, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := makeKey(workspace, KeyTypeNode, id)
	if err := s.db.Delete(key, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	// 从缓存中删除
	s.cache.Delete(key)

	return nil
}

// GetEntryPoint retrieves the entry point from storage
func (s *Storage) GetEntryPoint(workspace int64) (*int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := makeKey(workspace, KeyTypeEntry, 0)
	value, closer, err := s.db.Get(key)
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	defer closer.Close()

	entryPoint := bytesToInt64(value)
	return &entryPoint, nil
}

// SetEntryPoint stores the entry point in storage
func (s *Storage) SetEntryPoint(workspace int64, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := makeKey(workspace, KeyTypeEntry, 0)
	value := int64ToBytes(id)
	return s.db.Set(key, value, pebble.Sync)
}

// Close closes the storage and its underlying database
func (s *Storage) Close() error {
	s.cache.Clear()
	return s.db.Close()
}
