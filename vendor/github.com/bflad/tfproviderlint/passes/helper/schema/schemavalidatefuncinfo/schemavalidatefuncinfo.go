package schemavalidatefuncinfo

import (
	"go/ast"
	"reflect"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "schemavalidatefuncinfo",
	Doc:  "find github.com/hashicorp/terraform-plugin-sdk/helper/schema SchemaValidateFunc declarations for later passes",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*schema.SchemaValidateFuncInfo{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}
	var result []*schema.SchemaValidateFuncInfo

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		funcType := astutils.FuncTypeFromNode(n)

		if funcType == nil {
			return
		}

		if !astutils.IsFieldListType(funcType.Params, 0, astutils.IsExprTypeInterface) {
			return
		}

		if !astutils.IsFieldListType(funcType.Params, 1, astutils.IsExprTypeString) {
			return
		}

		if !astutils.IsFieldListType(funcType.Results, 0, astutils.IsExprTypeArrayString) {
			return
		}

		if !astutils.IsFieldListType(funcType.Results, 1, astutils.IsExprTypeArrayError) {
			return
		}

		result = append(result, schema.NewSchemaValidateFuncInfo(n, pass.TypesInfo))
	})

	return result, nil
}
