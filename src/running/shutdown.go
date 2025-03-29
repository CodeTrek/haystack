package running

import (
	"context"
	"fmt"
)

var shutdown context.Context

func InitShutdown() (func(), error) {
	if shutdown != nil {
		return func() {}, fmt.Errorf("shutdown already initialized")
	}

	s, cancel := context.WithCancel(context.Background())
	shutdown = s
	return cancel, nil
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
