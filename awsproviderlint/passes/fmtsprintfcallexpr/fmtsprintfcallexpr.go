package fmtsprintfcallexpr

import (
	"go/ast"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "fmtsprintfcallexpr",
	Doc:  "find fmt.Sprintf() *ast.CallExpr for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*ast.CallExpr{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	var result []*ast.CallExpr

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		callExpr := n.(*ast.CallExpr)

		switch fun := callExpr.Fun.(type) {
		case *ast.SelectorExpr:
			if fun.Sel.Name != "Sprintf" {
				return
			}

			switch x := fun.X.(type) {
			case *ast.Ident:
				if pass.TypesInfo.ObjectOf(x).(*types.PkgName).Imported().Path() != "fmt" {
					return
				}
			default:
				return
			}
		default:
			return
		}

		result = append(result, callExpr)
	})

	return result, nil
}
