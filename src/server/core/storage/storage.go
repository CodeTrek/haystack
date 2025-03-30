package storage

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"search-indexer/running"
	"search-indexer/server/core/storage/leveldb"
	"sync"
	"time"
)

var db *leveldb.DB

const StorageVersion = "1.0"

var closeOnce sync.Once

func Init() error {
	dataPath := filepath.Join(running.RootPath(), "data")
	dbPath := filepath.Join(dataPath, "leveldb")
	versionPath := filepath.Join(dataPath, "version")

	os.MkdirAll(dataPath, 0755)
	os.WriteFile(versionPath, []byte(StorageVersion), 0644)

	var err error
	db, err = leveldb.OpenDB(dbPath)
	if err != nil {
		return err
	}

	return nil
}

func CloseAndWait() {
	closeOnce.Do(func() {
		log.Println("Closing storage...")
		defer log.Println("Storage closed.")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				log.Println("Storage close timeout, force quiting...")
				return
			case <-time.After(1 * time.Second):
			}

			db.Close()
			if db.IsClosed() {
				break
			}
			log.Println("Waiting for storage to be closed...")
		}
	})
}
