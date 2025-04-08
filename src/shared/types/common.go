package types

import "encoding/json"

type CommonResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    *json.RawMessage `json:"data,omitempty"`
}

type ServerStatus struct {
	ShuttingDown bool   `json:"shutting_down"`
	Restarting   bool   `json:"restarting"`
	PID          int    `json:"pid"`
	Version      string `json:"version"`
}
