package analysisutils

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// StdlibFunctionCallExprRunner returns an Analyzer runner for function *ast.CallExpr
func StdlibFunctionCallExprRunner(packagePath string, functionName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.CallExpr)(nil),
		}
		var result []*ast.CallExpr

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			callExpr := n.(*ast.CallExpr)

			if !astutils.IsStdlibPackageFunc(callExpr.Fun, pass.TypesInfo, packagePath, functionName) {
				return
			}

			result = append(result, callExpr)
		})

		return result, nil
	}
}

// StdlibFunctionSelectorExprRunner returns an Analyzer runner for function *ast.SelectorExpr
func StdlibFunctionSelectorExprRunner(packagePath string, functionName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.SelectorExpr)(nil),
		}
		var result []*ast.SelectorExpr

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			selectorExpr := n.(*ast.SelectorExpr)

			if !astutils.IsStdlibPackageFunc(selectorExpr, pass.TypesInfo, packagePath, functionName) {
				return
			}

			result = append(result, selectorExpr)
		})

		return result, nil
	}
}
