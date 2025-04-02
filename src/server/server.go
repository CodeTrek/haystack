package server

import (
	"fmt"
	"log"
	"search-indexer/conf"
	"search-indexer/running"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"search-indexer/server/indexer"
	"search-indexer/server/searcher"
	"search-indexer/server/server"
	"sync"
)

func Run() {
	log.Println("Starting search indexer...")

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

	server.StartServer(wg, fmt.Sprintf("127.0.0.1:%d", conf.Get().Global.Port))

	wg.Wait()
	storage.CloseAndWait()
}
