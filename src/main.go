package main

import (
	"flag"
	"log"
	"os"
	"search-indexer/runtime"
	"search-indexer/server"
)

func main() {
	flag.Parse()
	if err := runtime.Init(); err != nil {
		os.Exit(1)
	}

	if runtime.IsServerMode() {
		server.Run()
	} else {
		log.Fatal("Client mode not implemented yet")
	}
}
