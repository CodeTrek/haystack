package indexer

import (
	"context"
	"fmt"
	"sync"
)

type Indexer struct {
}

func Run(shutdown context.Context, wg *sync.WaitGroup) {
	fmt.Println("Starting indexer...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		indexerMain(shutdown)
		<-shutdown.Done()
	}()
}

func indexerMain(shutdown context.Context) {
	demo(shutdown)
}
