package utils

import (
	"fmt"

	tree_sitter "github.com/tree-sitter/go-tree-sitter"
)

func DFS(root *tree_sitter.Tree) {
	cursor := root.Walk()
	defer cursor.Close()

	done := false
	for !done {
		// Preorder: Print self first
		if cursor.Node().IsNamed() {
			fmt.Printf("%s ", cursor.Node().Kind())
		}

		// Recursively go to the leftmost child
		if cursor.GotoFirstChild() {
			continue
		}
		// No leftmost child, go to sibling

		// No right sibling. Go up the parents until there is a right sibling or you reach root.
		for !cursor.GotoNextSibling() {
			if !cursor.GotoParent() {
				done = true
				fmt.Println()
				break
			}
		}
	}
}
