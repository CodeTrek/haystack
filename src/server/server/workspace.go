package server

import (
	"encoding/json"
	"fmt"
	"haystack/server/core/workspace"
	"haystack/server/indexer"
	"haystack/shared/types"
	"log"
	"net/http"
)

func handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	var request types.CreateWorkspaceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	ws, _ := workspace.GetByPath(request.Workspace)
	if ws != nil {
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    0,
			Message: "Workspace already exists",
		})
		return
	}

	ws, err = workspace.GetOrCreate(request.Workspace)
	if err != nil {
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: fmt.Sprintf("Failed to create workspace: %v", err),
		})
		return
	}

	indexer.SyncIfNeeded(ws.Path)

	log.Printf("Created workspace %s", request.Workspace)

	json.NewEncoder(w).Encode(types.CommonResponse{
		Code:    0,
		Message: "Ok",
	})
}

func handleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	var request types.DeleteWorkspaceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	ws, err := workspace.GetByPath(request.Workspace)
	if err != nil {
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: fmt.Sprintf("Failed to delete workspace: %v", err),
		})
		return
	}

	err = ws.Delete()
	if err != nil {
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: fmt.Sprintf("Failed to delete workspace: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(types.CommonResponse{
		Code:    0,
		Message: "Ok",
	})
}

func handleListWorkspace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(types.ListWorkspaceResponse{
		Code:    0,
		Message: "Ok",
		Data: types.Workspaces{
			Workspaces: workspace.GetAll(),
		},
	})
}
