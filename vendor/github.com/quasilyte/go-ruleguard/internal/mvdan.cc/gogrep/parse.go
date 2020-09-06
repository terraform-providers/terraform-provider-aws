// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package gogrep

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

func (m *matcher) transformSource(expr string) (string, []posOffset, error) {
	toks, err := m.tokenize([]byte(expr))
	if err != nil {
		return "", nil, fmt.Errorf("cannot tokenize expr: %v", err)
	}
	var offs []posOffset
	lbuf := lineColBuffer{line: 1, col: 1}
	addOffset := func(length int) {
		lbuf.offs -= length
		offs = append(offs, posOffset{
			atLine: lbuf.line,
			atCol:  lbuf.col,
			offset: length,
		})
	}
	if len(toks) > 0 && toks[0].tok == tokAggressive {
		toks = toks[1:]
		m.aggressive = true
	}
	lastLit := false
	for _, t := range toks {
		if lbuf.offs >= t.pos.Offset && lastLit && t.lit != "" {
			lbuf.WriteString(" ")
		}
		for lbuf.offs < t.pos.Offset {
			lbuf.WriteString(" ")
		}
		if t.lit == "" {
			lbuf.WriteString(t.tok.String())
			lastLit = false
			continue
		}
		if isWildName(t.lit) {
			// to correct the position offsets for the extra
			// info attached to ident name strings
			addOffset(len(wildPrefix) - 1)
		}
		lbuf.WriteString(t.lit)
		lastLit = strings.TrimSpace(t.lit) != ""
	}
	// trailing newlines can cause issues with commas
	return strings.TrimSpace(lbuf.String()), offs, nil
}

func (m *matcher) parseExpr(expr string) (ast.Node, error) {
	exprStr, offs, err := m.transformSource(expr)
	if err != nil {
		return nil, err
	}
	node, _, err := parseDetectingNode(m.fset, exprStr)
	if err != nil {
		err = subPosOffsets(err, offs...)
		return nil, fmt.Errorf("cannot parse expr: %v", err)
	}
	return node, nil
}

type lineColBuffer struct {
	bytes.Buffer
	line, col, offs int
}

func (l *lineColBuffer) WriteString(s string) (n int, err error) {
	for _, r := range s {
		if r == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.offs++
	}
	return l.Buffer.WriteString(s)
}

var tmplDecl = template.Must(template.New("").Parse(`` +
	`package p; {{ . }}`))

var tmplExprs = template.Must(template.New("").Parse(`` +
	`package p; var _ = []interface{}{ {{ . }}, }`))

var tmplStmts = template.Must(template.New("").Parse(`` +
	`package p; func _() { {{ . }} }`))

var tmplType = template.Must(template.New("").Parse(`` +
	`package p; var _ {{ . }}`))

var tmplValSpec = template.Must(template.New("").Parse(`` +
	`package p; var {{ . }}`))

func execTmpl(tmpl *template.Template, src string) string {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, src); err != nil {
		panic(err)
	}
	return buf.String()
}

func noBadNodes(node ast.Node) bool {
	any := false
	ast.Inspect(node, func(n ast.Node) bool {
		if any {
			return false
		}
		switch n.(type) {
		case *ast.BadExpr, *ast.BadDecl:
			any = true
		}
		return true
	})
	return !any
}

func parseType(fset *token.FileSet, src string) (ast.Expr, *ast.File, error) {
	asType := execTmpl(tmplType, src)
	f, err := parser.ParseFile(fset, "", asType, 0)
	if err != nil {
		err = subPosOffsets(err, posOffset{1, 1, 17})
		return nil, nil, err
	}
	vs := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	return vs.Type, f, nil
}

// parseDetectingNode tries its best to parse the ast.Node contained in src, as
// one of: *ast.File, ast.Decl, ast.Expr, ast.Stmt, *ast.ValueSpec.
// It also returns the *ast.File used for the parsing, so that the returned node
// can be easily type-checked.
func parseDetectingNode(fset *token.FileSet, src string) (ast.Node, *ast.File, error) {
	file := fset.AddFile("", fset.Base(), len(src))
	scan := scanner.Scanner{}
	scan.Init(file, []byte(src), nil, 0)
	if _, tok, _ := scan.Scan(); tok == token.EOF {
		return nil, nil, fmt.Errorf("empty source code")
	}
	var mainErr error

	// first try as a whole file
	if f, err := parser.ParseFile(fset, "", src, 0); err == nil && noBadNodes(f) {
		return f, f, nil
	}

	// then as a single declaration, or many
	asDecl := execTmpl(tmplDecl, src)
	if f, err := parser.ParseFile(fset, "", asDecl, 0); err == nil && noBadNodes(f) {
		if len(f.Decls) == 1 {
			return f.Decls[0], f, nil
		}
		return f, f, nil
	}

	// then as value expressions
	asExprs := execTmpl(tmplExprs, src)
	if f, err := parser.ParseFile(fset, "", asExprs, 0); err == nil && noBadNodes(f) {
		vs := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
		cl := vs.Values[0].(*ast.CompositeLit)
		if len(cl.Elts) == 1 {
			return cl.Elts[0], f, nil
		}
		return exprList(cl.Elts), f, nil
	}

	// then try as statements
	asStmts := execTmpl(tmplStmts, src)
	if f, err := parser.ParseFile(fset, "", asStmts, 0); err == nil && noBadNodes(f) {
		bl := f.Decls[0].(*ast.FuncDecl).Body
		if len(bl.List) == 1 {
			return bl.List[0], f, nil
		}
		return stmtList(bl.List), f, nil
	} else {
		// Statements is what covers most cases, so it will give
		// the best overall error message. Show positions
		// relative to where the user's code is put in the
		// template.
		mainErr = subPosOffsets(err, posOffset{1, 1, 22})
	}

	// type expressions not yet picked up, for e.g. chans and interfaces
	if typ, f, err := parseType(fset, src); err == nil && noBadNodes(f) {
		return typ, f, nil
	}

	// value specs
	asValSpec := execTmpl(tmplValSpec, src)
	if f, err := parser.ParseFile(fset, "", asValSpec, 0); err == nil && noBadNodes(f) {
		vs := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
		return vs, f, nil
	}
	return nil, nil, mainErr
}

type posOffset struct {
	atLine, atCol int
	offset        int
}

func subPosOffsets(err error, offs ...posOffset) error {
	list, ok := err.(scanner.ErrorList)
	if !ok {
		return err
	}
	for i, err := range list {
		for _, off := range offs {
			if err.Pos.Line != off.atLine {
				continue
			}
			if err.Pos.Column < off.atCol {
				continue
			}
			err.Pos.Column -= off.offset
		}
		list[i] = err
	}
	return list
}

const (
	_ token.Token = -iota
	tokAggressive
)

type fullToken struct {
	pos token.Position
	tok token.Token
	lit string
}

type caseStatus uint

const (
	caseNone caseStatus = iota
	caseNeedBlock
	caseHere
)

func (m *matcher) tokenize(src []byte) ([]fullToken, error) {
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))

	var err error
	onError := func(pos token.Position, msg string) {
		switch msg { // allow certain extra chars
		case `illegal character U+0024 '$'`:
		case `illegal character U+007E '~'`:
		default:
			err = fmt.Errorf("%v: %s", pos, msg)
		}
	}

	// we will modify the input source under the scanner's nose to
	// enable some features such as regexes.
	s.Init(file, src, onError, scanner.ScanComments)

	next := func() fullToken {
		pos, tok, lit := s.Scan()
		return fullToken{fset.Position(pos), tok, lit}
	}

	caseStat := caseNone

	var toks []fullToken
	for t := next(); t.tok != token.EOF; t = next() {
		switch t.lit {
		case "$": // continues below
		case "~":
			toks = append(toks, fullToken{t.pos, tokAggressive, ""})
			continue
		case "switch", "select", "case":
			if t.lit == "case" {
				caseStat = caseNone
			} else {
				caseStat = caseNeedBlock
			}
			fallthrough
		default: // regular Go code
			if t.tok == token.LBRACE && caseStat == caseNeedBlock {
				caseStat = caseHere
			}
			toks = append(toks, t)
			continue
		}
		wt, err := m.wildcard(t.pos, next)
		if err != nil {
			return nil, err
		}
		if caseStat == caseHere {
			toks = append(toks, fullToken{wt.pos, token.IDENT, "case"})
		}
		toks = append(toks, wt)
		if caseStat == caseHere {
			toks = append(toks, fullToken{wt.pos, token.COLON, ""})
			toks = append(toks, fullToken{wt.pos, token.IDENT, "gogrep_body"})
		}
	}
	return toks, err
}

func (m *matcher) wildcard(pos token.Position, next func() fullToken) (fullToken, error) {
	wt := fullToken{pos, token.IDENT, wildPrefix}
	t := next()
	var info varInfo
	if t.tok == token.MUL {
		t = next()
		info.any = true
	}
	if t.tok != token.IDENT {
		return wt, fmt.Errorf("%v: $ must be followed by ident, got %v",
			t.pos, t.tok)
	}
	id := len(m.vars)
	wt.lit += strconv.Itoa(id)
	info.name = t.lit
	m.vars = append(m.vars, info)
	return wt, nil
}

type typeCheck struct {
	op   string // "type", "asgn", "conv"
	expr ast.Expr
}

type attribute interface{}

type typProperty string

type typUnderlying string

func (m *matcher) parseAttrs(src string) (attribute, error) {
	toks, err := m.tokenize([]byte(src))
	if err != nil {
		return nil, err
	}
	i := -1
	var t fullToken
	next := func() fullToken {
		if i++; i < len(toks) {
			return toks[i]
		}
		return fullToken{tok: token.EOF, pos: t.pos}
	}
	t = next()
	op := t.lit
	switch op { // the ones that don't take args
	case "comp", "addr":
		if t = next(); t.tok != token.SEMICOLON {
			return nil, fmt.Errorf("%v: wanted EOF, got %v", t.pos, t.tok)
		}
		return typProperty(op), nil
	}
	opPos := t.pos
	if t = next(); t.tok != token.LPAREN {
		return nil, fmt.Errorf("%v: wanted (", t.pos)
	}
	var attr attribute
	switch op {
	case "rx":
		t = next()
		rxStr, err := strconv.Unquote(t.lit)
		if err != nil {
			return nil, fmt.Errorf("%v: %v", t.pos, err)
		}
		if !strings.HasPrefix(rxStr, "^") {
			rxStr = "^" + rxStr
		}
		if !strings.HasSuffix(rxStr, "$") {
			rxStr = rxStr + "$"
		}
		rx, err := regexp.Compile(rxStr)
		if err != nil {
			return nil, fmt.Errorf("%v: %v", t.pos, err)
		}
		attr = rx
	case "type", "asgn", "conv":
		t = next()
		start := t.pos.Offset
		for open := 1; open > 0; t = next() {
			switch t.tok {
			case token.LPAREN:
				open++
			case token.RPAREN:
				open--
			case token.EOF:
				return nil, fmt.Errorf("%v: expected ) to close (", t.pos)
			}
		}
		end := t.pos.Offset - 1
		typeStr := strings.TrimSpace(string(src[start:end]))
		fset := token.NewFileSet()
		typeExpr, _, err := parseType(fset, typeStr)
		if err != nil {
			return nil, err
		}
		attr = typeCheck{op, typeExpr}
		i -= 2 // since we went past RPAREN above
	case "is":
		switch t = next(); t.lit {
		case "basic", "array", "slice", "struct", "interface",
			"pointer", "func", "map", "chan":
		default:
			return nil, fmt.Errorf("%v: unknown type: %q", t.pos,
				t.lit)
		}
		attr = typUnderlying(t.lit)
	default:
		return nil, fmt.Errorf("%v: unknown op %q", opPos, op)
	}
	if t = next(); t.tok != token.RPAREN {
		return nil, fmt.Errorf("%v: wanted ), got %v", t.pos, t.tok)
	}
	if t = next(); t.tok != token.SEMICOLON {
		return nil, fmt.Errorf("%v: wanted EOF, got %v", t.pos, t.tok)
	}
	return attr, nil
}

// using a prefix is good enough for now
const wildPrefix = "gogrep_"

func isWildName(name string) bool {
	return strings.HasPrefix(name, wildPrefix)
}

func fromWildName(s string) int {
	if !isWildName(s) {
		return -1
	}
	n, err := strconv.Atoi(s[len(wildPrefix):])
	if err != nil {
		return -1
	}
	return n
}
