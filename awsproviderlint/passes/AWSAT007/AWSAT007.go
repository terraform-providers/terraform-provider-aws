// Package AWSAT007 defines an Analyzer that checks for
// hardcoded AWS partition DNS suffixes
package AWSAT007

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for hardcoded instance types

The AWSAT007 analyzer reports hardcoded instance types. For tests
to work across AWS partitions, instance types should not be
hardcoded.
`

const analyzerName = "AWSAT007"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		inspect.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BasicLit)(nil),
	}

	sizes := []string{"nano", "micro", "small", "medium", "large", "metal"}
	attribute_suffixes := []string{"type", "class"}

	re := regexp.MustCompile(`(` + strings.Join(attribute_suffixes, "|") + `)\s*=\s*"[a-z0-9.]{2,15}(` + strings.Join(sizes, "|") + `)`)
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.BasicLit)

		if ignorer.ShouldIgnore(analyzerName, x) {
			return
		}

		if x.Kind != token.STRING {
			return
		}

		if !re.MatchString(x.Value) {
			return
		}

		pass.Reportf(x.ValuePos, "%s: avoid hardcoding instance type, use data source aws_ec2_instance_type_offering", analyzerName)
	})
	return nil, nil
}
