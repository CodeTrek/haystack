package main

import (
	"flag"
	"haystack/client"
	"haystack/conf"
	"haystack/server"
	"haystack/shared/running"
	"log"
	"os"
	"path/filepath"
)

var version = "dev"

func main() {
	conf.SetVersion(version)

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
		client.Run()
	}
}
