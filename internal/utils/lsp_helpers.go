package utils

import (
	"github.com/caleb-fringer/imp_lsp/internal/lsp"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

// Converts a tree_sitter.Point struct to an lsp.Position struct
func PositionFromTS(p tree_sitter.Point) lsp.Position {
	return lsp.Position{
		Line:      p.Row,
		Character: p.Column,
	}
}

// Converts a tree_sitter.Range struct to an lsp.Range struct
func RangeFromTS(r tree_sitter.Range) lsp.Range {
	return lsp.Range{
		Start: PositionFromTS(r.StartPoint),
		End:   PositionFromTS(r.EndPoint),
	}
}
