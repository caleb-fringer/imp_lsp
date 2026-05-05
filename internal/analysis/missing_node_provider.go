package analysis

import (
	"errors"
	"fmt"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type missingNodeProvider struct {
	language       *tree_sitter.Language
	query          *tree_sitter.Query
	querySrc       string
	diagnosticCode lsp.Code
	captures       tree_sitter.QueryCaptures
}

func NewMissingNodeProvider(langauge *tree_sitter.Language) *missingNodeProvider {
	return &missingNodeProvider{
		language:       langauge,
		querySrc:       `(MISSING) @missing-node`,
		diagnosticCode: lsp.MissingNode,
	}
}

func (p *missingNodeProvider) GetName() string {
	return "Missing Node"
}

func (p *missingNodeProvider) Construct() error {
	query, err := tree_sitter.NewQuery(p.language, p.querySrc)
	if err != nil {
		return err
	}
	p.query = query
	return nil
}

func (p *missingNodeProvider) Execute(ctx AnalysisContext) error {
	cursor := ctx.QueryCursor
	root := ctx.Root
	documentSrc := ctx.DocumentSrc
	if cursor == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.QueryCursor is nil!")
	}
	if root == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.Root is nil!")
	}

	p.captures = cursor.Captures(p.query, root, documentSrc)
	return nil
}

func (p *missingNodeProvider) GetDiagnostics() []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)

	for match, idx := p.captures.Next(); match != nil; match, idx = p.captures.Next() {
		capture := match.Captures[idx]
		node := capture.Node
		expectedNode := node.Kind()

		diagnostic := lsp.Diagnostic{
			Range:    utils.RangeFromTS(node.Range()),
			Severity: lsp.Error,
			Code:     &p.diagnosticCode,
			Source:   DiagnosticSource,
			Message:  fmt.Sprintf("Syntax error: Expected %v", expectedNode),
		}
		diagnostics = append(diagnostics, diagnostic)
	}

	return diagnostics
}

func (p *missingNodeProvider) Close() {
	p.query.Close()
}
