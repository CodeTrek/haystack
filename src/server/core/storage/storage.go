package storage

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"search-indexer/running"
	"search-indexer/server/conf"
	"search-indexer/server/core/storage/pebble"
	"sync"
	"time"
)

var db *pebble.DB

const StorageVersion = "1.0"

var closeOnce sync.Once

func Init() error {
	homePath := conf.Get().HomePath
	if homePath == "" {
		homePath = running.DefaultRootPath()
	}

	storagePath := filepath.Join(homePath, "data")

	log.Printf("Init storage path: %s", storagePath)

	dbPath := filepath.Join(storagePath, "index")
	versionPath := filepath.Join(storagePath, "version")

	os.MkdirAll(storagePath, 0755)
	os.WriteFile(versionPath, []byte(StorageVersion), 0644)

	var err error
	db, err = pebble.OpenDB(dbPath)
	if err != nil {
		return err
	}

	return nil
}

func CloseAndWait() {
	closeOnce.Do(func() {
		FlushPendingWrites(true)

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
