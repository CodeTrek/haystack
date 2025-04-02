package server

import (
	"context"
	"log"
	"net/http"
	"search-indexer/running"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Client represents a WebSocket client connection
type Client struct {
	conn *websocket.Conn
}

// Hub maintains the set of active clients
type Hub struct {
	clients map[*Client]bool

	register   chan *Client
	unregister chan *Client
	shutdown   chan struct{}
	mutex      sync.Mutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		shutdown:   make(chan struct{}),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			delete(h.clients, client)

		case <-h.shutdown:
			for client := range h.clients {
				client.conn.Close()
			}
			return
		}
	}
}

// handleConnection handles the WebSocket connection
func (c *Client) handleConnection(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Echo the message back
		if err := c.conn.WriteMessage(messageType, message); err != nil {
			log.Printf("error: %v", err)
			break
		}
	}
}

// ServeWs handles WebSocket requests from the peer
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
	}

	hub.register <- client
	go client.handleConnection(hub)
}

// StartServer initializes and starts the WebSocket server
func StartServer(wg *sync.WaitGroup, addr string) {
	wg.Add(1)
	defer wg.Done()

	hub := NewHub()
	go hub.Run()

	var shuttingDown atomic.Bool
	server := &http.Server{
		Addr: addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if shuttingDown.Load() {
				http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
				return
			}
			ServeWs(hub, w, r)
		}),
	}

	// Start server in a goroutine
	go func() {
		log.Printf("WebSocket server starting on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	// Wait for shutdown signal
	<-running.GetShutdown().Done()
	shuttingDown.Store(true)

	// Close all client connections
	hub.shutdown <- struct{}{}

	// Create shutdown context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("WebSocket server exiting")
}
