package running

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	shutdown         context.Context
	cancel           func()
	initShutdownOnce sync.Once
	shutdownOnce     sync.Once

	ErrShutdown = errors.New("server is shutting down")
)

func InitShutdown(wg *sync.WaitGroup) {
	initShutdownOnce.Do(func() {
		shutdown, cancel = context.WithCancel(context.Background())
		wg.Add(1)

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		go func() {
			defer wg.Done()

			select {
			case <-c:
				log.Println("Received interrupt signal, shutting down...")
				Shutdown()
			case <-shutdown.Done():
			}
		}()
	})
}

func Shutdown() {
	shutdownOnce.Do(func() {
		cancel()
	})
}

func GetShutdown() context.Context {
	return shutdown
}

func WaitingForShutdown() {
	<-shutdown.Done()
}

func IsShuttingDown() bool {
	select {
	case <-shutdown.Done():
		return true
	default:
		return false
	}
}
