package leveldb

import (
	"fmt"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// Batch represents a batch of operations
type Batch struct {
	db    *DB
	batch *leveldb.Batch
	mutex sync.Mutex
}

// Put adds a put operation to the batch
func (b *Batch) Put(key, value []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.batch.Put(key, value)
}

// Delete adds a delete operation to the batch
func (b *Batch) Delete(key []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.batch.Delete(key)
}

// Write executes the batch operations
func (b *Batch) Write() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.db.mutex.Lock()
	defer b.db.mutex.Unlock()

	if b.db.closed {
		return fmt.Errorf("database is closed")
	}

	if err := b.db.db.Write(b.batch, &opt.WriteOptions{Sync: true}); err != nil {
		return err
	}
	return nil
}
