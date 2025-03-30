package indexer

import (
	"search-indexer/running"
	"sync"
)

type Writer struct {
}

func (w *Writer) start(wg *sync.WaitGroup) {
	wg.Add(1)
	go w.run(wg)
}

func (w *Writer) run(wg *sync.WaitGroup) {
	defer wg.Done()
	running.WaitingForShutdown()
}
