package types

import (
	"time"
)

type Workspace struct {
	ID         string `json:"id"`
	Path       string `json:"path"`
	TotalFiles int    `json:"total_files"`

	CreatedAt    time.Time `json:"created_time"`
	LastAccessed time.Time `json:"last_accessed_time"`
	LastFullSync time.Time `json:"last_full_sync_time"`
	Indexing     bool      `json:"indexing"`
}

type Workspaces struct {
	Workspaces []Workspace `json:"workspaces"`
}

type CreateWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}

type CreateWorkspaceResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    Workspace `json:"data"`
}

type DeleteWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}

type DeleteWorkspaceResponse struct {
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    Workspace `json:"data"`
}

type ListWorkspaceResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    Workspaces `json:"data"`
}

type GetWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}

type GetWorkspaceResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    *Workspace `json:"data,omitempty"`
}

type SyncWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}
