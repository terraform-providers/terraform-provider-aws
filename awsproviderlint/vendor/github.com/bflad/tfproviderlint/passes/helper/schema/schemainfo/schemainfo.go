package schemainfo

import (
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemamapcompositelit"
)

var Analyzer = &analysis.Analyzer{
	Name: "schemainfo",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/schema.Schema literals for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		schemamapcompositelit.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*schema.SchemaInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	schemamapcompositelits := pass.ResultOf[schemamapcompositelit.Analyzer].([]*ast.CompositeLit)
	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}
	var result []*schema.SchemaInfo

	for _, smap := range schemamapcompositelits {
		for _, mapSchema := range schema.GetSchemaMapSchemas(smap) {
			result = append(result, schema.NewSchemaInfo(mapSchema, pass.TypesInfo))
		}
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.CompositeLit)

		if !isSchemaSchema(pass, x) {
			return
		}

		result = append(result, schema.NewSchemaInfo(x, pass.TypesInfo))
	})

	return result, nil
}

func isSchemaSchema(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	switch v := cl.Type.(type) {
	default:
		return false
	case *ast.SelectorExpr:
		return schema.IsTypeSchema(pass.TypesInfo.TypeOf(v))
	}
}
