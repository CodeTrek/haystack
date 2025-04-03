package storage

import (
	"context"
	"haystack/conf"
	"haystack/server/core/storage/pebble"
	"haystack/shared/running"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var db *pebble.DB

const StorageVersion = "1.0"

var closeOnce sync.Once

func Init() error {
	homePath := conf.Get().Global.HomePath
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

	go func() {
		timer := time.NewTicker(1 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-running.GetShutdown().Done():
				return
			case <-timer.C:
				flushPendingWrites(false)
			}
		}
	}()

	return nil
}

func CloseAndWait() {
	closeOnce.Do(func() {
		flushPendingWrites(true)

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
