package leveldb

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// DB represents a LevelDB database instance
type DB struct {
	db    *leveldb.DB
	path  string
	mutex sync.RWMutex
}

// OpenDB opens a LevelDB database at the specified path
func OpenDB(path string) (*DB, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	db, err := leveldb.OpenFile(absPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %v", err)
	}

	return &DB{
		db:   db,
		path: absPath,
	}, nil
}

// Close closes the database
func (d *DB) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.db != nil {
		if err := d.db.Close(); err != nil {
			return fmt.Errorf("failed to close leveldb: %v", err)
		}
		d.db = nil
	}
	return nil
}

// Put stores a key-value pair
func (d *DB) Put(key, value []byte) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.db.Put(key, value, nil); err != nil {
		return fmt.Errorf("failed to put data: %v", err)
	}
	return nil
}

// Get retrieves the value for a key
func (d *DB) Get(key []byte) ([]byte, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	value, err := d.db.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %v", err)
	}
	return value, nil
}

// Delete removes a key-value pair
func (d *DB) Delete(key []byte) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if err := d.db.Delete(key, nil); err != nil {
		return fmt.Errorf("failed to delete data: %v", err)
	}
	return nil
}

// Batch performs multiple operations in a single atomic batch
func (d *DB) Batch() *Batch {
	return &Batch{
		db:    d,
		batch: new(leveldb.Batch),
	}
}

// Scan performs a range scan over the database
func (d *DB) Scan(prefix []byte, limit int) ([][2][]byte, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	iter := d.db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()

	var results [][2][]byte
	for iter.Next() {
		if limit > 0 && len(results) >= limit {
			break
		}

		key := make([]byte, len(iter.Key()))
		value := make([]byte, len(iter.Value()))
		copy(key, iter.Key())
		copy(value, iter.Value())
		results = append(results, [2][]byte{key, value})
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("scan failed: %v", err)
	}

	return results, nil
}

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

	if err := b.db.db.Write(b.batch, &opt.WriteOptions{Sync: true}); err != nil {
		return fmt.Errorf("batch write failed: %v", err)
	}
	return nil
}
