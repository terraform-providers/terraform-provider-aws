// Copyright (c) 2017, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package gogrep

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var usage = func() {
	fmt.Fprint(os.Stderr, `usage: gogrep commands [packages]

gogrep performs a query on the given Go packages.

  -r      search dependencies recursively too
  -tests  search test files too (and direct test deps, with -r)

A command is one of the following:

  -x pattern    find all nodes matching a pattern
  -g pattern    discard nodes not matching a pattern
  -v pattern    discard nodes matching a pattern
  -a attribute  discard nodes without an attribute
  -s pattern    substitute with a given syntax tree
  -p number     navigate up a number of node parents
  -w            write the entire source code back

A pattern is a piece of Go code which may include dollar expressions. It can be
a number of statements, a number of expressions, a declaration, or an entire
file.

A dollar expression consist of '$' and a name. Dollar expressions with the same
name within a query always match the same node, excluding "_". Example:

       -x '$x.$_ = $x' # assignment of self to a field in self

If '*' is before the name, it will match any number of nodes. Example:

       -x 'fmt.Fprintf(os.Stdout, $*_)' # all Fprintfs on stdout

By default, the resulting nodes will be printed one per line to standard output.
To update the input files, use -w.
`)
}

func main() {
	m := matcher{
		out: os.Stdout,
		ctx: &build.Default,
	}
	err := m.fromArgs(".", os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type matcher struct {
	out io.Writer
	ctx *build.Context

	fset *token.FileSet

	parents map[ast.Node]ast.Node

	recursive, tests bool
	aggressive       bool

	// information about variables (wildcards), by id (which is an
	// integer starting at 0)
	vars []varInfo

	// node values recorded by name, excluding "_" (used only by the
	// actual matching phase)
	values map[string]ast.Node
	scope  *types.Scope

	*types.Info
	stdImporter types.Importer
}

type varInfo struct {
	name string
	any  bool
}

func (m *matcher) info(id int) varInfo {
	if id < 0 {
		return varInfo{}
	}
	return m.vars[id]
}

type exprCmd struct {
	name  string
	src   string
	value interface{}
}

type strCmdFlag struct {
	name string
	cmds *[]exprCmd
}

func (o *strCmdFlag) String() string { return "" }
func (o *strCmdFlag) Set(val string) error {
	*o.cmds = append(*o.cmds, exprCmd{name: o.name, src: val})
	return nil
}

type boolCmdFlag struct {
	name string
	cmds *[]exprCmd
}

func (o *boolCmdFlag) String() string { return "" }
func (o *boolCmdFlag) Set(val string) error {
	if val != "true" {
		return fmt.Errorf("flag can only be true")
	}
	*o.cmds = append(*o.cmds, exprCmd{name: o.name})
	return nil
}
func (o *boolCmdFlag) IsBoolFlag() bool { return true }

func (m *matcher) fromArgs(wd string, args []string) error {
	m.fset = token.NewFileSet()
	cmds, args, err := m.parseCmds(args)
	if err != nil {
		return err
	}
	pkgs, err := m.load(wd, args...)
	if err != nil {
		return err
	}
	var all []ast.Node
	for _, pkg := range pkgs {
		m.Info = pkg.TypesInfo
		nodes := make([]ast.Node, len(pkg.Syntax))
		for i, f := range pkg.Syntax {
			nodes[i] = f
		}
		all = append(all, m.matches(cmds, nodes)...)
	}
	for _, n := range all {
		fpos := m.fset.Position(n.Pos())
		if strings.HasPrefix(fpos.Filename, wd) {
			fpos.Filename = fpos.Filename[len(wd)+1:]
		}
		fmt.Fprintf(m.out, "%v: %s\n", fpos, singleLinePrint(n))
	}
	return nil
}

func (m *matcher) parseCmds(args []string) ([]exprCmd, []string, error) {
	flagSet := flag.NewFlagSet("gogrep", flag.ExitOnError)
	flagSet.Usage = usage
	flagSet.BoolVar(&m.recursive, "r", false, "search dependencies recursively too")
	flagSet.BoolVar(&m.tests, "tests", false, "search test files too (and direct test deps, with -r)")

	var cmds []exprCmd
	flagSet.Var(&strCmdFlag{
		name: "x",
		cmds: &cmds,
	}, "x", "")
	flagSet.Var(&strCmdFlag{
		name: "g",
		cmds: &cmds,
	}, "g", "")
	flagSet.Var(&strCmdFlag{
		name: "v",
		cmds: &cmds,
	}, "v", "")
	flagSet.Var(&strCmdFlag{
		name: "a",
		cmds: &cmds,
	}, "a", "")
	flagSet.Var(&strCmdFlag{
		name: "s",
		cmds: &cmds,
	}, "s", "")
	flagSet.Var(&strCmdFlag{
		name: "p",
		cmds: &cmds,
	}, "p", "")
	flagSet.Var(&boolCmdFlag{
		name: "w",
		cmds: &cmds,
	}, "w", "")
	flagSet.Parse(args)
	paths := flagSet.Args()

	if len(cmds) < 1 {
		return nil, nil, fmt.Errorf("need at least one command")
	}
	for i, cmd := range cmds {
		switch cmd.name {
		case "w":
			continue // no expr
		case "p":
			n, err := strconv.Atoi(cmd.src)
			if err != nil {
				return nil, nil, err
			}
			cmds[i].value = n
		case "a":
			m, err := m.parseAttrs(cmd.src)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot parse mods: %v", err)
			}
			cmds[i].value = m
		default:
			node, err := m.parseExpr(cmd.src)
			if err != nil {
				return nil, nil, err
			}
			cmds[i].value = node
		}
	}
	return cmds, paths, nil
}

type bufferJoinLines struct {
	bytes.Buffer
	last string
}

var rxNeedSemicolon = regexp.MustCompile(`([])}a-zA-Z0-9"'` + "`" + `]|\+\+|--)$`)

func (b *bufferJoinLines) Write(p []byte) (n int, err error) {
	if string(p) == "\n" {
		if b.last == "\n" {
			return 1, nil
		}
		if rxNeedSemicolon.MatchString(b.last) {
			b.Buffer.WriteByte(';')
		}
		b.Buffer.WriteByte(' ')
		b.last = "\n"
		return 1, nil
	}
	p = bytes.Trim(p, "\t")
	n, err = b.Buffer.Write(p)
	b.last = string(p)
	return
}

func (b *bufferJoinLines) String() string {
	return strings.TrimSuffix(b.Buffer.String(), "; ")
}

// inspect is like ast.Inspect, but it supports our extra nodeList Node
// type (only at the top level).
func inspect(node ast.Node, fn func(ast.Node) bool) {
	// ast.Walk barfs on ast.Node types it doesn't know, so
	// do the first level manually here
	list, ok := node.(nodeList)
	if !ok {
		ast.Inspect(node, fn)
		return
	}
	if !fn(list) {
		return
	}
	for i := 0; i < list.len(); i++ {
		ast.Inspect(list.at(i), fn)
	}
	fn(nil)
}

var emptyFset = token.NewFileSet()

func singleLinePrint(node ast.Node) string {
	var buf bufferJoinLines
	inspect(node, func(node ast.Node) bool {
		bl, ok := node.(*ast.BasicLit)
		if !ok || bl.Kind != token.STRING {
			return true
		}
		if !strings.HasPrefix(bl.Value, "`") {
			return true
		}
		if !strings.Contains(bl.Value, "\n") {
			return true
		}
		bl.Value = strconv.Quote(bl.Value[1 : len(bl.Value)-1])
		return true
	})
	printNode(&buf, emptyFset, node)
	return buf.String()
}

func printNode(w io.Writer, fset *token.FileSet, node ast.Node) {
	switch x := node.(type) {
	case exprList:
		if len(x) == 0 {
			return
		}
		printNode(w, fset, x[0])
		for _, n := range x[1:] {
			fmt.Fprintf(w, ", ")
			printNode(w, fset, n)
		}
	case stmtList:
		if len(x) == 0 {
			return
		}
		printNode(w, fset, x[0])
		for _, n := range x[1:] {
			fmt.Fprintf(w, "; ")
			printNode(w, fset, n)
		}
	default:
		err := printer.Fprint(w, fset, node)
		if err != nil && strings.Contains(err.Error(), "go/printer: unsupported node type") {
			// Should never happen, but make it obvious when it does.
			panic(fmt.Errorf("cannot print node %T: %v", node, err))
		}
	}
}
