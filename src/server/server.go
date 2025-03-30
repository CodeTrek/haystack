package server

import (
	"fmt"
	"log"
	"search-indexer/running"
	"search-indexer/server/conf"
	"search-indexer/server/core/parser"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"search-indexer/server/indexer"
	"search-indexer/server/searcher"
	"sync"
)

func Run() {
	fmt.Println("Starting search indexer...")

	if err := conf.Load(); err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	wg := &sync.WaitGroup{}
	running.InitShutdown(wg)

	if err := storage.Init(wg); err != nil {
		log.Fatal("Error initializing storage:", err)
		running.Shutdown()
		return
	}

	if err := workspace.Init(wg); err != nil {
		log.Fatal("Error initializing workspace:", err)
		running.Shutdown()
		return
	}

	parser.Init()
	indexer.Run(wg)
	searcher.Run(wg)

	if conf.Get().ForTest.Path != "" {
		indexer.SyncIfNeeded(conf.Get().ForTest.Path)
	}

	wg.Wait()
}
