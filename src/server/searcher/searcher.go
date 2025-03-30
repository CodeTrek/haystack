package searcher

import (
	"fmt"
	"log"
	"search-indexer/running"
	"sync"
)

func Run(wg *sync.WaitGroup) {
	fmt.Println("Starting searcher...")

	wg.Add(1)
	go func() {
		defer wg.Done()
		running.WaitingForShutdown()
		log.Println("Searcher shutdown.")
	}()
}
