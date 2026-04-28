package lsp

type InitializeRequest struct {
	Request
	Params InitializeRequestParams `json:"params"`
}

type InitializeRequestParams struct {
	ClientInfo         *ClientInfo        `json:"clientInfo,omitempty"`
	ClientCapabilities ClientCapabilities `json:"capabilities"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ClientCapabilities struct {
	/**
	 * General client capabilities.
	 *
	 * @since 3.16.0
	 */
	General *General `json:"general,omitempty"`
}

type General struct {
	/**
	 * The position encodings supported by the client. Client and server
	 * have to agree on the same position encoding to ensure that offsets
	 * (e.g. character position in a line) are interpreted the same on both
	 * side.
	 *
	 * To keep the protocol backwards compatible the following applies: if
	 * the value 'utf-16' is missing from the array of position encodings
	 * servers can assume that the client supports UTF-16. UTF-16 is
	 * therefore a mandatory encoding.
	 *
	 * If omitted it defaults to ['utf-16'].
	 *
	 * Implementation considerations: since the conversion from one encoding
	 * into another requires the content of the file / line the conversion
	 * is best done where the file is read which is usually on the server
	 * side.
	 *
	 * @since 3.17.0
	 */
	PositionEncodings []PositionEncodingKind `json:"positionEncodings,omitempty"`
}

type PositionEncodingKind string

const (
	UTF8  PositionEncodingKind = "utf-8"
	UTF16 PositionEncodingKind = "utf-16"
	UTF32 PositionEncodingKind = "utf-32"
)

type InitializeResponse struct {
	Response
	Result InitializeResult `json:"result"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	PositionEncodingKind *PositionEncodingKind `json:"positionEncoding,omitempty"`
	TextDocumentSync     *TextDocumentSyncKind `json:"textDocumentSync,omitempty"`
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

	// Hacky thing to allow me to take a pointer of a constant.
	encoding := UTF8

	return InitializeResponse{
		Response: Response{
			JsonRPC: "2.0",
			ID:      &id,
		},
		Result: InitializeResult{
			Capabilities: ServerCapabilities{
				PositionEncodingKind: &encoding,
				TextDocumentSync:     &sync,
			},
			ServerInfo: ServerInfo{
				Name:    "imp_lsp",
				Version: "v0.0.1",
			},
		},
	}
}
