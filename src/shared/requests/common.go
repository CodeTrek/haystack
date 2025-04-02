package requests

type CommonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
