package analysis

import (
	"fmt"
	"strings"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	tree_sitter_imp "github.com/caleb-fringer/tree-sitter-imp/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type uri string
type document struct {
	*lsp.TextDocumentItem
	tree *tree_sitter.Tree
}

type ServerState struct {
	documents map[uri]*document
	parser    *tree_sitter.Parser
}

// Create a new State object
func NewState() (*ServerState, error) {
	parser := tree_sitter.NewParser()
	language := tree_sitter.NewLanguage(tree_sitter_imp.Language())
	err := parser.SetLanguage(language)
	if err != nil {
		return nil, fmt.Errorf("Failed to set parser's language: %v", err)
	}

	return &ServerState{
		documents: make(map[uri]*document),
		parser:    parser,
	}, nil
}

// Close open tree_sitter.Tree's and tree_sitter.Parser
func (s *ServerState) Close() {
	for _, doc := range s.documents {
		doc.tree.Close()
	}
	s.parser.Close()
}

// Add the provided TextDocumentItem to the state of opened documents, and
// parse its contents.
// Returns an error if the the document has already been opened.
func (s *ServerState) OpenDocument(textDocItem *lsp.TextDocumentItem) error {
	// Strip off the protocol prefix
	filePath, found := strings.CutPrefix(textDocItem.URI, "file://")
	if !found {
		filePath = textDocItem.URI
	}

	if _, exists := s.documents[uri(textDocItem.URI)]; exists {
		return fmt.Errorf("Document %s has already been opened. Clients should only open a document once.\n", filePath)
	}

	// Parse the document and add the CST to the state.
	s.documents[uri(textDocItem.URI)] = &document{
		textDocItem,
		s.parser.Parse([]byte(textDocItem.Text), nil),
	}
	return nil
}
