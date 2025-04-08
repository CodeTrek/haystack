package client

import (
	"flag"
	"fmt"
	"haystack/shared/running"
)

func handleInstall(args []string) {
	// Create a new flag set for install command
	installFlags := flag.NewFlagSet("install", flag.ExitOnError)
	force := installFlags.Bool("f", false, "Force installation, overwrite existing files")

	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("Usage: " + running.ExecutableName() + " install [options]")
		fmt.Println("Options:")
		installFlags.PrintDefaults()
		return
	}

	// Parse the flags
	if err := installFlags.Parse(args); err != nil {
		fmt.Println("Error parsing flags:", err)
		return
	}

	if running.IsDevVersion() {
		fmt.Println("Dev version is not supported for installation")
		return
	}

	isRunning := running.IsServerRunning()
	if isRunning {
		fmt.Println("Stopping server...")
		handleServerStop()
		if running.IsServerRunning() {
			fmt.Println("Failed to stop server")
			return
		}
	}

	installServer(*force)

	fmt.Println("Install completed")
	if isRunning {
		fmt.Println("Starting server...")
		handleServerStart()
	}
}

func installServer(_ bool) {
	fmt.Println("Not implemented yet!")
}
