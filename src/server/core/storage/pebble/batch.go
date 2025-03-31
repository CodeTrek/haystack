package pebble

import (
	"fmt"
	"sync"

	"github.com/cockroachdb/pebble"
)

// Batch represents a batch of operations
type Batch struct {
	db    *DB
	batch *pebble.Batch
	mu    sync.Mutex
}

// Put adds a key-value pair to the batch
func (b *Batch) Put(key, value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Set(key, value, nil); err != nil {
		return fmt.Errorf("failed to put data in batch: %v", err)
	}
	return nil
}

// Delete adds a delete operation to the batch
func (b *Batch) Delete(key []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Delete(key, nil); err != nil {
		return fmt.Errorf("failed to delete data in batch: %v", err)
	}
	return nil
}

// Commit commits the batch to the database
func (b *Batch) Commit() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Commit(pebble.Sync); err != nil {
		return fmt.Errorf("failed to commit batch: %v", err)
	}
	return nil
}

// Reset resets the batch for reuse
func (b *Batch) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.batch.Reset()
}

// Close closes the batch
func (b *Batch) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Close(); err != nil {
		return fmt.Errorf("failed to close batch: %v", err)
	}
	return nil
}
