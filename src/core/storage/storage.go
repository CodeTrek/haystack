package storage

import (
	"log"
	"os"
	"path/filepath"
)

var storagePath string

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user's home directory: %v", err)
		return err
	}
	storagePath = filepath.Join(homeDir, ".search-indexer")
	if err := os.Mkdir(storagePath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("Failed to create storage directory: %v", err)
			return err
		}
	}

	return nil
}

func GetStoragePath() string {
	return storagePath
}

func RunningLock() {
	// lock := filepath.Join(storagePath, "running.lock")
}
