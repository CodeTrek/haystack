package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"search-indexer/conf"
	"search-indexer/running"
	"search-indexer/server"
)

func main() {
	flag.Parse()
	if err := running.Init(); err != nil {
		os.Exit(1)
	}

	if err := conf.Load(); err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	lockFile := filepath.Join(conf.Get().Global.HomePath, "server.lock")
	if running.IsServerMode() {
		cleanup, err := running.CheckAndLockServer(lockFile)
		if err != nil {
			log.Fatal("Error locking and running as server:", err)
			return
		}
		defer cleanup()

		server.Run()
	} else {
		log.Fatal("Client mode not implemented yet")
	}
}
