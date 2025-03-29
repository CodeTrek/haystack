package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"search-indexer/running"
	"search-indexer/server/conf"
	"search-indexer/server/core/parser"
	"search-indexer/server/core/storage"
	"search-indexer/server/indexer"
	"search-indexer/server/searcher"
	"sync"
	"syscall"
)

func Run() {
	fmt.Println("Starting search indexer...")

	if err := conf.Load(); err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	cancel, _ := running.InitShutdown()
	wg := &sync.WaitGroup{}

	if err := storage.Init(wg); err != nil {
		log.Fatal("Error initializing storage:", err)
		return
	}

	parser.Init()
	indexer.Run(wg)
	searcher.Run(wg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	cancel()
	wg.Wait()
}
