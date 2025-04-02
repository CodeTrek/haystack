package server

import (
	"fmt"
	"log"
	"search-indexer/conf"
	"search-indexer/server/core/storage"
	"search-indexer/server/core/workspace"
	"search-indexer/server/indexer"
	"search-indexer/server/searcher"
	"search-indexer/server/server"
	"search-indexer/shared/running"
	"sync"
)

func Run(lockFile string) {
	cleanup, err := running.CheckAndLockServer(lockFile)
	if err != nil {
		log.Fatal("Error locking and running as server:", err)
		return
	}
	defer cleanup()

	initLog()

	log.Println("Starting search indexer...")

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

	log.Println("Search indexer stopped")
}
