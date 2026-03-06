package parser

import (
	"github.com/goccy/go-yaml/ast"
)

type Visitor struct {
	matches []*ast.StringNode
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}

	stringNode, ok := node.(*ast.StringNode)
	if !ok {
		return v
	}

	ok = IsPlaceholderCandidate(stringNode.Value)
	if !ok {
		return v
	}

	v.matches = append(v.matches, stringNode)

	return v
}
