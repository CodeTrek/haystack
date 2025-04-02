package running

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	shutdown         context.Context
	cancel           func()
	initShutdownOnce sync.Once
	shutdownOnce     sync.Once
	restart          atomic.Bool

	ErrShutdown = errors.New("server is shutting down")
)

func InitShutdown(wg *sync.WaitGroup) {
	restart.Store(false)

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

func Restart() {
	restart.Store(true)
	Shutdown()
}

func IsRestart() bool {
	return restart.Load()
}

func StartNewServer() {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
		return
	}

	args := os.Args[1:]
	procAttr := &os.ProcAttr{
		Dir:   wd,
		Files: []*os.File{nil, os.Stdout, os.Stderr},
		Env:   os.Environ(),
	}

	process, err := os.StartProcess(executable, append([]string{executable}, args...), procAttr)
	if err != nil {
		log.Printf("Failed to start new process: %v", err)
		return
	}

	log.Printf("Started new process with PID: %d", process.Pid)
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
