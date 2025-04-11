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

// DeleteRange deletes a range of keys in the batch
func (b *Batch) DeleteRange(start, end []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.DeleteRange(start, end, nil); err != nil {
		return fmt.Errorf("failed to delete range in batch: %v", err)
	}
	return nil
}

// DeletePrefix deletes all keys with the given prefix in the batch
func (b *Batch) DeletePrefix(prefix []byte) error {
	return b.DeleteRange(prefix, append(prefix, 0xFF))
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
