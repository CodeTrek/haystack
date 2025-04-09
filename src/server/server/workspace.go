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

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	ws, _ := workspace.GetByPath(request.Workspace)
	if ws != nil {
		log.Printf("Create workspace `%s`: already exists", request.Workspace)
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    0,
			Message: "Workspace already exists",
		})
		return
	}

	ws, err = workspace.GetOrCreate(request.Workspace)
	if err != nil {
		log.Printf("Create workspace `%s`: failed to get or create: %v", request.Workspace, err)
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: fmt.Sprintf("Failed to create workspace: %v", err),
		})
		return
	}

	indexer.SyncIfNeeded(ws.Path)

	log.Printf("Created workspace `%s`", request.Workspace)
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
		log.Printf("Delete workspace `%s`: failed to get: %v", request.Workspace, err)
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: fmt.Sprintf("Failed to delete workspace: %v", err),
		})
		return
	}

	err = ws.Delete()
	if err != nil {
		log.Printf("Delete workspace `%s`: failed to delete: %v", request.Workspace, err)
		json.NewEncoder(w).Encode(types.CommonResponse{
			Code:    1,
			Message: fmt.Sprintf("Failed to delete workspace: %v", err),
		})
		return
	}

	log.Printf("Deleted workspace `%s`", request.Workspace)
	json.NewEncoder(w).Encode(types.CommonResponse{
		Code:    0,
		Message: "Ok",
	})
}

func handleListWorkspace(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}()

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

func handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	var request types.GetWorkspaceRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ws, err := workspace.GetByPath(request.Workspace)
	if err != nil {
		log.Printf("Get workspace `%s`: %v", request.Workspace, err)
		json.NewEncoder(w).Encode(types.GetWorkspaceResponse{
			Code:    1,
			Message: "Not found",
		})
		return
	}

	ws.Mutex.Lock()
	totalFiles := ws.TotalFiles
	indexing := ws.IndexingStatus != nil
	if totalFiles == 0 && indexing {
		totalFiles = ws.IndexingStatus.TotalFiles
	}
	ws.Mutex.Unlock()

	json.NewEncoder(w).Encode(types.GetWorkspaceResponse{
		Code:    0,
		Message: "Ok",
		Data: &types.Workspace{
			ID:           ws.ID,
			Path:         ws.Path,
			TotalFiles:   totalFiles,
			CreatedAt:    ws.CreatedAt,
			LastAccessed: ws.LastAccessed,
			LastFullSync: ws.LastFullSync,
			Indexing:     indexing,
		},
	})
}
