package main

import (
	"flag"
	"haystack/conf"
	"haystack/server"
	"haystack/shared/running"
	"log"
	"os"
	"path/filepath"
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
