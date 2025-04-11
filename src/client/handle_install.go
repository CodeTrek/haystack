package client

import (
	"flag"
	"fmt"
	"haystack/shared/running"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	fsutils "haystack/utils/fs"
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

	if running.InstallPath() == running.ExecutablePath() {
		fmt.Println("Haystack is already installed.")
		return
	}

	// Get current version
	newVersion := running.Version()

	// Get install path
	installTarget := filepath.Join(running.InstallPath(), running.ExecutableName())

	// Run version command to get installed version
	installedVersion := getInstalledVersion(installTarget)

	if !*force && !isUpgradeNeeded(installedVersion, newVersion) {
		fmt.Println("Haystack is already up to date.")
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

	if err := installServer(running.Executable(), installTarget); err != nil {
		fmt.Println("Failed to install server:", err)
		return
	}

	fmt.Println("Install completed")
	if isRunning {
		fmt.Println("Starting server...")
		handleServerStart()
	}
}

func installServer(executable, installTarget string) error {
	os.MkdirAll(filepath.Dir(installTarget), 0755)
	if err := fsutils.CopyFile(executable, installTarget); err != nil {
		return err
	}

	os.Chmod(installTarget, 0755)
	return nil
}

func getInstalledVersion(installTarget string) string {
	_, err := os.Stat(installTarget)
	if err != nil {
		return ""
	}

	cmd := exec.Command(installTarget, "version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}

func isUpgradeNeeded(installedVersion, newVersion string) bool {
	if installedVersion == "" {
		return true
	}

	// version is in the format 0.0.0
	installedVersionParts := strings.Split(installedVersion, ".")
	newVersionParts := strings.Split(newVersion, ".")

	for i := range installedVersionParts {
		// Convert to int
		installedVersionInt, _ := strconv.Atoi(installedVersionParts[i])
		newVersionInt, _ := strconv.Atoi(newVersionParts[i])
		if installedVersionInt < newVersionInt {
			return true
		}
	}

	return false
}
