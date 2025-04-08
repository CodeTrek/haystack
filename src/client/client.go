package client

import (
	"fmt"
	"os"

	"haystack/conf"
	"haystack/shared/running"
)

func Run() {
	// Check if there are enough command line arguments
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	processCommand(os.Args[1:])
}

func processCommand(args []string) {
	command := args[0]

	switch command {
	case "search":
		handleSearch(args[1:])
	case "workspace":
		handleWorkspace(args[1:])
	case "server":
		handleServer(args[1:])
	case "version":
		fmt.Println("Version:", conf.Version())
	case "help":
		if len(args) > 1 {
			processCommand(append(args[1:], "-h"))
		} else {
			printUsage()
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: " + running.ExecutableName() + " <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  version               Show current version")
	fmt.Println("  search <query>        Search for documents matching the query")
	fmt.Println("  server <commands>     Server commands")
	fmt.Println("  workspace <commands>  Workspace commands")
	fmt.Println("  help <command>        Show help for a specific command")
}
