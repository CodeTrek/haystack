package server

import (
	"log"
	"search-indexer/conf"
	"search-indexer/running"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"search-indexer/server/indexer"
	"search-indexer/server/searcher"
	"sync"
)

func Run() {
	log.Println("Starting search indexer...")

	if err := conf.Load(); err != nil {
		log.Fatal("Error loading config:", err)
		return
	}

	initLog()

	wg := &sync.WaitGroup{}
	running.InitShutdown(wg)

	if err := storage.Init(); err != nil {
		log.Fatal("Error initializing storage:", err)
		running.Shutdown()
		return
	}

	if err := workspace.Init(wg); err != nil {
		log.Fatal("Error initializing workspace:", err)
		running.Shutdown()
		return
	}

	indexer.Run(wg)
	searcher.Run(wg)

	if conf.Get().ForTest.Path != "" {
		indexer.SyncIfNeeded(conf.Get().ForTest.Path)
	}

	wg.Wait()
	storage.CloseAndWait()
}
