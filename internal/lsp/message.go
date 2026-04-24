package lsp

type Request struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
}

type Response struct {
	JsonRPC string `json:"jsonrpc"`
	ID      *int   `json:"id,omitempty"`
}

type Notification struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
}
