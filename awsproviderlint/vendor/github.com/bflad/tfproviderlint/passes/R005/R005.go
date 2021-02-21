// Package R005 defines an Analyzer that checks for
// ResourceData.HasChange() calls that can be combined into
// a single HasChanges() call.
package R005

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for ResourceData.HasChange() calls that can be combined into a single HasChanges() call

The R005 analyzer reports when multiple HasChange() calls in a conditional
can be combined into a single HasChanges() call.`

const analyzerName = "R005"

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
		(*ast.BinaryExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		binaryExpr := n.(*ast.BinaryExpr)

		if ignorer.ShouldIgnore(analyzerName, n) {
			return
		}

		if binaryExpr.Op != token.LOR {
			return
		}

		if !isHasChangeCall(binaryExpr.X, pass.TypesInfo) {
			return
		}

		if !isHasChangeCall(binaryExpr.Y, pass.TypesInfo) {
			return
		}

		pass.Reportf(binaryExpr.Pos(), "%s: multiple ResourceData.HasChange() calls can be combined with single HasChanges() call", analyzerName)
	})

	return nil, nil
}

func isHasChangeCall(e ast.Expr, info *types.Info) bool {
	switch e := e.(type) {
	case *ast.CallExpr:
		return schema.IsReceiverMethod(e.Fun, info, schema.TypeNameResourceData, "HasChange")
	}

	return false
}
