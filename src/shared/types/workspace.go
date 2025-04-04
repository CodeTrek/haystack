package requests

type CreateWorkspaceRequest struct {
	Path string `json:"path"`
}

type DeleteWorkspaceRequest struct {
	Path string `json:"path"`
}

type ListWorkspaceResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Workspaces []string `json:"workspaces"`
	} `json:"data"`
}
