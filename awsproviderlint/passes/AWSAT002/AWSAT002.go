// Package AWSAT002 defines an Analyzer that checks for
// hardcoded AMI IDs
package AWSAT002

import (
	"go/ast"
	"go/token"
	"regexp"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for hardcoded AMI IDs

The AWSAT002 analyzer reports hardcoded AMI IDs. AMI IDs are region dependent and tests will fail in any region or partition other than where the AMI was created.
`

const analyzerName = "AWSAT002"

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
	re := regexp.MustCompile("ami-[0-9a-z]{8,17}")
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

		pass.Reportf(x.ValuePos, "%s: AMI IDs should not be hardcoded", analyzerName)
	})
	return nil, nil
}
