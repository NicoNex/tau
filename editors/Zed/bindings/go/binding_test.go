package tree_sitter_tau_test

import (
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
	"github.com/tree-sitter/tree-sitter-tau"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_tau.Language())
	if language == nil {
		t.Errorf("Error loading Tau grammar")
	}
}
