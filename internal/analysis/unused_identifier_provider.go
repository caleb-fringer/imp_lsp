package analysis

import (
	"errors"
	"fmt"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Identifier string

type unusedIdentifierProvider struct {
	language             *tree_sitter.Language
	declarationsQuery    *tree_sitter.Query
	usageQuery           *tree_sitter.Query
	declarationsQuerySrc string
	usageQuerySrc        string
	unusedIdentifiers    map[Identifier]tree_sitter.Node
	diagnosticCode       lsp.Code
}

func NewUnusedIdentifierProvider(language *tree_sitter.Language) *unusedIdentifierProvider {
	return &unusedIdentifierProvider{
		language: language,
		declarationsQuerySrc: `(assignment 
			id: (identifier) @id 
			val: [(integer) (boolean)] @val)`,
		usageQuerySrc:  `(expression/identifier) @id`,
		diagnosticCode: lsp.UnusedIdentifier,
	}
}

func (p *unusedIdentifierProvider) GetName() string {
	return "Unused Variable"
}

func (p *unusedIdentifierProvider) Construct() error {
	declarationsQuery, err := tree_sitter.NewQuery(p.language, p.declarationsQuerySrc)
	if err != nil {
		return err
	}
	p.declarationsQuery = declarationsQuery

	usageQuery, err := tree_sitter.NewQuery(p.language, p.usageQuerySrc)
	if err != nil {
		return err
	}
	p.usageQuery = usageQuery
	return nil
}

// Recomputes the unused identifiers, replacing the previous map with the new one.
func (p *unusedIdentifierProvider) Execute(ctx AnalysisContext) error {
	cursor := ctx.QueryCursor
	root := ctx.Root
	documentSrc := ctx.DocumentSrc
	if cursor == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.QueryCursor is nil!")
	}
	if root == nil {
		return errors.New("Error executing DiagnosticQuery: provided ctx.Root is nil!")
	}

	declarationsQuery_matches := cursor.Matches(p.declarationsQuery, root, documentSrc)

	unusedIdentifiers := make(map[Identifier]tree_sitter.Node)
	for match := declarationsQuery_matches.Next(); match != nil; match = declarationsQuery_matches.Next() {
		id_node := match.Captures[0].Node
		id := id_node.Utf8Text(documentSrc)
		val := match.Captures[1].Node.Utf8Text(documentSrc)
		fmt.Printf("Identifer: %v, value: %v\n", id, val)
		unusedIdentifiers[Identifier(id)] = id_node
	}

	// Iterate over usages, deleting them from the unsued_identifiers map.
	usageQuery_matches := cursor.Matches(p.usageQuery, root, documentSrc)
	for match := usageQuery_matches.Next(); match != nil; match = usageQuery_matches.Next() {
		id := match.Captures[0].Node.Utf8Text(documentSrc)
		delete(unusedIdentifiers, Identifier(id))
	}

	p.unusedIdentifiers = unusedIdentifiers
	return nil
}

func (p *unusedIdentifierProvider) GetDiagnostics() []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, len(p.unusedIdentifiers))
	i := 0
	for identifier, node := range p.unusedIdentifiers {
		diagnostics[i] = lsp.Diagnostic{
			Range:    utils.RangeFromTS(node.Range()),
			Severity: lsp.Warning,
			Code:     &p.diagnosticCode,
			Source:   DiagnosticSource,
			Message:  fmt.Sprintf("%s is declared but not used", identifier),
		}
		i++
	}
	return diagnostics
}

func (p *unusedIdentifierProvider) Close() {
	p.declarationsQuery.Close()
	p.usageQuery.Close()
}
