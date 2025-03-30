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

func Init(wg *sync.WaitGroup) error {
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		running.WaitingForShutdown()
		log.Println("Closing storage...")

		for {
			if err := db.Close(); err != nil {
				log.Printf("failed to close storage: %v", err)
			}

			if db.IsClosed() {
				break
			}

			log.Println("Waiting for storage to be closed...")
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			<-ctx.Done()
		}

		log.Println("Storage closed.")
	}()

	return nil
}
