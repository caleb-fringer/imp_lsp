package analysis

import (
	"errors"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type unexpectedTokenProvider struct {
	language       *tree_sitter.Language
	query          *tree_sitter.Query
	querySrc       string
	diagnosticCode lsp.Code
	errorLocations []tree_sitter.Range
}

func NewUnexpectedTokenProvider(language *tree_sitter.Language) *unexpectedTokenProvider {
	return &unexpectedTokenProvider{
		language:       language,
		querySrc:       `(ERROR) @error-node`,
		diagnosticCode: lsp.UnexpectedToken,
		errorLocations: make([]tree_sitter.Range, 0),
	}
}

func (p *unexpectedTokenProvider) GetName() string {
	return "Unexpected Token"
}

func (p *unexpectedTokenProvider) Construct() error {
	query, err := tree_sitter.NewQuery(p.language, p.querySrc)
	if err != nil {
		return nil
	}
	p.query = query
	return nil
}

func (p *unexpectedTokenProvider) Execute(ctx AnalysisContext) error {
	cursor := ctx.QueryCursor
	root := ctx.Root
	documentSrc := ctx.DocumentSrc
	if cursor == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.QueryCursor is nil!")
	}
	if root == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.Root is nil!")
	}

	// Local storage for errors
	errorLocations := make([]tree_sitter.Range, 0)

	matches := cursor.Matches(p.query, root, documentSrc)
	for match := matches.Next(); match != nil; match = matches.Next() {
		errorNode := match.Captures[0].Node
		// Attempt to only capture the outermost error node.
		if errorNode.Parent().IsError() {
			continue
		}
		errorRange := errorNode.Range()
		errorLocations = append(errorLocations, errorRange)
	}
	// Need to overwrite to clear previous errors
	p.errorLocations = errorLocations
	return nil
}

func (p *unexpectedTokenProvider) GetDiagnostics() []lsp.Diagnostic {
	n := len(p.errorLocations)
	diagnostics := make([]lsp.Diagnostic, n)
	for i, loc := range p.errorLocations {
		tokenRange := utils.RangeFromTS(loc)
		diagnostics[i] = lsp.Diagnostic{
			Range:    tokenRange,
			Severity: lsp.Error,
			Code:     &p.diagnosticCode,
			Source:   DiagnosticSource,
			Message:  "Unexpected token",
		}
	}
	return diagnostics
}

func (p *unexpectedTokenProvider) Close() {
	p.query.Close()
}
