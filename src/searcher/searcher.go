package searcher

import (
	"context"
	"fmt"
	"sync"
)

func Run(shutdown context.Context, wg *sync.WaitGroup) {
	fmt.Println("Starting searcher...")

	wg.Add(1)
	go func() {
		defer wg.Done()

		<-shutdown.Done()
	}()
}
