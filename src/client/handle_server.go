package client

import (
	"encoding/json"
	"fmt"
	"haystack/shared/running"
	"haystack/shared/types"
	"time"
)

func handleServer(args []string) {
	if len(args) < 1 || args[0] == "-h" {
		fmt.Println("Usage: " + running.ExecutableName() + " server <command>")
		fmt.Println("Commands:")
		fmt.Println("  status    Show server status")
		fmt.Println("  start     Start the server")
		fmt.Println("  stop      Stop the server")
		fmt.Println("  restart   Restart the server")
		return
	}

	command := args[0]
	switch command {
	case "status":
		handleServerStatus()
	case "start":
		handleServerStart()
	case "stop":
		handleServerStop()
	case "restart":
		handleServerRestart()
	default:
		fmt.Printf("Unknown server command: %s\n", command)
		fmt.Println("Available commands: status, start, stop, restart")
	}
}

func getRunningState() (*types.ServerStatus, error) {
	result, err := serverRequest("/server/status", []byte{})
	if err != nil {
		return nil, fmt.Errorf("error getting server status: %v", err)
	}

	var status types.ServerStatus
	if err := json.Unmarshal(*result.Body.Data, &status); err != nil {
		return nil, fmt.Errorf("error unmarshalling server status: %v", err)
	}

	return &status, nil
}

func handleServerRestart() {
	if !running.IsServerRunning() {
		running.StartNewServer()
		return
	}

	status, err := getRunningState()
	if err != nil {
		fmt.Printf("Error getting server status: %v\n", err)
		return
	}

	if status.ShuttingDown || status.Restarting {
		fmt.Println("Server is shutting down or restarting")
		return
	}

	_, err = serverRequest("/server/restart", []byte{})
	if err != nil {
		fmt.Printf("Error restarting server: %v\n", err)
		return
	}

	fmt.Println("Server restarted")
}

func handleServerStatus() {
	if !running.IsServerRunning() {
		fmt.Println("Server is not running")
		return
	}

	status, err := getRunningState()
	if err != nil {
		fmt.Printf("Error getting server status: %v\n", err)
		return
	}

	fmt.Printf(`Server status:
  PID: %d
  Version: %s
  Is shutting down: %t
  Is restarting: %t
	`, status.PID, status.Version, status.ShuttingDown, status.Restarting)
}

func handleServerStart() {
	if running.IsServerRunning() {
		fmt.Println("Server is already running")
		return
	}
	running.StartNewServer()
}

func handleServerStop() {
	if !running.IsServerRunning() {
		fmt.Println("Server is not running")
		return
	}
	_, err := serverRequest("/server/stop", []byte{})
	if err != nil {
		fmt.Printf("Error stopping server: %v\n", err)
		return
	}

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			fmt.Println("Server did not stop in time")
			return
		default:
			time.Sleep(1 * time.Second)
			if !running.IsServerRunning() {
				fmt.Println("Server stopped")
				return
			}
		}
	}
}
