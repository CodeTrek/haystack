package leveldb

import (
	"bytes"
	"testing"
)

func TestDB(t *testing.T) {
	// Create a temporary database
	dbPath := t.TempDir()
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test Put and Get
	key := []byte("test-key")
	value := []byte("test-value")

	if err := db.Put(key, value); err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	result, err := db.Get(key)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if !bytes.Equal(result, value) {
		t.Errorf("Expected value %q, got %q", value, result)
	}

	// Test non-existent key
	result, err = db.Get([]byte("non-existent"))
	if err != nil {
		t.Fatalf("Failed to get non-existent key: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil for non-existent key, got %q", result)
	}

	// Test Delete
	if err := db.Delete(key); err != nil {
		t.Fatalf("Failed to delete data: %v", err)
	}

	result, err = db.Get(key)
	if err != nil {
		t.Fatalf("Failed to get deleted key: %v", err)
	}
	if result != nil {
		t.Errorf("Expected nil for deleted key, got %q", result)
	}
}

func TestBatch(t *testing.T) {
	// Create a temporary database
	dbPath := t.TempDir()
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create batch operations
	batch := db.Batch()
	keys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
	}
	values := [][]byte{
		[]byte("value1"),
		[]byte("value2"),
		[]byte("value3"),
	}

	// Add operations to batch
	for i := range keys {
		batch.Put(keys[i], values[i])
	}

	// Execute batch
	if err := batch.Write(); err != nil {
		t.Fatalf("Failed to write batch: %v", err)
	}

	// Verify results
	for i := range keys {
		result, err := db.Get(keys[i])
		if err != nil {
			t.Fatalf("Failed to get key %q: %v", keys[i], err)
		}
		if !bytes.Equal(result, values[i]) {
			t.Errorf("Expected value %q for key %q, got %q", values[i], keys[i], result)
		}
	}
}

func TestScan(t *testing.T) {
	// Create a temporary database
	dbPath := t.TempDir()
	db, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Insert test data with prefix
	prefix := []byte("test:")
	data := map[string]string{
		"test:1":  "value1",
		"test:2":  "value2",
		"test:3":  "value3",
		"other:1": "other1",
	}

	for k, v := range data {
		if err := db.Put([]byte(k), []byte(v)); err != nil {
			t.Fatalf("Failed to put data: %v", err)
		}
	}

	// Test scanning with prefix
	results, err := db.Scan(prefix, 0)
	if err != nil {
		t.Fatalf("Failed to scan: %v", err)
	}

	// Verify results
	expectedCount := 3 // number of items with "test:" prefix
	if len(results) != expectedCount {
		t.Errorf("Expected %d results, got %d", expectedCount, len(results))
	}

	for _, kv := range results {
		key := string(kv[0])
		value := string(kv[1])
		expectedValue, exists := data[key]
		if !exists {
			t.Errorf("Unexpected key in results: %s", key)
			continue
		}
		if value != expectedValue {
			t.Errorf("Expected value %q for key %q, got %q", expectedValue, key, value)
		}
	}

	// Test scanning with limit
	limit := 2
	results, err = db.Scan(prefix, limit)
	if err != nil {
		t.Fatalf("Failed to scan with limit: %v", err)
	}

	if len(results) != limit {
		t.Errorf("Expected %d results with limit, got %d", limit, len(results))
	}
}
