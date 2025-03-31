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
// @param:
//
//	sync: specify if the write should be synced to disk immediately, default is true for data safety.
//	     However, it will hurt the performance. Set to false if you want to improve the performance.
func (b *Batch) Write(sync bool) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.db.mutex.Lock()
	defer b.db.mutex.Unlock()

	if b.db.closed {
		return fmt.Errorf("database is closed")
	}

	opts := &opt.WriteOptions{
		Sync: sync,
	}

	if err := b.db.db.Write(b.batch, opts); err != nil {
		return err
	}
	return nil
}
