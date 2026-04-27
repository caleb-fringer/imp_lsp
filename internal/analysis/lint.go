package analysis

import (
	"fmt"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	tree_sitter_imp "github.com/caleb-fringer/tree-sitter-imp/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Identifier string

func unusedVariables(root *tree_sitter.Tree, sourceCode []byte) (map[Identifier]tree_sitter.Node, error) {
	declarations_query, err := tree_sitter.NewQuery(
		tree_sitter.NewLanguage(tree_sitter_imp.Language()),
		`(assignment 
			id: (identifier) @id 
			val: [(integer) (boolean)] @val)`)

	if err != nil {
		return nil, fmt.Errorf("Error constructing variables query: %v", err)
	}
	defer declarations_query.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	query_matches := cursor.Matches(declarations_query, root.RootNode(), sourceCode)

	unused_identifiers := make(map[Identifier]tree_sitter.Node)
	for match := query_matches.Next(); match != nil; match = query_matches.Next() {
		id_node := match.Captures[0].Node
		id := id_node.Utf8Text(sourceCode)
		val := match.Captures[1].Node.Utf8Text(sourceCode)
		fmt.Printf("Identifer: %v, value: %v\n", id, val)
		unused_identifiers[Identifier(id)] = id_node
	}

	references_query, err := tree_sitter.NewQuery(
		tree_sitter.NewLanguage(tree_sitter_imp.Language()),
		`(expression/identifier) @id`)

	query_matches = cursor.Matches(references_query, root.RootNode(), sourceCode)
	for match := query_matches.Next(); match != nil; match = query_matches.Next() {
		id := match.Captures[0].Node.Utf8Text(sourceCode)
		delete(unused_identifiers, Identifier(id))
	}
	return unused_identifiers, nil
}

// An abstraction over tree_sitter queries which generate diagnostics.
// A diagnosticQuery may consist of many subqueries, but these are
// abstracted away to make collecting diagnostics easy.
type diagnosticQuery interface {
	getName() string
	construct() error
	execute(cursor *tree_sitter.QueryCursor) error
	getDiagnostics() []lsp.Diagnostic
	close()
}

func (s *ServerState) collectDiagnostics(document uri) ([]lsp.Diagnostic, error) {
	var result []lsp.Diagnostic
	for _, query := range s.diagnosticQueries {
		err := query.construct()
		if err != nil {
			s.logger.Printf("Error constructing query %v: %v\n", query.getName(), err)
			continue
		}
		// TODO: Maybe this should be handled by server.Close()?
		defer query.close()

		err = query.execute(s.queryCursor)
		if err != nil {
			s.logger.Printf("Error executing query %v: %v\n", query.getName(), err)
			continue
		}

		// If the query executed successfully, I think this should be fine to
		// not return an error
		diagnostics := query.getDiagnostics()
		result = append(result, diagnostics...)
	}
	return result, nil
}
