package storage

import (
	"fmt"
	"haystack/server/core/storage/pebble"
	"sync/atomic"
)

var putCount = atomic.Int64{}
var deleteCount = atomic.Int64{}

func NewBatchWrite(db *pebble.DB) *BatchWrite {
	batch := db.Batch()
	if batch == nil {
		return nil
	}

	return &BatchWrite{
		batch: batch,
		count: atomic.Int32{},
	}
}

type BatchWrite struct {
	batch *pebble.Batch
	count atomic.Int32
}

// Put adds a key-value pair to the batch
func (b *BatchWrite) Put(key, value []byte) error {
	putCount.Add(1)

	if err := b.batch.Put(key, value); err != nil {
		return fmt.Errorf("failed to put data in batch: %v", err)
	}
	return b.increaseAndTryCommit()
}

// Delete adds a delete operation to the batch
func (b *BatchWrite) Delete(key []byte) error {
	deleteCount.Add(1)

	if err := b.batch.Delete(key); err != nil {
		return fmt.Errorf("failed to delete data in batch: %v", err)
	}

	return b.increaseAndTryCommit()
}

// Commit commits the batch to the database
func (b *BatchWrite) Commit() error {
	if b.count.Load() == 0 {
		return nil
	}

	if err := b.batch.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch: %v", err)
	}
	return nil
}

func (b *BatchWrite) increaseAndTryCommit() error {
	b.count.Add(1)
	if b.count.Load() >= 512 {
		err := b.batch.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit batch: %v", err)
		}

		b.count.Store(0)
		b.batch.Reset()
	}

	return nil
}
