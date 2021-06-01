package resourceinfo

import (
	"go/ast"
	"reflect"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "resourceinfo",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/schema.Resource literals for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*schema.ResourceInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*schema.ResourceInfo

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !isSchemaResource(pass, x) {
			return
		}

		result = append(result, schema.NewResourceInfo(x, pass.TypesInfo))
	})

	return result, nil
}

func isSchemaResource(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.SelectorExpr:
		return schema.IsTypeResource(pass.TypesInfo.TypeOf(v))
	}
}
