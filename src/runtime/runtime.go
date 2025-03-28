package runtime

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

var (
	rootPath   string
	serverMode = flag.Bool("server", false, "Run in server mode")
)

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user's home directory: %v", err)
		return err
	}

	rootPath = filepath.Join(homeDir, ".search-indexer")
	if err := os.Mkdir(rootPath, 0755); err != nil {
		if !os.IsExist(err) {
			log.Fatalf("Failed to create data directory: %v", err)
			return err
		}
	}

	return nil
}

func IsServerMode() bool {
	return *serverMode
}
