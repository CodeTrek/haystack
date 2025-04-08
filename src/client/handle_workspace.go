package client

import (
	"encoding/json"
	"fmt"
	"haystack/shared/running"
	"haystack/shared/types"
)

func handleWorkspace(args []string) {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("Usage: " + running.ExecutableName() + " workspace <command>")
		fmt.Println("Commands:")
		fmt.Println("  list      List workspaces")
		fmt.Println("  create    Create a new workspace")
		fmt.Println("  delete    Delete a workspace")
		return
	}

	command := args[0]
	switch command {
	case "list":
		handleWorkspaceList()
	case "create":
		handleWorkspaceCreate()
	case "delete":
		handleWorkspaceDelete()
	default:
		fmt.Printf("Unknown workspace command: %s\n", command)
		fmt.Println("Available commands: list, create, delete")
	}
}

func handleWorkspaceList() {
	result, err := serverRequest("/workspace/list", []byte{})
	if err != nil {
		fmt.Printf("Error listing workspaces: %v\n", err)
		return
	}

	var workspaces types.Workspaces
	if err := json.Unmarshal(*result.Body.Data, &workspaces); err != nil {
		fmt.Printf("Error listing workspaces: %v\n", err)
		return
	}

	for _, workspace := range workspaces.Workspaces {
		fmt.Printf(`Workspace %s:
  Path: %s
  Created at: %s
  Last accessed: %s
  Last full sync: %s\n`,
			workspace.ID, workspace.Path, workspace.CreatedAt, workspace.LastAccessed, workspace.LastFullSync)
	}
}

func handleWorkspaceCreate() {
	fmt.Println("Not implemented yet!")
}

func handleWorkspaceDelete() {
	fmt.Println("Not implemented yet!")
}
