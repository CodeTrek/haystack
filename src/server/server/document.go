package server

import (
	"encoding/json"
	"haystack/server/core/workspace"
	"haystack/server/indexer"
	"haystack/shared/types"
	"log"
	"net/http"
	"path/filepath"
)

func handleUpdateDocument(w http.ResponseWriter, r *http.Request) {
	var request types.DocumentUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	workspace, err := workspace.GetByPath(request.Workspace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	indexer.AddOrSyncFile(workspace, filepath.Join(workspace.Path, request.Path))

	log.Printf("Updated document %s in workspace %s", request.Path, request.Workspace)

	json.NewEncoder(w).Encode(types.CommonResponse{
		Code:    0,
		Message: "Ok",
	})
}

func handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	var request types.DocumentDeleteRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	workspace, err := workspace.GetByPath(request.Workspace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	indexer.RemoveFile(workspace, filepath.Join(workspace.Path, request.Path))
	log.Printf("Deleted document %s in workspace %s", request.Path, request.Workspace)

	json.NewEncoder(w).Encode(types.CommonResponse{
		Code:    0,
		Message: "Ok",
	})
}
