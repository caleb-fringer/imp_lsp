package analysis

import (
	"errors"
	"fmt"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
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
// created diagnostics, but the implementation is abstracted into an
// AnalysisContext.
type DiagnosticsProvider interface {
	GetName() string
	Construct() error
	Execute(ctx AnalysisContext) error
	GetDiagnostics() []lsp.Diagnostic
	Close()
}

type Identifier string

type unusedVariableQuery struct {
	language             *tree_sitter.Language
	declarationsQuery    *tree_sitter.Query
	usageQuery           *tree_sitter.Query
	declarationsQuerySrc string
	usageQuerySrc        string
	unusedIdentifiers    map[Identifier]tree_sitter.Node
	diagnosticCode       lsp.Code
}

func NewUnusedVariableQuery(language *tree_sitter.Language) *unusedVariableQuery {
	return &unusedVariableQuery{
		language: language,
		declarationsQuerySrc: `(assignment 
			id: (identifier) @id 
			val: [(integer) (boolean)] @val)`,
		usageQuerySrc:  `(expression/identifier) @id`,
		diagnosticCode: lsp.UnusedIdentifier,
	}
}

func (q *unusedVariableQuery) GetName() string {
	return "Unused Variable"
}

func (q *unusedVariableQuery) Construct() error {
	declarationsQuery, err := tree_sitter.NewQuery(q.language, q.declarationsQuerySrc)
	if err != nil {
		return err
	}
	q.declarationsQuery = declarationsQuery

	usageQuery, err := tree_sitter.NewQuery(q.language, q.usageQuerySrc)
	if err != nil {
		return err
	}
	q.usageQuery = usageQuery
	return nil
}

// Recomputes the unused identifiers, replacing the previous map with the new one.
func (q *unusedVariableQuery) Execute(ctx AnalysisContext) error {
	cursor := ctx.QueryCursor
	root := ctx.Root
	documentSrc := ctx.DocumentSrc
	if cursor == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.QueryCursor is nil!")
	}
	if root == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.Root is nil!")
	}

	declarationsQuery_matches := cursor.Matches(q.declarationsQuery, root, documentSrc)

	unusedIdentifiers := make(map[Identifier]tree_sitter.Node)
	for match := declarationsQuery_matches.Next(); match != nil; match = declarationsQuery_matches.Next() {
		id_node := match.Captures[0].Node
		id := id_node.Utf8Text(documentSrc)
		val := match.Captures[1].Node.Utf8Text(documentSrc)
		fmt.Printf("Identifer: %v, value: %v\n", id, val)
		unusedIdentifiers[Identifier(id)] = id_node
	}

	// Iterate over usages, deleting them from the unsued_identifiers map.
	usageQuery_matches := cursor.Matches(q.usageQuery, root, documentSrc)
	for match := usageQuery_matches.Next(); match != nil; match = usageQuery_matches.Next() {
		id := match.Captures[0].Node.Utf8Text(documentSrc)
		delete(unusedIdentifiers, Identifier(id))
	}

	q.unusedIdentifiers = unusedIdentifiers
	return nil
}

func (q *unusedVariableQuery) GetDiagnostics() []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, len(q.unusedIdentifiers))
	i := 0
	for identifier, node := range q.unusedIdentifiers {
		diagnostics[i] = lsp.Diagnostic{
			Range:    utils.RangeFromTS(node.Range()),
			Severity: lsp.Warning,
			Code:     &q.diagnosticCode,
			Source:   DiagnosticSource,
			Message:  fmt.Sprintf("%s is declared but not used", identifier),
		}
		i++
	}
	return diagnostics
}

func (q *unusedVariableQuery) Close() {
	q.declarationsQuery.Close()
	q.usageQuery.Close()
}

type unexpectedTokenQuery struct {
	language       *tree_sitter.Language
	query          *tree_sitter.Query
	querySrc       string
	diagnosticCode lsp.Code
	errorLocations []tree_sitter.Range
}

func NewUnexpectedTokenQuery(language *tree_sitter.Language) *unexpectedTokenQuery {
	return &unexpectedTokenQuery{
		language:       language,
		querySrc:       `(ERROR) @error-node`,
		diagnosticCode: lsp.UnexpectedToken,
		errorLocations: make([]tree_sitter.Range, 0),
	}
}

func (q *unexpectedTokenQuery) GetName() string {
	return "Unexpected Token"
}

func (q *unexpectedTokenQuery) Construct() error {
	query, err := tree_sitter.NewQuery(q.language, q.querySrc)
	if err != nil {
		return nil
	}
	q.query = query
	return nil
}

func (q *unexpectedTokenQuery) Execute(ctx AnalysisContext) error {
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

	matches := cursor.Matches(q.query, root, documentSrc)
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
	q.errorLocations = errorLocations
	return nil
}

func (q *unexpectedTokenQuery) GetDiagnostics() []lsp.Diagnostic {
	n := len(q.errorLocations)
	diagnostics := make([]lsp.Diagnostic, n)
	for i, loc := range q.errorLocations {
		tokenRange := utils.RangeFromTS(loc)
		diagnostics[i] = lsp.Diagnostic{
			Range:    tokenRange,
			Severity: lsp.Error,
			Code:     &q.diagnosticCode,
			Source:   DiagnosticSource,
			Message:  "Unexpected token",
		}
	}
	return diagnostics
}

func (q *unexpectedTokenQuery) Close() {
	q.query.Close()
}

func (s *ServerState) collectDiagnostics(documentUri uri) ([]lsp.Diagnostic, error) {
	document := s.documents[documentUri]
	result := make([]lsp.Diagnostic, 0)
	for _, query := range s.diagnosticQueries {
		err := query.Construct()
		if err != nil {
			s.logger.Printf("Error constructing query %v: %v\n", query.GetName(), err)
			continue
		}
		// TODO: Maybe this should be handled by server.Close()?
		defer query.Close()

		ctx := AnalysisContext{
			QueryCursor: s.queryCursor,
			Root:        document.tree.RootNode(),
			DocumentSrc: []byte(document.Text),
		}
		err = query.Execute(ctx)
		if err != nil {
			s.logger.Printf("Error executing query %v: %v\n", query.GetName(), err)
			continue
		}

		// If the query executed successfully, I think this should be fine to
		// not return an error
		diagnostics := query.GetDiagnostics()
		result = append(result, diagnostics...)
	}
	return result, nil
}
