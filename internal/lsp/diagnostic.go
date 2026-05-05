package lsp

type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           DiagnosticSeverity             `json:"severity"`
	Code               *Code                          `json:"code"`
	CodeDescription    *CodeDescription               `json:"codeDescription"`
	Source             string                         `json:"source"`
	Message            string                         `json:"message"`
	Tags               []DiagnosticTag                `json:"tags"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation"`
	Data               any                            `json:"data"`
}

type DiagnosticSeverity uint

const (
	Error DiagnosticSeverity = iota + 1
	Warning
	Information
	Hint
)

type Code int

const (
	UnusedIdentifier Code = iota
	UnexpectedToken
)

type CodeDescription struct {
	/**
	 * An URI to open with more information about the diagnostic error.
	 */
	Href string `json:"href"`
}

type DiagnosticTag uint

const (
	/**
	 * Unused or unnecessary code.
	 *
	 * Clients are allowed to render diagnostics with this tag faded out
	 * instead of having an error squiggle.
	 */
	Unnecessary DiagnosticTag = iota + 1
	/**
	 * Deprecated or obsolete code.
	 *
	 * Clients are allowed to rendered diagnostics with this tag strike through.
	 */
	Deprecated
)

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

type PublishDiagnosticsNotification struct {
	Notification
	Params PublishDiagnosticsParams `json:"params"`
}

type PublishDiagnosticsParams struct {
	/**
	 * The URI for which diagnostic information is reported.
	 */
	URI string `json:"uri"`
	/**
	 * Optional the version number of the document the diagnostics are published
	 * for.
	 *
	 * @since 3.15.0
	 */
	Version *int `json:"version"`
	/**
	 * An array of diagnostic information items.
	 */
	Diagnostics []Diagnostic `json:"diagnostics"`
}

func NewDiagnostic(srcRange Range, severity DiagnosticSeverity, source string, message string) *Diagnostic {
	return &Diagnostic{
		Range:    srcRange,
		Severity: severity,
		Source:   source,
		Message:  message,
	}
}

func NewPublishDiagnosticsNotification(diagnostics []Diagnostic, uri string, documentVersion int) *PublishDiagnosticsNotification {
	return &PublishDiagnosticsNotification{
		Notification: Notification{
			JsonRPC: "2.0",
			Method:  "textDocument/publishDiagnostics",
		},
		Params: PublishDiagnosticsParams{
			URI:         uri,
			Version:     &documentVersion,
			Diagnostics: diagnostics,
		},
	}
}
