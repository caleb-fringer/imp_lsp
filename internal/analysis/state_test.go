package analysis

import (
	"testing"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	tree_sitter_imp "github.com/caleb-fringer/tree-sitter-imp/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

var testDocument = &lsp.TextDocumentItem{
	URI:        "file:///home/caleb/src/imp_lsp/declare_not_used.imp",
	LanguageID: "imp",
	Version:    1,
	Text: `x := 10
		   y := false
		   if y then
			3
		   else
		   	4
		   end`,
}

var changeNotification = lsp.NewDidChangeNotification(testDocument.URI, testDocument.Version+1, "z"+testDocument.Text[1:])

func TestNewState(t *testing.T) {
	state, err := NewState()
	if err != nil {
		t.Errorf("Failed to construct State: %v\n", err)
	}
	if state == nil {
		t.Error("Newly constructed State is nil.")
	}

	if state.documents == nil {
		t.Error("Newly constructed State has a nil documents map.")
	}

	if state.parser == nil {
		t.Error("Newly constructed State has a nil parser.")
	}
}

func TestOpenDocument(t *testing.T) {
	state, err := NewState()
	if err != nil {
		t.Errorf("Failed to construct State: %v\n", err)
	}
	defer state.Close()

	err = state.OpenDocument(testDocument)
	if err != nil {
		t.Fatalf("Failed to open document: %v\n", err)
	}

	tree := state.documents[uri(testDocument.URI)].tree
	if tree == nil {
		t.Fatal("Document has no parsed syntax tree in State.documents map.")
	}

	root := tree.RootNode()
	if root.Kind() != "source_file" {
		t.Errorf("Expected root node's kind to be source_file, got %v.\n", root.Kind())
	}
}

func TestEditDocument(t *testing.T) {
	state, err := NewState()
	if err != nil {
		t.Errorf("Failed to construct State: %v\n", err)
	}
	defer state.Close()

	err = state.OpenDocument(testDocument)
	if err != nil {
		t.Fatalf("Failed to open document: %v\n", err)
	}

	tree := state.documents[uri(testDocument.URI)].tree
	if tree == nil {
		t.Fatal("Document has no parsed syntax tree in State.documents map.")
	}

	declarations_query, queryErr := tree_sitter.NewQuery(
		tree_sitter.NewLanguage(tree_sitter_imp.Language()),
		`(assignment 
			id: (identifier) @id 
			val: [(integer) (boolean)] @val)`)

	if queryErr != nil {
		t.Fatalf("Error constructing variables query: %v", err)
	}
	defer declarations_query.Close()

	cursor := tree_sitter.NewQueryCursor()
	defer cursor.Close()

	query_matches := cursor.Matches(declarations_query, tree.RootNode(), []byte(testDocument.Text))

	assignment_node := query_matches.Next()
	id := assignment_node.Captures[0].Node.Utf8Text([]byte(testDocument.Text))
	if id != "x" {
		t.Errorf("Expected identifier to have value `x` before update, got %v\n", id)
	}

	// After this point, the previous tree pointers are invalid.
	if err := state.EditDocument(changeNotification); err != nil {
		t.Fatalf("Error editing document: %v", err)
	}
	new_tree := state.documents[uri(testDocument.URI)].tree

	updated_matches := cursor.Matches(declarations_query, new_tree.RootNode(), []byte(testDocument.Text))

	updated_assignment_node := updated_matches.Next()
	updated_id := updated_assignment_node.Captures[0].Node.Utf8Text([]byte(testDocument.Text))
	if updated_id != "z" {
		t.Errorf("Expected identifier to have value `z` after update, got %v\n", id)
	}
}
