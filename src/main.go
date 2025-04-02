package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"search-indexer/conf"
	"search-indexer/server"
	"search-indexer/shared/running"
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
		server.Run(lockFile)
		if running.IsRestart() {
			running.StartNewServer()
		}
	} else {
		log.Fatal("Client mode not implemented yet")
	}
}
