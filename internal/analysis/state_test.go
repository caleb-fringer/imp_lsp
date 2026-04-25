package analysis

import (
	"testing"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
)

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

	document := &lsp.TextDocumentItem{
		URI:        "file:///home/caleb/src/imp_lsp/declare_not_used.imp",
		LanguageID: "imp",
		Version:    1,
		Text: `
			x := 10
			y := false
			if y then
				3
			else
				4
			end
		`,
	}

	err = state.OpenDocument(document)
	if err != nil {
		t.Fatalf("Failed to open document: %v\n", err)
	}

	tree := state.documents[document]
	if tree == nil {
		t.Fatal("Document has no parsed syntax tree in State.documents map.")
	}

	root := tree.RootNode()
	if root.Kind() != "source_file" {
		t.Errorf("Expected root node's kind to be source_file, got %v.\n", root.Kind())
	}
}
