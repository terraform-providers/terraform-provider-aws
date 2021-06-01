// Package XR004 defines an Analyzer that checks for
// ResourceData.Set() calls missing error checking with
// complex types
package XR004

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for ResourceData.Set() calls missing error checking with complex values

The XR004 analyzer reports Set() calls that receive a complex value type, but
does not perform error checking. This error checking is to prevent issues where
the code is not able to properly set the Terraform state for drift detection.

Reference: https://www.terraform.io/docs/extend/best-practices/detecting-drift.html#error-checking-aggregate-types`

const analyzerName = "XR004"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.ExprStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		exprStmt := n.(*ast.ExprStmt)

		callExpr, ok := exprStmt.X.(*ast.CallExpr)

		if !ok {
			return
		}

		if !schema.IsReceiverMethod(callExpr.Fun, pass.TypesInfo, schema.TypeNameResourceData, "Set") {
			return
		}

		if ignorer.ShouldIgnore(analyzerName, callExpr) {
			return
		}

		if len(callExpr.Args) < 2 {
			return
		}

		if isBasicType(pass.TypesInfo.TypeOf(callExpr.Args[1]).Underlying()) {
			return
		}

		pass.Reportf(callExpr.Pos(), "%s: ResourceData.Set() should perform error checking with complex values", analyzerName)
	})

	return nil, nil
}

func isBasicType(t types.Type) bool {
	switch t := t.(type) {
	case *types.Basic:
		return isAllowedBasicType(t)
	case *types.Pointer:
		return isBasicType(t.Elem())
	}

	return false
}

var allowedBasicKindTypes = []types.BasicKind{
	types.Bool,
	types.Float32,
	types.Float64,
	types.Int,
	types.Int8,
	types.Int16,
	types.Int32,
	types.Int64,
	types.String,
	types.UntypedNil,
}

func isAllowedBasicType(b *types.Basic) bool {
	for _, allowedBasicKindType := range allowedBasicKindTypes {
		if b.Kind() == allowedBasicKindType {
			return true
		}
	}

	return false
}
