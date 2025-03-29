package indexer

import (
	"fmt"
	"search-indexer/running"
	"sync"
)

type Indexer struct {
}

func Run(wg *sync.WaitGroup) {
	fmt.Println("Starting indexer...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		indexerMain()
		running.WaitingForShutdown()
	}()
}

func indexerMain() {
	demo()
}
