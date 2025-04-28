package server

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/codetrek/haystack/conf"
	"github.com/codetrek/haystack/server/core/fulltext"
	"github.com/codetrek/haystack/server/core/workspace"
	"github.com/codetrek/haystack/server/indexer"
	"github.com/codetrek/haystack/server/searcher"
	"github.com/codetrek/haystack/server/server"
	"github.com/codetrek/haystack/shared/running"
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

	if err := fulltext.Init(); err != nil {
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
	fulltext.CloseAndWait()

	log.Println("Haystack server stopped")
}
