package gogrep

import (
	"go/ast"
	"go/token"
	"go/types"
)

// This is an ugly way to use gogrep as a library.
// It can go away when there will be another option.

// Parse creates a gogrep pattern out of a given string expression.
func Parse(fset *token.FileSet, expr string) (*Pattern, error) {
	m := matcher{
		fset: fset,
		Info: &types.Info{},
	}
	node, err := m.parseExpr(expr)
	if err != nil {
		return nil, err
	}
	return &Pattern{m: &m, Expr: node}, nil
}

// Pattern is a compiled gogrep pattern.
type Pattern struct {
	Expr ast.Node
	m    *matcher
}

// MatchData describes a successful pattern match.
type MatchData struct {
	Node   ast.Node
	Values map[string]ast.Node
}

// MatchNode calls cb if n matches a pattern.
func (p *Pattern) MatchNode(n ast.Node, cb func(MatchData)) {
	p.m.values = map[string]ast.Node{}
	if p.m.node(p.Expr, n) {
		cb(MatchData{
			Values: p.m.values,
			Node:   n,
		})
	}
}

// Match calls cb for any pattern match found in n.
func (p *Pattern) Match(n ast.Node, cb func(MatchData)) {
	cmd := exprCmd{name: "x", value: p.Expr}
	matches := p.m.cmdRange(cmd, []submatch{{
		values: map[string]ast.Node{},
		node:   n,
	}})
	for _, match := range matches {
		cb(MatchData{
			Values: match.values,
			Node:   match.node,
		})
	}
}
