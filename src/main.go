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
	running.SetVersion(version)

	flag.Parse()
	if err := running.Init(); err != nil {
		os.Exit(1)
	}

	if err := conf.Load(); err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	lockFile := filepath.Join(conf.Get().Global.HomePath, "server.lock")
	running.RegisterLockFile(lockFile)

	if running.IsDaemonMode() {
		if !running.IsDevVersion() {
			if running.InstallPath() != running.ExecutablePath() {
				log.Fatal("Haystack daemon is running in a non-standard path. Please run `haystack install` to install it in the standard path.")
				os.Exit(1)
			}
		}
		server.Run()
		if running.IsRestart() {
			running.StartNewServer()
		}
	} else {
		client.Run()
	}
}
