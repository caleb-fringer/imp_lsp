package analysis

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	tree_sitter_imp "github.com/caleb-fringer/tree-sitter-imp/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type uri string
type document struct {
	// TODO: Probably should refactor this to not depend on lsp.TextDocumentItem
	*lsp.TextDocumentItem
	tree *tree_sitter.Tree
}

type ServerState struct {
	documents map[uri]*document
	parser    *tree_sitter.Parser
	logger    *log.Logger
}

// Create a new State object
func NewState(logger *log.Logger) (*ServerState, error) {
	parser := tree_sitter.NewParser()
	language := tree_sitter.NewLanguage(tree_sitter_imp.Language())
	err := parser.SetLanguage(language)
	if err != nil {
		return nil, fmt.Errorf("Failed to set parser's language: %v", err)
	}

	return &ServerState{
		documents: make(map[uri]*document),
		parser:    parser,
		logger:    logger,
	}, nil
}

// Close all tree_sitter & other resources.
func (s *ServerState) Close() {
	for _, doc := range s.documents {
		doc.tree.Close()
	}
	s.parser.Close()
}

// TODO: Probably should refactor this to not depend on lsp.TextDocumentItem
// Add the provided TextDocumentItem to the state of opened documents, and
// parse its contents. Returns diagnostics resulting from running the linter
// on the newly parsed CST.
// Returns an error if the the document has already been opened.
func (s *ServerState) OpenDocument(textDocItem *lsp.TextDocumentItem) (diagnostics []lsp.Diagnostic, err error) {
	// Strip off the protocol prefix
	filePath, found := strings.CutPrefix(textDocItem.URI, "file://")
	if !found {
		filePath = textDocItem.URI
	}

	if _, exists := s.documents[uri(textDocItem.URI)]; exists {
		return nil, fmt.Errorf("Document %s has already been opened. Clients should only open a document once.\n", filePath)
	}

	// Parse the document and add the CST to the state.
	s.documents[uri(textDocItem.URI)] = &document{
		textDocItem,
		s.parser.Parse([]byte(textDocItem.Text), nil),
	}
	return diagnostics, nil
}

// WARNING: Incremental document synchronization is not yet supported, although
// the following method documentation indicates that they are. This documentation
// reflects the future API which will support incremental updates.
//
// TODO: Refactor to not edepend on lsp.DidChangeTextDocumentNotification
//
// Handle a DidChangeTextDocumentNotification by editing the associated
// document's Tree and re-parsing it. Returns diagnostics resulting from
// running the linter on the newly parsed CST.
// Returns an error when:
//   - The document is not in the ServerState (has not been opened).
//   - The changeEvent's version is behind the server's version.
//   - The changeEvent's version is not immediately succeeding the server's
//     version, and incremental syncing has been selected (missing intermediate
//     updates).
//   - The new changes could not be parsed.
func (s *ServerState) EditDocument(changeEvent *lsp.DidChangeTextDocumentNotification) (diagnostics []lsp.Diagnostic, err error) {
	eventUri := changeEvent.Params.TextDocument.URI
	eventVersion := changeEvent.Params.TextDocument.Version
	doc, exists := s.documents[uri(eventUri)]
	if !exists {
		return nil, fmt.Errorf("Document %s was not found; are you sure you opened it?\n", eventUri)
	}

	// TODO: Handle the case where we use incremental synchronization and the
	// changeEvent's version is more than 1 + the server's version (missing
	// intermediate updates).
	if eventVersion <= doc.Version {
		return nil, fmt.Errorf("Document %s provided is older than the ServerState's current version:\n"+
			"\tProvided version: %v\n"+
			"\tcurrent version: %v\n",
			eventUri,
			eventVersion,
			doc.Version,
		)
	}

	changes := changeEvent.Params.ContentChanges
	for _, change := range changes {
		// Range field is only present when incremental syncing is used.
		if change.Range != nil {
			return nil, errors.New("Incremental syncing is not implemented.")
		}
		// Delete and overwrite the previous tree w/ the new document.
		doc.Version = eventVersion
		doc.Text = change.Text
		doc.tree.Close()
		doc.tree = s.parser.Parse([]byte(change.Text), nil)
	}

	return diagnostics, nil
}
