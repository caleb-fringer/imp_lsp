package analysis

import (
	"fmt"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

const DiagnosticSource = "imp_lsp"

// An abstraction over tree_sitter queries which generate diagnostics.
// A DiagnosticQuery may consist of many subqueries, but these are
// abstracted away to make collecting diagnostics easy.
type DiagnosticQuery interface {
	GetName() string
	Construct() error
	Execute(cursor *tree_sitter.QueryCursor,
		root *tree_sitter.Node,
		documentSrc []byte) error
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
func (q *unusedVariableQuery) Execute(cursor *tree_sitter.QueryCursor, root *tree_sitter.Node, documentSrc []byte) error {
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

		err = query.Execute(s.queryCursor, document.tree.RootNode(), []byte(document.Text))
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
