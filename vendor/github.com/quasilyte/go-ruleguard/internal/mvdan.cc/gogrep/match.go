// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package gogrep

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
)

func (m *matcher) matches(cmds []exprCmd, nodes []ast.Node) []ast.Node {
	m.parents = make(map[ast.Node]ast.Node)
	m.fillParents(nodes...)
	initial := make([]submatch, len(nodes))
	for i, node := range nodes {
		initial[i].node = node
		initial[i].values = make(map[string]ast.Node)
	}
	final := m.submatches(cmds, initial)
	finalNodes := make([]ast.Node, len(final))
	for i := range finalNodes {
		finalNodes[i] = final[i].node
	}
	return finalNodes
}

func (m *matcher) fillParents(nodes ...ast.Node) {
	stack := make([]ast.Node, 1, 32)
	for _, node := range nodes {
		inspect(node, func(node ast.Node) bool {
			if node == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			if _, ok := node.(nodeList); !ok {
				m.parents[node] = stack[len(stack)-1]
			}
			stack = append(stack, node)
			return true
		})
	}
}

type submatch struct {
	node   ast.Node
	values map[string]ast.Node
}

func valsCopy(values map[string]ast.Node) map[string]ast.Node {
	v2 := make(map[string]ast.Node, len(values))
	for k, v := range values {
		v2[k] = v
	}
	return v2
}

func (m *matcher) submatches(cmds []exprCmd, subs []submatch) []submatch {
	if len(cmds) == 0 {
		return subs
	}
	cmd := cmds[0]
	var fn func(exprCmd, []submatch) []submatch
	switch cmd.name {
	case "x":
		fn = m.cmdRange
	case "g":
		fn = m.cmdFilter(true)
	case "v":
		fn = m.cmdFilter(false)
	case "s":
		fn = m.cmdSubst
	case "a":
		fn = m.cmdAttr
	case "p":
		fn = m.cmdParents
	case "w":
		if len(cmds) > 1 {
			panic("-w must be the last command")
		}
		fn = m.cmdWrite
	default:
		panic(fmt.Sprintf("unknown command: %q", cmd.name))
	}
	return m.submatches(cmds[1:], fn(cmd, subs))
}

func (m *matcher) cmdRange(cmd exprCmd, subs []submatch) []submatch {
	var matches []submatch
	seen := map[nodePosHash]bool{}

	// The values context for each new submatch must be a new copy
	// from its parent submatch. If we don't do this copy, all the
	// submatches would share the same map and have side effects.
	var startValues map[string]ast.Node

	match := func(exprNode, node ast.Node) {
		if node == nil {
			return
		}
		m.values = valsCopy(startValues)
		found := m.topNode(exprNode, node)
		if found == nil {
			return
		}
		hash := posHash(found)
		if !seen[hash] {
			matches = append(matches, submatch{
				node:   found,
				values: m.values,
			})
			seen[hash] = true
		}
	}
	for _, sub := range subs {
		startValues = valsCopy(sub.values)
		m.walkWithLists(cmd.value.(ast.Node), sub.node, match)
	}
	return matches
}

func (m *matcher) cmdFilter(wantAny bool) func(exprCmd, []submatch) []submatch {
	return func(cmd exprCmd, subs []submatch) []submatch {
		var matches []submatch
		any := false
		match := func(exprNode, node ast.Node) {
			if node == nil {
				return
			}
			found := m.topNode(exprNode, node)
			if found != nil {
				any = true
			}
		}
		for _, sub := range subs {
			any = false
			m.values = sub.values
			m.walkWithLists(cmd.value.(ast.Node), sub.node, match)
			if any == wantAny {
				matches = append(matches, sub)
			}
		}
		return matches
	}
}

func (m *matcher) cmdAttr(cmd exprCmd, subs []submatch) []submatch {
	var matches []submatch
	for _, sub := range subs {
		m.values = sub.values
		if m.attrApplies(sub.node, cmd.value.(attribute)) {
			matches = append(matches, sub)
		}
	}
	return matches
}

func (m *matcher) cmdParents(cmd exprCmd, subs []submatch) []submatch {
	for i := range subs {
		sub := &subs[i]
		reps := cmd.value.(int)
		for j := 0; j < reps; j++ {
			sub.node = m.parentOf(sub.node)
		}
	}
	return subs
}

func (m *matcher) attrApplies(node ast.Node, attr interface{}) bool {
	if rx, ok := attr.(*regexp.Regexp); ok {
		if exprStmt, ok := node.(*ast.ExprStmt); ok {
			// since we prefer matching entire statements, get the
			// ident from the ExprStmt
			node = exprStmt.X
		}
		ident, ok := node.(*ast.Ident)
		return ok && rx.MatchString(ident.Name)
	}
	expr, _ := node.(ast.Expr)
	if expr == nil {
		return false // only exprs have types
	}
	t := m.Info.TypeOf(expr)
	if t == nil {
		return false // an expr, but no type?
	}
	tv := m.Info.Types[expr]
	switch x := attr.(type) {
	case typeCheck:
		want := m.resolveType(m.scope, x.expr)
		switch {
		case x.op == "type" && !types.Identical(t, want):
			return false
		case x.op == "asgn" && !types.AssignableTo(t, want):
			return false
		case x.op == "conv" && !types.ConvertibleTo(t, want):
			return false
		}
	case typProperty:
		switch {
		case x == "comp" && !types.Comparable(t):
			return false
		case x == "addr" && !tv.Addressable():
			return false
		}
	case typUnderlying:
		u := t.Underlying()
		uok := true
		switch x {
		case "basic":
			_, uok = u.(*types.Basic)
		case "array":
			_, uok = u.(*types.Array)
		case "slice":
			_, uok = u.(*types.Slice)
		case "struct":
			_, uok = u.(*types.Struct)
		case "interface":
			_, uok = u.(*types.Interface)
		case "pointer":
			_, uok = u.(*types.Pointer)
		case "func":
			_, uok = u.(*types.Signature)
		case "map":
			_, uok = u.(*types.Map)
		case "chan":
			_, uok = u.(*types.Chan)
		}
		if !uok {
			return false
		}
	}
	return true
}

func (m *matcher) walkWithLists(exprNode, node ast.Node, fn func(exprNode, node ast.Node)) {
	visit := func(node ast.Node) bool {
		fn(exprNode, node)
		for _, list := range nodeLists(node) {
			fn(exprNode, list)
			if id := m.wildAnyIdent(exprNode); id != nil {
				// so that "$*a" will match "a, b"
				fn(exprList([]ast.Expr{id}), list)
				// so that "$*a" will match "a; b"
				fn(toStmtList(id), list)
			}
		}
		return true
	}
	inspect(node, visit)
}

func (m *matcher) topNode(exprNode, node ast.Node) ast.Node {
	sts1, ok1 := exprNode.(stmtList)
	sts2, ok2 := node.(stmtList)
	if ok1 && ok2 {
		// allow a partial match at the top level
		return m.nodes(sts1, sts2, true)
	}
	if m.node(exprNode, node) {
		return node
	}
	return nil
}

// optNode is like node, but for those nodes that can be nil and are not
// part of a list. For example, init and post statements in a for loop.
func (m *matcher) optNode(expr, node ast.Node) bool {
	if ident := m.wildAnyIdent(expr); ident != nil {
		if m.node(toStmtList(ident), toStmtList(node)) {
			return true
		}
	}
	return m.node(expr, node)
}

func (m *matcher) node(expr, node ast.Node) bool {
	switch node.(type) {
	case *ast.File, *ast.FuncType, *ast.BlockStmt, *ast.IfStmt,
		*ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.CaseClause,
		*ast.CommClause, *ast.ForStmt, *ast.RangeStmt:
		if scope := m.Info.Scopes[node]; scope != nil {
			m.scope = scope
		}
	}
	if !m.aggressive {
		if expr == nil || node == nil {
			return expr == node
		}
	} else {
		if expr == nil && node == nil {
			return true
		}
		if node == nil {
			expr, node = node, expr
		}
	}
	switch x := expr.(type) {
	case nil: // only in aggressive mode
		y, ok := node.(*ast.Ident)
		return ok && y.Name == "_"

	case *ast.File:
		y, ok := node.(*ast.File)
		if !ok || !m.node(x.Name, y.Name) || len(x.Decls) != len(y.Decls) ||
			len(x.Imports) != len(y.Imports) {
			return false
		}
		for i, decl := range x.Decls {
			if !m.node(decl, y.Decls[i]) {
				return false
			}
		}
		for i, imp := range x.Imports {
			if !m.node(imp, y.Imports[i]) {
				return false
			}
		}
		return true

	case *ast.Ident:
		y, yok := node.(*ast.Ident)
		if !isWildName(x.Name) {
			// not a wildcard
			return yok && x.Name == y.Name
		}
		if _, ok := node.(ast.Node); !ok {
			return false // to not include our extra node types
		}
		id := fromWildName(x.Name)
		info := m.info(id)
		if info.any {
			return false
		}
		if info.name == "_" {
			// values are discarded, matches anything
			return true
		}
		prev, ok := m.values[info.name]
		if !ok {
			// first occurrence, record value
			m.values[info.name] = node
			return true
		}
		// multiple uses must match
		return m.node(prev, node)

	// lists (ys are generated by us while walking)
	case exprList:
		y, ok := node.(exprList)
		return ok && m.exprs(x, y)
	case stmtList:
		y, ok := node.(stmtList)
		return ok && m.stmts(x, y)

	// lits
	case *ast.BasicLit:
		y, ok := node.(*ast.BasicLit)
		return ok && x.Kind == y.Kind && x.Value == y.Value
	case *ast.CompositeLit:
		y, ok := node.(*ast.CompositeLit)
		return ok && m.node(x.Type, y.Type) && m.exprs(x.Elts, y.Elts)
	case *ast.FuncLit:
		y, ok := node.(*ast.FuncLit)
		return ok && m.node(x.Type, y.Type) && m.node(x.Body, y.Body)

	// types
	case *ast.ArrayType:
		y, ok := node.(*ast.ArrayType)
		return ok && m.node(x.Len, y.Len) && m.node(x.Elt, y.Elt)
	case *ast.MapType:
		y, ok := node.(*ast.MapType)
		return ok && m.node(x.Key, y.Key) && m.node(x.Value, y.Value)
	case *ast.StructType:
		y, ok := node.(*ast.StructType)
		return ok && m.fields(x.Fields, y.Fields)
	case *ast.Field:
		// TODO: tags?
		y, ok := node.(*ast.Field)
		if !ok {
			return false
		}
		if len(x.Names) == 0 && x.Tag == nil && m.node(x.Type, y) {
			// Allow $var to match a field.
			return true
		}
		return m.idents(x.Names, y.Names) && m.node(x.Type, y.Type)
	case *ast.FuncType:
		y, ok := node.(*ast.FuncType)
		return ok && m.fields(x.Params, y.Params) &&
			m.fields(x.Results, y.Results)
	case *ast.InterfaceType:
		y, ok := node.(*ast.InterfaceType)
		return ok && m.fields(x.Methods, y.Methods)
	case *ast.ChanType:
		y, ok := node.(*ast.ChanType)
		return ok && x.Dir == y.Dir && m.node(x.Value, y.Value)

	// other exprs
	case *ast.Ellipsis:
		y, ok := node.(*ast.Ellipsis)
		return ok && m.node(x.Elt, y.Elt)
	case *ast.ParenExpr:
		y, ok := node.(*ast.ParenExpr)
		return ok && m.node(x.X, y.X)
	case *ast.UnaryExpr:
		y, ok := node.(*ast.UnaryExpr)
		return ok && x.Op == y.Op && m.node(x.X, y.X)
	case *ast.BinaryExpr:
		y, ok := node.(*ast.BinaryExpr)
		return ok && x.Op == y.Op && m.node(x.X, y.X) && m.node(x.Y, y.Y)
	case *ast.CallExpr:
		y, ok := node.(*ast.CallExpr)
		return ok && m.node(x.Fun, y.Fun) && m.exprs(x.Args, y.Args) &&
			bothValid(x.Ellipsis, y.Ellipsis)
	case *ast.KeyValueExpr:
		y, ok := node.(*ast.KeyValueExpr)
		return ok && m.node(x.Key, y.Key) && m.node(x.Value, y.Value)
	case *ast.StarExpr:
		y, ok := node.(*ast.StarExpr)
		return ok && m.node(x.X, y.X)
	case *ast.SelectorExpr:
		y, ok := node.(*ast.SelectorExpr)
		return ok && m.node(x.X, y.X) && m.node(x.Sel, y.Sel)
	case *ast.IndexExpr:
		y, ok := node.(*ast.IndexExpr)
		return ok && m.node(x.X, y.X) && m.node(x.Index, y.Index)
	case *ast.SliceExpr:
		y, ok := node.(*ast.SliceExpr)
		return ok && m.node(x.X, y.X) && m.node(x.Low, y.Low) &&
			m.node(x.High, y.High) && m.node(x.Max, y.Max)
	case *ast.TypeAssertExpr:
		y, ok := node.(*ast.TypeAssertExpr)
		return ok && m.node(x.X, y.X) && m.node(x.Type, y.Type)

	// decls
	case *ast.GenDecl:
		y, ok := node.(*ast.GenDecl)
		return ok && x.Tok == y.Tok && m.specs(x.Specs, y.Specs)
	case *ast.FuncDecl:
		y, ok := node.(*ast.FuncDecl)
		return ok && m.fields(x.Recv, y.Recv) && m.node(x.Name, y.Name) &&
			m.node(x.Type, y.Type) && m.node(x.Body, y.Body)

	// specs
	case *ast.ValueSpec:
		y, ok := node.(*ast.ValueSpec)
		if !ok || !m.node(x.Type, y.Type) {
			return false
		}
		if m.aggressive && len(x.Names) == 1 {
			for i := range y.Names {
				if m.node(x.Names[i], y.Names[i]) &&
					(x.Values == nil || m.node(x.Values[i], y.Values[i])) {
					return true
				}
			}
		}
		return m.idents(x.Names, y.Names) && m.exprs(x.Values, y.Values)

	// stmt bridge nodes
	case *ast.ExprStmt:
		if id, ok := x.X.(*ast.Ident); ok && isWildName(id.Name) {
			// prefer matching $x as a statement, as it's
			// the parent
			return m.node(id, node)
		}
		y, ok := node.(*ast.ExprStmt)
		return ok && m.node(x.X, y.X)
	case *ast.DeclStmt:
		y, ok := node.(*ast.DeclStmt)
		return ok && m.node(x.Decl, y.Decl)

	// stmts
	case *ast.EmptyStmt:
		_, ok := node.(*ast.EmptyStmt)
		return ok
	case *ast.LabeledStmt:
		y, ok := node.(*ast.LabeledStmt)
		return ok && m.node(x.Label, y.Label) && m.node(x.Stmt, y.Stmt)
	case *ast.SendStmt:
		y, ok := node.(*ast.SendStmt)
		return ok && m.node(x.Chan, y.Chan) && m.node(x.Value, y.Value)
	case *ast.IncDecStmt:
		y, ok := node.(*ast.IncDecStmt)
		return ok && x.Tok == y.Tok && m.node(x.X, y.X)
	case *ast.AssignStmt:
		y, ok := node.(*ast.AssignStmt)
		if !m.aggressive {
			return ok && x.Tok == y.Tok &&
				m.exprs(x.Lhs, y.Lhs) && m.exprs(x.Rhs, y.Rhs)
		}
		if ok {
			return m.exprs(x.Lhs, y.Lhs) && m.exprs(x.Rhs, y.Rhs)
		}
		vs, ok := node.(*ast.ValueSpec)
		return ok && m.nodesMatch(exprList(x.Lhs), identList(vs.Names)) &&
			m.exprs(x.Rhs, vs.Values)
	case *ast.GoStmt:
		y, ok := node.(*ast.GoStmt)
		return ok && m.node(x.Call, y.Call)
	case *ast.DeferStmt:
		y, ok := node.(*ast.DeferStmt)
		return ok && m.node(x.Call, y.Call)
	case *ast.ReturnStmt:
		y, ok := node.(*ast.ReturnStmt)
		return ok && m.exprs(x.Results, y.Results)
	case *ast.BranchStmt:
		y, ok := node.(*ast.BranchStmt)
		return ok && x.Tok == y.Tok && m.node(maybeNilIdent(x.Label), maybeNilIdent(y.Label))
	case *ast.BlockStmt:
		if m.aggressive && m.node(stmtList(x.List), node) {
			return true
		}
		y, ok := node.(*ast.BlockStmt)
		if !ok {
			return false
		}
		if x == nil || y == nil {
			return x == y
		}
		return m.cases(x.List, y.List) || m.stmts(x.List, y.List)
	case *ast.IfStmt:
		y, ok := node.(*ast.IfStmt)
		if !ok {
			return false
		}
		condAny := m.wildAnyIdent(x.Cond)
		if condAny != nil && x.Init == nil {
			// if $*x { ... } on the left
			left := toStmtList(condAny)
			return m.node(left, toStmtList(y.Init, y.Cond)) &&
				m.node(x.Body, y.Body) && m.optNode(x.Else, y.Else)
		}
		return m.optNode(x.Init, y.Init) && m.node(x.Cond, y.Cond) &&
			m.node(x.Body, y.Body) && m.node(x.Else, y.Else)
	case *ast.CaseClause:
		y, ok := node.(*ast.CaseClause)
		return ok && m.exprs(x.List, y.List) && m.stmts(x.Body, y.Body)
	case *ast.SwitchStmt:
		y, ok := node.(*ast.SwitchStmt)
		if !ok {
			return false
		}
		tagAny := m.wildAnyIdent(x.Tag)
		if tagAny != nil && x.Init == nil {
			// switch $*x { ... } on the left
			left := toStmtList(tagAny)
			return m.node(left, toStmtList(y.Init, y.Tag)) &&
				m.node(x.Body, y.Body)
		}
		return m.optNode(x.Init, y.Init) && m.node(x.Tag, y.Tag) && m.node(x.Body, y.Body)
	case *ast.TypeSwitchStmt:
		y, ok := node.(*ast.TypeSwitchStmt)
		return ok && m.optNode(x.Init, y.Init) && m.node(x.Assign, y.Assign) && m.node(x.Body, y.Body)
	case *ast.CommClause:
		y, ok := node.(*ast.CommClause)
		return ok && m.node(x.Comm, y.Comm) && m.stmts(x.Body, y.Body)
	case *ast.SelectStmt:
		y, ok := node.(*ast.SelectStmt)
		return ok && m.node(x.Body, y.Body)
	case *ast.ForStmt:
		condIdent := m.wildAnyIdent(x.Cond)
		if condIdent != nil && x.Init == nil && x.Post == nil {
			// "for $*x { ... }" on the left
			left := toStmtList(condIdent)
			// also accept RangeStmt on the right
			switch y := node.(type) {
			case *ast.ForStmt:
				return m.node(left, toStmtList(y.Init, y.Cond, y.Post)) &&
					m.node(x.Body, y.Body)
			case *ast.RangeStmt:
				return m.node(left, toStmtList(y.Key, y.Value, y.X)) &&
					m.node(x.Body, y.Body)
			default:
				return false
			}
		}
		y, ok := node.(*ast.ForStmt)
		if !ok {
			return false
		}
		return m.optNode(x.Init, y.Init) && m.node(x.Cond, y.Cond) &&
			m.optNode(x.Post, y.Post) && m.node(x.Body, y.Body)
	case *ast.RangeStmt:
		y, ok := node.(*ast.RangeStmt)
		return ok && m.node(x.Key, y.Key) && m.node(x.Value, y.Value) &&
			m.node(x.X, y.X) && m.node(x.Body, y.Body)

	case *ast.TypeSpec:
		y, ok := node.(*ast.TypeSpec)
		return ok && m.node(x.Name, y.Name) && m.node(x.Type, y.Type)

	case *ast.FieldList:
		// we ignore these, for now
		return false
	default:
		panic(fmt.Sprintf("unexpected node: %T", x))
	}
}

func (m *matcher) wildAnyIdent(node ast.Node) *ast.Ident {
	switch x := node.(type) {
	case *ast.ExprStmt:
		return m.wildAnyIdent(x.X)
	case *ast.Ident:
		if !isWildName(x.Name) {
			return nil
		}
		if !m.info(fromWildName(x.Name)).any {
			return nil
		}
		return x
	}
	return nil
}

// resolveType resolves a type expression from a given scope.
func (m *matcher) resolveType(scope *types.Scope, expr ast.Expr) types.Type {
	switch x := expr.(type) {
	case *ast.Ident:
		_, obj := scope.LookupParent(x.Name, token.NoPos)
		if obj == nil {
			// TODO: error if all resolveType calls on a type
			// expression fail? or perhaps resolve type expressions
			// across the entire program?
			return nil
		}
		return obj.Type()
	case *ast.ArrayType:
		elt := m.resolveType(scope, x.Elt)
		if x.Len == nil {
			return types.NewSlice(elt)
		}
		bl, ok := x.Len.(*ast.BasicLit)
		if !ok || bl.Kind != token.INT {
			panic(fmt.Sprintf("TODO: %T", x))
		}
		len, _ := strconv.ParseInt(bl.Value, 0, 0)
		return types.NewArray(elt, len)
	case *ast.StarExpr:
		return types.NewPointer(m.resolveType(scope, x.X))
	case *ast.ChanType:
		dir := types.SendRecv
		switch x.Dir {
		case ast.SEND:
			dir = types.SendOnly
		case ast.RECV:
			dir = types.RecvOnly
		}
		return types.NewChan(dir, m.resolveType(scope, x.Value))
	case *ast.SelectorExpr:
		scope = m.findScope(scope, x.X)
		return m.resolveType(scope, x.Sel)
	default:
		panic(fmt.Sprintf("resolveType TODO: %T", x))
	}
}

func (m *matcher) findScope(scope *types.Scope, expr ast.Expr) *types.Scope {
	switch x := expr.(type) {
	case *ast.Ident:
		_, obj := scope.LookupParent(x.Name, token.NoPos)
		if pkg, ok := obj.(*types.PkgName); ok {
			return pkg.Imported().Scope()
		}
		// try to fall back to std
		if m.stdImporter == nil {
			m.stdImporter = importer.Default()
		}
		path := x.Name
		if longer, ok := stdImportFixes[path]; ok {
			path = longer
		}
		pkg, err := m.stdImporter.Import(path)
		if err != nil {
			panic(fmt.Sprintf("findScope err: %v", err))
		}
		return pkg.Scope()
	default:
		panic(fmt.Sprintf("findScope TODO: %T", x))
	}
}

var stdImportFixes = map[string]string{
	// go list std | grep -vE 'vendor|internal' | grep '/' | sed -r 's@^(.*)/([^/]*)$@"\2": "\1/\2",@' | sort
	// (after commenting out the less likely duplicates)
	"adler32":   "hash/adler32",
	"aes":       "crypto/aes",
	"ascii85":   "encoding/ascii85",
	"asn1":      "encoding/asn1",
	"ast":       "go/ast",
	"atomic":    "sync/atomic",
	"base32":    "encoding/base32",
	"base64":    "encoding/base64",
	"big":       "math/big",
	"binary":    "encoding/binary",
	"bits":      "math/bits",
	"build":     "go/build",
	"bzip2":     "compress/bzip2",
	"cgi":       "net/http/cgi",
	"cgo":       "runtime/cgo",
	"cipher":    "crypto/cipher",
	"cmplx":     "math/cmplx",
	"color":     "image/color",
	"constant":  "go/constant",
	"cookiejar": "net/http/cookiejar",
	"crc32":     "hash/crc32",
	"crc64":     "hash/crc64",
	"csv":       "encoding/csv",
	"debug":     "runtime/debug",
	"des":       "crypto/des",
	"doc":       "go/doc",
	"draw":      "image/draw",
	"driver":    "database/sql/driver",
	"dsa":       "crypto/dsa",
	"dwarf":     "debug/dwarf",
	"ecdsa":     "crypto/ecdsa",
	"elf":       "debug/elf",
	"elliptic":  "crypto/elliptic",
	"exec":      "os/exec",
	"fcgi":      "net/http/fcgi",
	"filepath":  "path/filepath",
	"flate":     "compress/flate",
	"fnv":       "hash/fnv",
	"format":    "go/format",
	"gif":       "image/gif",
	"gob":       "encoding/gob",
	"gosym":     "debug/gosym",
	"gzip":      "compress/gzip",
	"heap":      "container/heap",
	"hex":       "encoding/hex",
	"hmac":      "crypto/hmac",
	"http":      "net/http",
	"httptest":  "net/http/httptest",
	"httptrace": "net/http/httptrace",
	"httputil":  "net/http/httputil",
	"importer":  "go/importer",
	"iotest":    "testing/iotest",
	"ioutil":    "io/ioutil",
	"jpeg":      "image/jpeg",
	"json":      "encoding/json",
	"jsonrpc":   "net/rpc/jsonrpc",
	"list":      "container/list",
	"lzw":       "compress/lzw",
	"macho":     "debug/macho",
	"mail":      "net/mail",
	"md5":       "crypto/md5",
	"multipart": "mime/multipart",
	"palette":   "image/color/palette",
	"parser":    "go/parser",
	"parse":     "text/template/parse",
	"pe":        "debug/pe",
	"pem":       "encoding/pem",
	"pkix":      "crypto/x509/pkix",
	"plan9obj":  "debug/plan9obj",
	"png":       "image/png",
	//"pprof":   "net/http/pprof",
	"pprof":           "runtime/pprof",
	"printer":         "go/printer",
	"quick":           "testing/quick",
	"quotedprintable": "mime/quotedprintable",
	"race":            "runtime/race",
	//"rand":          "crypto/rand",
	"rand": "math/rand",
	"rc4":  "crypto/rc4",
	"ring": "container/ring",
	"rpc":  "net/rpc",
	"rsa":  "crypto/rsa",
	//"scanner":   "go/scanner",
	"scanner":     "text/scanner",
	"sha1":        "crypto/sha1",
	"sha256":      "crypto/sha256",
	"sha512":      "crypto/sha512",
	"signal":      "os/signal",
	"smtp":        "net/smtp",
	"sql":         "database/sql",
	"subtle":      "crypto/subtle",
	"suffixarray": "index/suffixarray",
	"syntax":      "regexp/syntax",
	"syslog":      "log/syslog",
	"tabwriter":   "text/tabwriter",
	"tar":         "archive/tar",
	//"template":  "html/template",
	"template":  "text/template",
	"textproto": "net/textproto",
	"tls":       "crypto/tls",
	"token":     "go/token",
	"trace":     "runtime/trace",
	"types":     "go/types",
	"url":       "net/url",
	"user":      "os/user",
	"utf16":     "unicode/utf16",
	"utf8":      "unicode/utf8",
	"x509":      "crypto/x509",
	"xml":       "encoding/xml",
	"zip":       "archive/zip",
	"zlib":      "compress/zlib",
}

func maybeNilIdent(x *ast.Ident) ast.Node {
	if x == nil {
		return nil
	}
	return x
}

func bothValid(p1, p2 token.Pos) bool {
	return p1.IsValid() == p2.IsValid()
}

type nodeList interface {
	at(i int) ast.Node
	len() int
	slice(from, to int) nodeList
	ast.Node
}

// nodes matches two lists of nodes. It uses a common algorithm to match
// wildcard patterns with any number of nodes without recursion.
func (m *matcher) nodes(ns1, ns2 nodeList, partial bool) ast.Node {
	ns1len, ns2len := ns1.len(), ns2.len()
	if ns1len == 0 {
		if ns2len == 0 {
			return ns2
		}
		return nil
	}
	partialStart, partialEnd := 0, ns2len
	i1, i2 := 0, 0
	next1, next2 := 0, 0

	// We need to keep a copy of m.values so that we can restart
	// with a different "any of" match while discarding any matches
	// we found while trying it.
	type restart struct {
		matches      map[string]ast.Node
		next1, next2 int
	}
	// We need to stack these because otherwise some edge cases
	// would not match properly. Since we have various kinds of
	// wildcards (nodes containing them, $_, and $*_), in some cases
	// we may have to go back and do multiple restarts to get to the
	// right starting position.
	var stack []restart
	push := func(n1, n2 int) {
		if n2 > ns2len {
			return // would be discarded anyway
		}
		stack = append(stack, restart{valsCopy(m.values), n1, n2})
		next1, next2 = n1, n2
	}
	pop := func() {
		i1, i2 = next1, next2
		m.values = stack[len(stack)-1].matches
		stack = stack[:len(stack)-1]
		next1, next2 = 0, 0
		if len(stack) > 0 {
			next1 = stack[len(stack)-1].next1
			next2 = stack[len(stack)-1].next2
		}
	}
	wildName := ""
	wildStart := 0

	// wouldMatch returns whether the current wildcard - if any -
	// matches the nodes we are currently trying it on.
	wouldMatch := func() bool {
		switch wildName {
		case "", "_":
			return true
		}
		list := ns2.slice(wildStart, i2)
		// check that it matches any nodes found elsewhere
		prev, ok := m.values[wildName]
		if ok && !m.node(prev, list) {
			return false
		}
		m.values[wildName] = list
		return true
	}
	for i1 < ns1len || i2 < ns2len {
		if i1 < ns1len {
			n1 := ns1.at(i1)
			id := fromWildNode(n1)
			info := m.info(id)
			if info.any {
				// keep track of where this wildcard
				// started (if info.name == wildName,
				// we're trying the same wildcard
				// matching one more node)
				if info.name != wildName {
					wildStart = i2
					wildName = info.name
				}
				// try to match zero or more at i2,
				// restarting at i2+1 if it fails
				push(i1, i2+1)
				i1++
				continue
			}
			if partial && i1 == 0 {
				// let "b; c" match "a; b; c"
				// (simulates a $*_ at the beginning)
				partialStart = i2
				push(i1, i2+1)
			}
			if i2 < ns2len && wouldMatch() && m.node(n1, ns2.at(i2)) {
				wildName = ""
				// ordinary match
				i1++
				i2++
				continue
			}
		}
		if partial && i1 == ns1len && wildName == "" {
			partialEnd = i2
			break // let "b; c" match "b; c; d"
		}
		// mismatch, try to restart
		if 0 < next2 && next2 <= ns2len && (i1 != next1 || i2 != next2) {
			pop()
			continue
		}
		return nil
	}
	if !wouldMatch() {
		return nil
	}
	return ns2.slice(partialStart, partialEnd)
}

func (m *matcher) nodesMatch(list1, list2 nodeList) bool {
	return m.nodes(list1, list2, false) != nil
}

func (m *matcher) exprs(exprs1, exprs2 []ast.Expr) bool {
	return m.nodesMatch(exprList(exprs1), exprList(exprs2))
}

func (m *matcher) idents(ids1, ids2 []*ast.Ident) bool {
	return m.nodesMatch(identList(ids1), identList(ids2))
}

func toStmtList(nodes ...ast.Node) stmtList {
	var stmts []ast.Stmt
	for _, node := range nodes {
		switch x := node.(type) {
		case nil:
		case ast.Stmt:
			stmts = append(stmts, x)
		case ast.Expr:
			stmts = append(stmts, &ast.ExprStmt{X: x})
		default:
			panic(fmt.Sprintf("unexpected node type: %T", x))
		}
	}
	return stmtList(stmts)
}

func (m *matcher) cases(stmts1, stmts2 []ast.Stmt) bool {
	for _, stmt := range stmts2 {
		switch stmt.(type) {
		case *ast.CaseClause, *ast.CommClause:
		default:
			return false
		}
	}
	var left []*ast.Ident
	for _, stmt := range stmts1 {
		var expr ast.Expr
		var bstmt ast.Stmt
		switch x := stmt.(type) {
		case *ast.CaseClause:
			if len(x.List) != 1 || len(x.Body) != 1 {
				return false
			}
			expr, bstmt = x.List[0], x.Body[0]
		case *ast.CommClause:
			if x.Comm == nil || len(x.Body) != 1 {
				return false
			}
			if commExpr, ok := x.Comm.(*ast.ExprStmt); ok {
				expr = commExpr.X
			}
			bstmt = x.Body[0]
		default:
			return false
		}
		xs, ok := bstmt.(*ast.ExprStmt)
		if !ok {
			return false
		}
		bodyIdent, ok := xs.X.(*ast.Ident)
		if !ok || bodyIdent.Name != "gogrep_body" {
			return false
		}
		id, ok := expr.(*ast.Ident)
		if !ok || !isWildName(id.Name) {
			return false
		}
		left = append(left, id)
	}
	return m.nodesMatch(identList(left), stmtList(stmts2))
}

func (m *matcher) stmts(stmts1, stmts2 []ast.Stmt) bool {
	return m.nodesMatch(stmtList(stmts1), stmtList(stmts2))
}

func (m *matcher) specs(specs1, specs2 []ast.Spec) bool {
	return m.nodesMatch(specList(specs1), specList(specs2))
}

func (m *matcher) fields(fields1, fields2 *ast.FieldList) bool {
	if fields1 == nil || fields2 == nil {
		return fields1 == fields2
	}
	return m.nodesMatch(fieldList(fields1.List), fieldList(fields2.List))
}

func fromWildNode(node ast.Node) int {
	switch node := node.(type) {
	case *ast.Ident:
		return fromWildName(node.Name)
	case *ast.ExprStmt:
		return fromWildNode(node.X)
	case *ast.Field:
		// Allow $var to represent an entire field; the lone identifier
		// gets picked up as an anonymous field.
		if len(node.Names) == 0 && node.Tag == nil {
			return fromWildNode(node.Type)
		}
	}
	return -1
}

func nodeLists(n ast.Node) []nodeList {
	var lists []nodeList
	addList := func(list nodeList) {
		if list.len() > 0 {
			lists = append(lists, list)
		}
	}
	switch x := n.(type) {
	case nodeList:
		addList(x)
	case *ast.CompositeLit:
		addList(exprList(x.Elts))
	case *ast.CallExpr:
		addList(exprList(x.Args))
	case *ast.AssignStmt:
		addList(exprList(x.Lhs))
		addList(exprList(x.Rhs))
	case *ast.ReturnStmt:
		addList(exprList(x.Results))
	case *ast.ValueSpec:
		addList(exprList(x.Values))
	case *ast.BlockStmt:
		addList(stmtList(x.List))
	case *ast.CaseClause:
		addList(exprList(x.List))
		addList(stmtList(x.Body))
	case *ast.CommClause:
		addList(stmtList(x.Body))
	}
	return lists
}

type exprList []ast.Expr
type identList []*ast.Ident
type stmtList []ast.Stmt
type specList []ast.Spec
type fieldList []*ast.Field

func (l exprList) len() int  { return len(l) }
func (l identList) len() int { return len(l) }
func (l stmtList) len() int  { return len(l) }
func (l specList) len() int  { return len(l) }
func (l fieldList) len() int { return len(l) }

func (l exprList) at(i int) ast.Node  { return l[i] }
func (l identList) at(i int) ast.Node { return l[i] }
func (l stmtList) at(i int) ast.Node  { return l[i] }
func (l specList) at(i int) ast.Node  { return l[i] }
func (l fieldList) at(i int) ast.Node { return l[i] }

func (l exprList) slice(i, j int) nodeList  { return l[i:j] }
func (l identList) slice(i, j int) nodeList { return l[i:j] }
func (l stmtList) slice(i, j int) nodeList  { return l[i:j] }
func (l specList) slice(i, j int) nodeList  { return l[i:j] }
func (l fieldList) slice(i, j int) nodeList { return l[i:j] }

func (l exprList) Pos() token.Pos  { return l[0].Pos() }
func (l identList) Pos() token.Pos { return l[0].Pos() }
func (l stmtList) Pos() token.Pos  { return l[0].Pos() }
func (l specList) Pos() token.Pos  { return l[0].Pos() }
func (l fieldList) Pos() token.Pos { return l[0].Pos() }

func (l exprList) End() token.Pos  { return l[len(l)-1].End() }
func (l identList) End() token.Pos { return l[len(l)-1].End() }
func (l stmtList) End() token.Pos  { return l[len(l)-1].End() }
func (l specList) End() token.Pos  { return l[len(l)-1].End() }
func (l fieldList) End() token.Pos { return l[len(l)-1].End() }
