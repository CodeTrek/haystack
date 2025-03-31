package leveldb

import (
	"fmt"
	"strings"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Snap struct {
	db   *DB
	snap *leveldb.Snapshot // A snapshot of the database to allow concurrent read operations
}

// Get retrieves the value for a key from the snapshot
func (s *Snap) Get(key []byte) ([]byte, error) {
	if s.snap == nil {
		return nil, fmt.Errorf("snapshot is closed")
	}

	// Read from the snapshot
	value, err := s.snap.Get(key, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get data from snapshot: %v", err)
	}
	return value, nil
}

// Scan performs a range scan over the snapshot
func (s *Snap) Scan(prefix []byte, cb func(key, value []byte) bool) error {
	if s.snap == nil {
		return fmt.Errorf("snapshot is closed")
	}

	// Use the snapshot for the scan
	iter := s.snap.NewIterator(util.BytesPrefix(prefix), nil)
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
