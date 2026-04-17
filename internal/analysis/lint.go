package analysis

import (
	"fmt"

	tree_sitter_imp "github.com/caleb-fringer/tree-sitter-imp/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Identifier string

func UnusedVariables(root *tree_sitter.Tree, sourceCode []byte) (map[Identifier]tree_sitter.Node, error) {
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
