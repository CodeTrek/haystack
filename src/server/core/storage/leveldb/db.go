package leveldb

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// DB represents a LevelDB database instance
type DB struct {
	path   string
	db     *leveldb.DB
	closed bool

	snap        *leveldb.Snapshot // A snapshot of the database to allow concurrent read operations
	activeSnaps map[*leveldb.Snapshot]int
	mutex       sync.RWMutex
}

// OpenDB opens a LevelDB database at the specified path
func OpenDB(path string) (*DB, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	db, err := leveldb.OpenFile(absPath, &opt.Options{
		Compression:            opt.SnappyCompression,
		WriteBuffer:            8 * opt.MiB,
		BlockSize:              16 * opt.KiB,
		CompactionTableSize:    4 * opt.MiB,
		Filter:                 filter.NewBloomFilter(10), // 10 bits per key
		CompactionL0Trigger:    24,
		WriteL0PauseTrigger:    32,
		WriteL0SlowdownTrigger: 28,

		// CompactionTableSizeMultiplier: 1.2,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open leveldb: %v", err)
	}

	// Create a new DB instance
	ldb := &DB{
		path:   absPath,
		db:     db,
		closed: false,

		activeSnaps: make(map[*leveldb.Snapshot]int),
	}

	/*
		// Create an initial snapshot
		if err := ldb.TakeSnapshot(); err != nil {
			db.Close()
			return nil, err
		}

		go func() {
			for {
				time.Sleep(1 * time.Second)
				if ldb.IsClosed() {
					return
				}
				ldb.TakeSnapshot()
			}
		}()
	*/

	return ldb, nil
}

// Close closes the database
func (d *DB) Close() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.closed {
		return fmt.Errorf("database is closed")
	}

	// If there are active snapshots, we cannot close the database
	if len(d.activeSnaps) > 1 || d.activeSnaps[d.snap] > 1 {
		return fmt.Errorf("cannot close database with %d active snapshots", len(d.activeSnaps))
	}

	d.releaseSnapInternal(d.snap)
	d.snap = nil
	d.closed = true

	if d.db != nil {
		if err := d.db.Close(); err != nil {
			return fmt.Errorf("failed to close leveldb: %v", err)
		}
		d.db = nil
	}
	return nil
}

func (d *DB) IsClosed() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.closed
}

/*
// TakeSnapshot releases the current snapshot and creates a new one
func (d *DB) TakeSnapshot() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.closed {
		return fmt.Errorf("database is closed")
	}

	// Create a new snapshot
	snap, err := d.db.GetSnapshot()
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %v", err)
	}

	// Release the current snapshot if it exists
	d.releaseSnapInternal(d.snap)
	d.snap = snap

	// Add the new snapshot to the active snapshots map
	d.activeSnaps[snap] = 1

	return nil
}

// GetSnapshot returns the current snapshot
// This is useful for advanced operations that need direct access to the snapshot
func (d *DB) GetSnapshot() (*Snap, func()) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.closed {
		return nil, func() {}
	}

	d.activeSnaps[d.snap]++

	snap := &Snap{
		db:   d,
		snap: d.snap,
	}

	return snap, func() {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		d.releaseSnapInternal(snap.snap)
		snap.snap = nil
	}
}
*/

// Put stores a key-value pair
func (d *DB) Put(key, value []byte) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.closed {
		return fmt.Errorf("database is closed")
	}

	options := &opt.WriteOptions{Sync: true}
	if err := d.db.Put(key, value, options); err != nil {
		return fmt.Errorf("failed to put data: %v", err)
	}
	return nil
}

// Get retrieves the value for a key
func (d *DB) Get(key []byte) ([]byte, error) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if d.closed {
		return nil, fmt.Errorf("database is closed")
	}

	// Read directly from the DB, not from the snapshot
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

	if d.closed {
		return fmt.Errorf("database is closed")
	}

	options := &opt.WriteOptions{Sync: true}
	if err := d.db.Delete(key, options); err != nil {
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
func (d *DB) Scan(prefix []byte, cb func(key, value []byte) bool) error {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	if d.closed {
		return fmt.Errorf("database is closed")
	}

	// Always use the DB directly, not the snapshot
	iter := d.db.NewIterator(util.BytesPrefix(prefix), nil)
	defer iter.Release()

	for iter.Next() {
		if !strings.HasPrefix(string(iter.Key()), string(prefix)) {
			break
		}

		if continueScan := cb(iter.Key(), iter.Value()); !continueScan {
			break
		}
	}
	return nil
}

func (d *DB) releaseSnapInternal(snap *leveldb.Snapshot) {
	if snap == nil {
		return
	}

	if count, exists := d.activeSnaps[snap]; exists && count > 1 {
		d.activeSnaps[snap]--
		return
	}

	snap.Release()
	delete(d.activeSnaps, snap)
}
