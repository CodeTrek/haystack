package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"search-indexer/server/conf"
	"search-indexer/server/core/parser"
	"search-indexer/server/indexer"
	"search-indexer/server/searcher"
	"sync"
	"syscall"
)

func Run() {
	fmt.Println("Starting search indexer...")

	if err := conf.Load(); err != nil {
		log.Fatal("Error loading config:", err)
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
