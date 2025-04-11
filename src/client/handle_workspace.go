package client

import (
	"encoding/json"
	"fmt"
	"haystack/shared/running"
	"haystack/shared/types"
	"path/filepath"
)

func handleWorkspace(args []string) {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		fmt.Println("Usage: " + running.ExecutableName() + " workspace <command>")
		fmt.Println("Commands:")
		fmt.Println("  list                  List workspaces")
		fmt.Println("  get <path>            Get a workspace")
		fmt.Println("  create                Create a new workspace")
		fmt.Println("  delete <path>         Delete a workspace")
		fmt.Println("  sync-all              Sync all workspaces")
		fmt.Println("  sync <path>           Sync a workspace")
		return
	}

	command := args[0]
	switch command {
	case "list":
		handleWorkspaceList()
	case "create":
		handleWorkspaceCreate()
	case "delete":
		handleWorkspaceDelete(args[1])
	case "sync-all":
		handleWorkspaceSyncAll()
	case "sync":
		handleWorkspaceSync(args[1])
	case "get":
		handleWorkspaceGet(args[1])
	default:
		fmt.Printf("Unknown workspace command: %s\n", command)
		fmt.Println("Available commands: get, list, create, delete, sync, sync-all")
	}
}

func handleWorkspaceSyncAll() {
	result, err := serverRequest("/workspace/sync-all", []byte{})
	if err != nil {
		fmt.Printf("Error syncing all workspaces: %v\n", err)
		return
	}

	fmt.Println("Message:", result.Body.Message)
}

func handleWorkspaceSync(workspacePath string) {
	if workspacePath == "" {
		fmt.Println("Usage: " + running.ExecutableName() + " workspace sync <workspace path>")
		return
	}
	if !filepath.IsAbs(workspacePath) {
		fmt.Println("Workspace path must be absolute")
		return
	}

	request := types.SyncWorkspaceRequest{
		Workspace: workspacePath,
	}

	requestJson, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error syncing workspace: %v\n", err)
		return
	}

	result, err := serverRequest("/workspace/sync", requestJson)
	if err != nil {
		fmt.Printf("Error syncing workspace: %v\n", err)
		return
	}

	fmt.Println("Message:", result.Body.Message)
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
		printWorkspace("", workspace)
	}
}

func handleWorkspaceGet(workspacePath string) {
	request := types.GetWorkspaceRequest{
		Workspace: workspacePath,
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error getting workspace: %v\n", err)
		return
	}

	result, err := serverRequest("/workspace/get", requestJson)
	if err != nil {
		fmt.Printf("Error getting workspace: %v\n", err)
		return
	}

	var workspace types.Workspace
	if err := json.Unmarshal(*result.Body.Data, &workspace); err != nil {
		fmt.Printf("Error getting workspace: %v\n", err)
		return
	}

	printWorkspace("", workspace)
}

func printWorkspace(prefix string, workspace types.Workspace) {
	fmt.Printf(`%s %s:
  Path: %s
  Created at: %s
  Last accessed: %s
  Last full sync: %s
  Total files: %d
  Indexing: %t
`,
		prefix, workspace.ID, workspace.Path, workspace.CreatedAt, workspace.LastAccessed, workspace.LastFullSync,
		workspace.TotalFiles, workspace.Indexing)
}

func handleWorkspaceCreate() {
	fmt.Println("Not implemented yet!")
}

func handleWorkspaceDelete(workspacePath string) {
	if workspacePath == "" {
		fmt.Println("Usage: " + running.ExecutableName() + " workspace delete <workspace path>")
		return
	}

	request := types.DeleteWorkspaceRequest{
		Workspace: workspacePath,
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error deleting workspace: %v\n", err)
		return
	}

	result, err := serverRequest("/workspace/delete", requestJson)
	if err != nil {
		fmt.Printf("Error deleting workspace: %v\n", err)
		return
	}

	var response types.Workspace
	if err := json.Unmarshal(*result.Body.Data, &response); err != nil {
		fmt.Printf("Error deleting workspace: %v\n", err)
		return
	}

	printWorkspace("Deleted", response)
}
