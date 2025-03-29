package main

import (
	"flag"
	"log"
	"os"
	"search-indexer/running"
	"search-indexer/server"
)

func main() {
	flag.Parse()
	if err := running.Init(); err != nil {
		os.Exit(1)
	}

	if running.IsServerMode() {
		server.Run()
	} else {
		log.Fatal("Client mode not implemented yet")
	}
}
