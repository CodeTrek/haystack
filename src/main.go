package main

import (
	"flag"
	"log"
	_ "net/http/pprof"
	"path/filepath"

	"github.com/codetrek/haystack/client"
	"github.com/codetrek/haystack/conf"
	"github.com/codetrek/haystack/server"
	"github.com/codetrek/haystack/shared/running"
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
