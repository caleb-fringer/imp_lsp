package analysis

import (
	"errors"
	"fmt"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type syntaxError struct {
	node  *tree_sitter.Node
	token string
}
type unexpectedTokenProvider struct {
	language       *tree_sitter.Language
	query          *tree_sitter.Query
	querySrc       string
	diagnosticCode lsp.Code
	syntaxErrors   []syntaxError
}

func NewUnexpectedTokenProvider(language *tree_sitter.Language) *unexpectedTokenProvider {
	return &unexpectedTokenProvider{
		language:       language,
		querySrc:       `(ERROR) @error-node`,
		diagnosticCode: lsp.UnexpectedToken,
		syntaxErrors:   make([]syntaxError, 0),
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
	syntaxErrors := make([]syntaxError, 0)

	matches := cursor.Matches(p.query, root, documentSrc)
	for match := matches.Next(); match != nil; match = matches.Next() {
		errorNode := match.Captures[0].Node
		// Attempt to only capture the outermost error node.
		if errorNode.Parent().IsError() {
			continue
		}
		token := errorNode.Utf8Text(documentSrc)
		syntaxError := syntaxError{
			node:  &errorNode,
			token: token,
		}
		syntaxErrors = append(syntaxErrors, syntaxError)
	}
	// Need to overwrite to clear previous errors
	p.syntaxErrors = syntaxErrors
	return nil
}

func (p *unexpectedTokenProvider) GetDiagnostics() []lsp.Diagnostic {
	n := len(p.syntaxErrors)
	diagnostics := make([]lsp.Diagnostic, n)
	for i, e := range p.syntaxErrors {
		node := e.node
		token := e.token

		tokenRange := utils.RangeFromTS(node.Range())
		diagnostics[i] = lsp.Diagnostic{
			Range:    tokenRange,
			Severity: lsp.Error,
			Code:     &p.diagnosticCode,
			Source:   DiagnosticSource,
			Message:  fmt.Sprintf("Unexpected token `%s`", token),
		}
	}
	return diagnostics
}

func (p *unexpectedTokenProvider) Close() {
	p.query.Close()
}
