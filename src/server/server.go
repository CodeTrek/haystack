package server

import (
	"fmt"
	"haystack/conf"
	"haystack/server/core/storage"
	"haystack/server/core/workspace"
	"haystack/server/indexer"
	"haystack/server/searcher"
	"haystack/server/server"
	"haystack/shared/running"
	"log"
	"strings"
	"sync"
)

func Run() {
	cleanup, err := running.CheckAndLockServer()
	if err != nil {
		log.Fatal("Error locking and running as server:", err)
		return
	}
	defer cleanup()

	initLog()

	log.Println(strings.Repeat("=", 64))
	log.Println("Starting haystack server...")

	wg := &sync.WaitGroup{}
	running.InitShutdown(wg)

	if err := storage.Init(); err != nil {
		log.Fatal("Error initializing storage:", err)
		running.Shutdown()
		return
	}

	if err := workspace.Init(); err != nil {
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

	log.Println("Haystack server stopped")
}
