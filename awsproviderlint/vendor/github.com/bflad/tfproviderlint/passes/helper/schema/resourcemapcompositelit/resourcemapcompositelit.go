package resourcemapcompositelit

import (
	"go/ast"
	"reflect"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "resourcemapcompositelit",
	Doc:  "find map[string]*schema.Resource literals for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*ast.CompositeLit{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*ast.CompositeLit

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !schema.IsMapStringResource(x, pass.TypesInfo) {
			return
		}

		result = append(result, x)
	})

	return result, nil
}
