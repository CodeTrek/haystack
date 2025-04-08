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
		fmt.Println("  list                  List workspaces")
		fmt.Println("  create                Create a new workspace")
		fmt.Println("  delete                Delete a workspace")
		fmt.Println("  get <workspace path>  Get a workspace")
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
	case "get":
		handleWorkspaceGet(args[1])
	default:
		fmt.Printf("Unknown workspace command: %s\n", command)
		fmt.Println("Available commands: list, create, delete, get")
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
		printWorkspace(workspace)
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

	printWorkspace(workspace)
}

func printWorkspace(workspace types.Workspace) {
	fmt.Printf(`Workspace %s:
  Path: %s
  Created at: %s
  Last accessed: %s
  Last full sync: %s
  Total files: %d
  Indexing: %t
`,
		workspace.ID, workspace.Path, workspace.CreatedAt, workspace.LastAccessed, workspace.LastFullSync,
		workspace.TotalFiles, workspace.Indexing)
}

func handleWorkspaceCreate() {
	fmt.Println("Not implemented yet!")
}

func handleWorkspaceDelete() {
	fmt.Println("Not implemented yet!")
}
