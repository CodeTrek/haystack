package main

import (
	"flag"
	"haystack/client"
	"haystack/conf"
	"haystack/server"
	"haystack/shared/running"
	"log"
	_ "net/http/pprof"
	"path/filepath"
)

var version = "dev"

func main() {
	running.SetVersion(version)

	flag.Parse()
	if err := conf.Load(); err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	lockFile := filepath.Join(conf.Get().Global.DataPath, "server.lock")
	running.RegisterLockFile(lockFile)

	if running.IsDaemonMode() {
		server.Run()
		if running.IsRestart() {
			running.StartNewServer()
		}
	} else {
		client.Run()
	}
}
