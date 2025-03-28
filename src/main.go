package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"search-indexer/core/parser"
	"search-indexer/core/storage"
	"search-indexer/indexer"
	"search-indexer/searcher"
	"sync"
	"syscall"
)

func main() {
	fmt.Println("Starting search indexer...")

	if err := storage.Init(); err != nil {
		fmt.Println("Error initializing storage:", err)
		return
	}

	shutdown, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	parser.Init()
	indexer.Run(shutdown, wg)
	searcher.Run(shutdown, wg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	/*
		go func() {
			for {
				var input string
				_, err := fmt.Scanln(&input)
				if err != nil {
					if err.Error() == "unexpected newline" {
						continue
					}
					c <- os.Interrupt
					return
				}
				if input == "exit" {
					c <- os.Interrupt
					return
				}
			}
		}()
	*/

	<-c
	cancel()
	wg.Wait()
}
