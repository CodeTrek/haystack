package pebble

import (
	"fmt"
	"sync"

	"github.com/cockroachdb/pebble"
)

// Batch represents a batch of operations
type Batch interface {
	Put(key, value []byte) error
	Delete(key []byte) error
	DeleteRange(start, end []byte) error
	DeletePrefix(prefix []byte) error
	Commit() error
	Reset()
	Close() error
}

type PebbleBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
	mu    sync.Mutex
}

// Put adds a key-value pair to the batch
func (b *PebbleBatch) Put(key, value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Set(key, value, nil); err != nil {
		return fmt.Errorf("failed to put data in batch: %v", err)
	}
	return nil
}

// Delete adds a delete operation to the batch
func (b *PebbleBatch) Delete(key []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Delete(key, nil); err != nil {
		return fmt.Errorf("failed to delete data in batch: %v", err)
	}
	return nil
}

// DeleteRange deletes a range of keys in the batch
func (b *PebbleBatch) DeleteRange(start, end []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.DeleteRange(start, end, nil); err != nil {
		return fmt.Errorf("failed to delete range in batch: %v", err)
	}
	return nil
}

// DeletePrefix deletes all keys with the given prefix in the batch
func (b *PebbleBatch) DeletePrefix(prefix []byte) error {
	return b.DeleteRange(prefix, append(prefix, 0xFF))
}

// Commit commits the batch to the database
func (b *PebbleBatch) Commit() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Commit(pebble.Sync); err != nil {
		return fmt.Errorf("failed to commit batch: %v", err)
	}
	return nil
}

// Reset resets the batch for reuse
func (b *PebbleBatch) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.batch.Reset()
}

// Close closes the batch
func (b *PebbleBatch) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.batch.Close(); err != nil {
		return fmt.Errorf("failed to close batch: %v", err)
	}
	return nil
}
