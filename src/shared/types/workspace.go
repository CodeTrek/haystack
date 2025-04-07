package types

type CreateWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}

type DeleteWorkspaceRequest struct {
	Workspace string `json:"workspace"`
}

type ListWorkspaceResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Workspaces []string `json:"workspaces"`
	} `json:"data"`
}
