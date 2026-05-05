package analysis

import (
	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type AnalysisContext struct {
	QueryCursor *tree_sitter.QueryCursor
	Root        *tree_sitter.Node
	DocumentSrc []byte
}

const DiagnosticSource = "imp_lsp"

// An abstraction over tree_sitter queries/walkers which generate diagnostics.
// A DiagnosticsProvider may use an AST walker or a series of queries to
// create diagnostics, but the implementation is abstracted into an
// AnalysisContext.
type DiagnosticsProvider interface {
	GetName() string
	Construct() error
	Execute(ctx AnalysisContext) error
	GetDiagnostics() []lsp.Diagnostic
	Close()
}
