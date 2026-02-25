package output

import "encoding/json"

type SuccessEnvelope struct {
	OK         bool            `json:"ok"`
	Command    string          `json:"command"`
	Args       map[string]any  `json:"args,omitempty"`
	Data       json.RawMessage `json:"data"`
	Pagination any             `json:"pagination,omitempty"`
}

type ErrorEnvelope struct {
	OK         bool   `json:"ok"`
	Command    string `json:"command"`
	Error      string `json:"error"`
	StatusCode int    `json:"status_code,omitempty"`
}
