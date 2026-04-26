package lsp

type InitializeRequest struct {
	Request
	Params InitializeRequestParams `json:"params"`
}

type InitializeRequestParams struct {
	ClientInfo *ClientInfo `json:"clientInfo"`
}
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResponse struct {
	Response
	Result InitializeResult `json:"result"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync TextDocumentSyncKind `json:"textDocumentSync"`
}

type TextDocumentSyncKind int

// TextDocumentSyncKind values
const (
	None = iota
	Full
	Incremental
)

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewInitializeResponse(id int, sync TextDocumentSyncKind) InitializeResponse {
	if sync == Incremental {
		panic("Incremental syncing is not implemented!")
	}
	return InitializeResponse{
		Response: Response{
			JsonRPC: "2.0",
			ID:      &id,
		},
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync: sync,
			},
			ServerInfo: ServerInfo{
				Name:    "imp_lsp",
				Version: "v0.0.1",
			},
		},
	}
}
