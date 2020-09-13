package ruleguard

import (
	"go/ast"
	"go/printer"
	"io/ioutil"
	"strings"

	"github.com/quasilyte/go-ruleguard/internal/mvdan.cc/gogrep"
)

type rulesRunner struct {
	ctx   *Context
	rules *GoRuleSet

	filename string
	src      []byte
}

func newRulesRunner(ctx *Context, rules *GoRuleSet) *rulesRunner {
	return &rulesRunner{
		ctx:   ctx,
		rules: rules,
	}
}

func (rr *rulesRunner) fileBytes() []byte {
	if rr.src != nil {
		return rr.src
	}

	// TODO(quasilyte): re-use src slice?
	src, err := ioutil.ReadFile(rr.filename)
	if err != nil || src == nil {
		// Assign a zero-length slice so rr.src
		// is never nil during the second fileBytes call.
		rr.src = make([]byte, 0)
	} else {
		rr.src = src
	}
	return rr.src
}

func (rr *rulesRunner) run(f *ast.File) error {
	// TODO(quasilyte): run local rules as well.

	rr.filename = rr.ctx.Fset.Position(f.Pos()).Filename

	for _, rule := range rr.rules.universal.uncategorized {
		rule.pat.Match(f, func(m gogrep.MatchData) {
			rr.handleMatch(rule, m)
		})
	}

	if rr.rules.universal.categorizedNum != 0 {
		ast.Inspect(f, func(n ast.Node) bool {
			cat := categorizeNode(n)
			for _, rule := range rr.rules.universal.rulesByCategory[cat] {
				matched := false
				rule.pat.MatchNode(n, func(m gogrep.MatchData) {
					matched = rr.handleMatch(rule, m)
				})
				if matched {
					break
				}
			}
			return true
		})
	}

	return nil
}

func (rr *rulesRunner) handleMatch(rule goRule, m gogrep.MatchData) bool {
	for name, node := range m.Values {
		expr, ok := node.(ast.Expr)
		if !ok {
			continue
		}
		filter, ok := rule.filters[name]
		if !ok {
			continue
		}
		if filter.typePred != nil {
			typ := rr.ctx.Types.TypeOf(expr)
			q := typeQuery{x: typ, ctx: rr.ctx}
			if !filter.typePred(q) {
				return false
			}
		}
		switch filter.addressable {
		case bool3true:
			if !isAddressable(rr.ctx.Types, expr) {
				return false
			}
		case bool3false:
			if isAddressable(rr.ctx.Types, expr) {
				return false
			}
		}
		switch filter.pure {
		case bool3true:
			if !isPure(rr.ctx.Types, expr) {
				return false
			}
		case bool3false:
			if isPure(rr.ctx.Types, expr) {
				return false
			}
		}
		switch filter.constant {
		case bool3true:
			if !isConstant(rr.ctx.Types, expr) {
				return false
			}
		case bool3false:
			if isConstant(rr.ctx.Types, expr) {
				return false
			}
		}
	}

	prefix := ""
	if rule.severity != "" {
		prefix = rule.severity + ": "
	}
	message := prefix + rr.renderMessage(rule.msg, m.Node, m.Values)
	node := m.Node
	if rule.location != "" {
		node = m.Values[rule.location]
	}
	var suggestion *Suggestion
	if rule.suggestion != "" {
		suggestion = &Suggestion{
			Replacement: []byte(rr.renderMessage(rule.suggestion, m.Node, m.Values)),
			From:        node.Pos(),
			To:          node.End(),
		}
	}
	rr.ctx.Report(node, message, suggestion)
	return true
}

func (rr *rulesRunner) renderMessage(msg string, n ast.Node, nodes map[string]ast.Node) string {
	var buf strings.Builder
	if strings.Contains(msg, "$$") {
		rr.writeNode(&buf, n)
		msg = strings.ReplaceAll(msg, "$$", buf.String())
	}
	if len(nodes) == 0 {
		return msg
	}
	for name, n := range nodes {
		key := "$" + name
		if !strings.Contains(msg, key) {
			continue
		}
		buf.Reset()
		rr.writeNode(&buf, n)
		// Don't interpolate strings that are too long.
		var replacement string
		if buf.Len() > 40 {
			replacement = key
		} else {
			replacement = buf.String()
		}
		msg = strings.ReplaceAll(msg, key, replacement)
	}
	return msg
}

func (rr *rulesRunner) writeNode(buf *strings.Builder, n ast.Node) {
	from := rr.ctx.Fset.Position(n.Pos()).Offset
	to := rr.ctx.Fset.Position(n.End()).Offset
	src := rr.fileBytes()
	if (from >= 0 && int(from) < len(src)) && (to >= 0 && int(to) < len(src)) {
		buf.Write(src[from:to])
		return
	}
	if err := printer.Fprint(buf, rr.ctx.Fset, n); err != nil {
		panic(err)
	}
}
