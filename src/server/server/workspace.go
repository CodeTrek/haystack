package server

import (
	"encoding/json"
	"fmt"
	"haystack/server/core/workspace"
	"haystack/shared/types"
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

	ws, _ := workspace.GetByPath(request.Path)
	if ws != nil {
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: "Workspace already exists",
		})
		return
	}

	_, err = workspace.GetOrCreate(request.Path)
	if err != nil {
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: fmt.Sprintf("Failed to create workspace: %v", err),
		})
		return
	}

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

	ws, err := workspace.GetByPath(request.Path)
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
	w.WriteHeader(http.StatusOK)

	workspaces := workspace.GetAll()
	json.NewEncoder(w).Encode(types.ListWorkspaceResponse{
		Code:    0,
		Message: "Ok",
		Data: struct {
			Workspaces []string `json:"workspaces"`
		}{
			Workspaces: workspaces,
		},
	})
}
