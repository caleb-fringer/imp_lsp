package analysis

import (
	"errors"
	"fmt"

	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

type Identifier string

type identifierDiagnosticsProvider struct {
	language                 *tree_sitter.Language
	declarationsQuery        *tree_sitter.Query
	usageQuery               *tree_sitter.Query
	declarationsQuerySrc     string
	usageQuerySrc            string
	declaredIdentifiers      map[Identifier]tree_sitter.Node
	usedIdentifiers          map[Identifier]tree_sitter.Node
	unusedIdentifierCode     lsp.Code
	undeclaredIdentifierCode lsp.Code
}

func NewIdentifierDiagnosticsProvider(
	language *tree_sitter.Language,
) *identifierDiagnosticsProvider {
	return &identifierDiagnosticsProvider{
		language: language,
		declarationsQuerySrc: `(assignment 
			id: (identifier) @id 
			val: [(integer) (boolean)] @val)`,
		usageQuerySrc:            `(expression/identifier) @id`,
		unusedIdentifierCode:     lsp.UnusedIdentifier,
		undeclaredIdentifierCode: lsp.UndeclaredIdentifier,
	}
}

func (p *identifierDiagnosticsProvider) GetName() string {
	return "Unused Variable"
}

func (p *identifierDiagnosticsProvider) Construct() error {
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
func (p *identifierDiagnosticsProvider) Execute(ctx AnalysisContext) error {
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

	declaredIdentifiers := make(map[Identifier]tree_sitter.Node)
	for match := declarationsQuery_matches.Next(); match != nil; match = declarationsQuery_matches.Next() {
		id_node := match.Captures[0].Node
		id := id_node.Utf8Text(documentSrc)
		val := match.Captures[1].Node.Utf8Text(documentSrc)
		fmt.Printf("Identifer: %v, value: %v\n", id, val)
		declaredIdentifiers[Identifier(id)] = id_node
	}

	// Iterate over usages, deleting them from the unused_identifiers map.
	usedIdentifiers := make(map[Identifier]tree_sitter.Node)
	usageQuery_matches := cursor.Matches(p.usageQuery, root, documentSrc)
	for match := usageQuery_matches.Next(); match != nil; match = usageQuery_matches.Next() {
		node := match.Captures[0].Node
		id := node.Utf8Text(documentSrc)
		usedIdentifiers[Identifier(id)] = node
	}

	p.declaredIdentifiers = declaredIdentifiers
	p.usedIdentifiers = usedIdentifiers
	return nil
}

func (p *identifierDiagnosticsProvider) GetDiagnostics() []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)
	// Remove identifiers in the intersection of declaredIdentifiers and
	// usedIdentifiers from BOTH maps.
	for identifier, node := range p.declaredIdentifiers {
		_, found := p.usedIdentifiers[identifier]
		// Delete from both maps
		if found {
			delete(p.declaredIdentifiers, identifier)
			delete(p.usedIdentifiers, identifier)
			continue
		}
		// Construct a diagnostic for the unused identifier
		unusedIdentifierDiagnostic := lsp.Diagnostic{
			Range:    utils.RangeFromTS(node.Range()),
			Severity: lsp.Warning,
			Code:     &p.unusedIdentifierCode,
			Source:   DiagnosticSource,
			Message:  fmt.Sprintf("%s is declared but not used", identifier),
		}
		diagnostics = append(diagnostics, unusedIdentifierDiagnostic)
	}

	// Now, iterate over the remaining usedIdentifiers to find identifiers that
	// have been used without being declared
	for identifier, node := range p.usedIdentifiers {
		undeclaredIdentifierDiagnostic := lsp.Diagnostic{
			Range:    utils.RangeFromTS(node.Range()),
			Severity: lsp.Error,
			Code:     &p.undeclaredIdentifierCode,
			Source:   DiagnosticSource,
			Message:  fmt.Sprintf("%s has not been declared", identifier),
		}
		diagnostics = append(diagnostics, undeclaredIdentifierDiagnostic)
	}
	return diagnostics
}

func (p *identifierDiagnosticsProvider) Close() {
	p.declarationsQuery.Close()
	p.usageQuery.Close()
}
