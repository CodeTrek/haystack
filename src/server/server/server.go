package server

import (
	"context"
	"haystack/shared/running"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// StartServer initializes and starts the HTTP server
func StartServer(wg *sync.WaitGroup, addr string) {
	wg.Add(1)
	defer wg.Done()

	var shuttingDown atomic.Bool
	server := &http.Server{
		Addr: addr,
	}

	http.HandleFunc("/", http.NotFound)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/api/v1/server/restart", handleRestart)
	http.HandleFunc("/api/v1/server/stop", handleStop)
	http.HandleFunc("/api/v1/server/status", handleStatus)

	http.HandleFunc("/api/v1/workspace/create", handleCreateWorkspace)
	http.HandleFunc("/api/v1/workspace/delete", handleDeleteWorkspace)
	http.HandleFunc("/api/v1/workspace/list", handleListWorkspace)

	http.HandleFunc("/api/v1/search/content", handleSearchContent)

	// Start server in a goroutine
	go func() {
		log.Printf("HTTP server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	// Wait for shutdown signal
	<-running.GetShutdown().Done()
	shuttingDown.Store(true)

	// Create shutdown context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("HTTP server exiting")
}
