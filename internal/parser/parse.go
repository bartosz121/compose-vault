package parser

import (
	"github.com/goccy/go-yaml/ast"
)

type Match struct {
	Node        *ast.StringNode
	Placeholder Placeholder
}

type MatchError struct {
	Node *ast.StringNode
	Err  error
}

func FindMatches(file *ast.File) ([]Match, []MatchError) {
	visitor := &Visitor{}
	for _, doc := range file.Docs {
		ast.Walk(visitor, doc)
	}

	var matches []Match
	var matchErrors []MatchError

	for _, n := range visitor.matches {
		placeholder, err := parseMatch(n)
		if err != nil {
			matchErrors = append(matchErrors, MatchError{Node: n, Err: err})
			continue
		}

		matches = append(matches, Match{Node: n, Placeholder: placeholder})
	}

	return matches, matchErrors
}

func parseMatch(node *ast.StringNode) (Placeholder, error) {
	placeholder, err := ParsePlaceholder(node.Value)
	if err != nil {
		return Placeholder{}, err
	}

	return placeholder, nil
}
