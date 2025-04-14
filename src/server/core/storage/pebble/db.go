package pebble

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
)

// DB represents a Pebble database instance
type DB struct {
	path   string
	db     *pebble.DB
	closed atomic.Bool
}

// OpenDB opens a Pebble database at the specified path
func OpenDB(path string) (*DB, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Configure Pebble options
	opts := &pebble.Options{
		WALMinSyncInterval: func() time.Duration {
			// Sync the WAL every 500us to avoid latency spikes.
			// Allow more operations to arrive and reduce IO operations
			return 500 * time.Microsecond
		},
		// Allow more files to be open
		MaxOpenFiles: 2000,

		// Set write buffer size to 8MB
		MemTableSize: 8 * 1024 * 1024,
		// Set max memtable count to 2
		MemTableStopWritesThreshold: 2,
		// Set L0 compaction threshold to 16
		L0CompactionThreshold: 16,
		// Set L0 stop writes threshold to 32
		L0StopWritesThreshold: 32,
		// Enable bloom filter
		Levels: []pebble.LevelOptions{
			{
				BlockSize:    16 * 1024,
				FilterPolicy: bloom.FilterPolicy(10),
			},
		},
	}

	db, err := pebble.Open(absPath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open pebble: %v", err)
	}

	// Create a new DB instance
	pdb := &DB{
		path:   absPath,
		db:     db,
		closed: atomic.Bool{},
	}
	pdb.closed.Store(false)

	return pdb, nil
}

// Close closes the database
func (d *DB) Close() error {
	if d.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	d.closed.Store(true)

	if d.db != nil {
		if err := d.db.Close(); err != nil {
			return fmt.Errorf("failed to close pebble: %v", err)
		}
		d.db = nil
	}
	return nil
}

func (d *DB) IsClosed() bool {
	return d.closed.Load()
}

// Put stores a key-value pair
func (d *DB) Put(key, value []byte) error {
	if d.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	// Use default write options (sync=true)
	if err := d.db.Set(key, value, pebble.Sync); err != nil {
		return fmt.Errorf("failed to put data: %v", err)
	}
	return nil
}

// Get retrieves the value for a key
func (d *DB) Get(key []byte) ([]byte, error) {
	if d.IsClosed() {
		return nil, fmt.Errorf("database is closed")
	}

	// Read directly from the DB
	value, closer, err := d.db.Get(key)
	if err == pebble.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get data: %v", err)
	}
	defer closer.Close()

	// Make a copy of the value since the original slice may be invalidated
	return append([]byte{}, value...), nil
}

// Delete removes a key-value pair
func (d *DB) Delete(key []byte) error {
	if d.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	// Use default write options (sync=true)
	if err := d.db.Delete(key, pebble.Sync); err != nil {
		return fmt.Errorf("failed to delete data: %v", err)
	}
	return nil
}

// Batch performs multiple operations in a single atomic batch
func (d *DB) Batch() *Batch {
	return &Batch{
		db:    d,
		batch: d.db.NewBatch(),
	}
}

// Scan performs a range scan over the database
// key and value will INVALIDATE after the callback
// so make sure to copy them if you need to use them later
// The callback should return true to continue scanning or false to stop
func (d *DB) Scan(prefix []byte, cb func(key, value []byte) bool) error {
	if d.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	// Create an iterator with the prefix
	iter, err := d.db.NewIter(&pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: append(prefix, 0xff),
	})
	if err != nil {
		return fmt.Errorf("failed to create iterator: %v", err)
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		// We're not going to copy silently for performance reasons.
		// It's user's responsibility to copy the key and value if they need to use them later.
		//
		// key := append([]byte{}, iter.Key()...)
		// value := append([]byte{}, iter.Value()...)

		if continueScan := cb(iter.Key(), iter.Value()); !continueScan {
			break
		}
	}
	return nil
}

// Scan performs a range scan over the database
// key and value will INVALIDATE after the callback
// so make sure to copy them if you need to use them later
// The callback should return true to continue scanning or false to stop
func (d *DB) ScanRange(begin []byte, end []byte, cb func(key, value []byte) bool) error {
	if d.IsClosed() {
		return fmt.Errorf("database is closed")
	}

	// Create an iterator with the prefix
	iter, err := d.db.NewIter(&pebble.IterOptions{
		LowerBound: begin,
		UpperBound: end,
	})
	if err != nil {
		return fmt.Errorf("failed to create iterator: %v", err)
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		// We're not going to copy silently for performance reasons.
		// It's user's responsibility to copy the key and value if they need to use them later.
		//
		// key := append([]byte{}, ...)
		// value := append([]byte{}, ...)

		if continueScan := cb(iter.Key(), iter.Value()); !continueScan {
			break
		}
	}
	return nil
}
