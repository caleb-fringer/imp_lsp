package main

import (
	"fmt"
	"os"

	"github.com/caleb-fringer/imp_lsp/internal/analysis"
	"github.com/caleb-fringer/imp_lsp/internal/utils"
	tree_sitter_imp "github.com/caleb-fringer/tree-sitter-imp/bindings/go"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func main() {
	file := "test.imp"
	data, err := os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading from %s: %v\n", file, err)
		os.Exit(1)
	}

	parser := tree_sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(tree_sitter.NewLanguage(tree_sitter_imp.Language()))

	tree := parser.Parse(data, nil)
	defer tree.Close()

	fmt.Println(tree.RootNode().ToSexp())
	utils.DFS(tree)

	file = "declare_not_used.imp"
	data, err = os.ReadFile(file)
	if err != nil {
		fmt.Printf("Error reading from %s: %v\n", file, err)
		os.Exit(1)
	}

	tree = parser.Parse(data, nil)
	fmt.Println(tree.RootNode().ToSexp())
	unused_variables, err := analysis.UnusedVariables(tree, data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for id := range unused_variables {
		fmt.Printf("Unused variable: %v\n", id)
	}
}
