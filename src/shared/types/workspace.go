package types

import (
	"time"
)

type CreateWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}

type DeleteWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}

type Workspace struct {
	ID   string `json:"id"`
	Path string `json:"path"`

	CreatedAt    time.Time `json:"created_time"`
	LastAccessed time.Time `json:"last_accessed_time"`
	LastFullSync time.Time `json:"last_full_sync_time"`
}

type Workspaces struct {
	Workspaces []Workspace `json:"workspaces"`
}

type ListWorkspaceResponse struct {
	Code    int        `json:"code"`
	Message string     `json:"message"`
	Data    Workspaces `json:"data"`
}
