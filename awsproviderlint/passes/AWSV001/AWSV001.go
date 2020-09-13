package AWSV001

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/validation/stringinslicecallexpr"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for validation.StringInSlice() using []string parameter

The AWSV001 analyzer reports when a validation.StringInSlice() call has the
first parameter of a []string, which suggests either that AWS API model
constants are not available or that the usage is prior to the AWS Go SDK adding
functions that return all values for the enumeration type.

If the API model constants are not available, this check can be ignored but it
is recommended to submit an AWS Support case to the AWS service team for adding
the constants.

If the hardcoded strings are AWS Go SDK constants, this check reports when the
first parameter should be switched to the newer ENUM_Values() function.
`

const analyzerName = "AWSV001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		stringinslicecallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	callExprs := pass.ResultOf[stringinslicecallexpr.Analyzer].([]*ast.CallExpr)
	commentIgnorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

	for _, callExpr := range callExprs {
		if commentIgnorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		if len(callExpr.Args) < 2 {
			continue
		}

		ast.Inspect(callExpr.Args[0], func(n ast.Node) bool {
			compositeLit, ok := n.(*ast.CompositeLit)

			if !ok {
				return true
			}

			if astutils.IsExprTypeArrayString(compositeLit.Type) {
				pass.Reportf(callExpr.Args[0].Pos(), "%s: prefer AWS Go SDK ENUM_Values() function (ignore if not applicable)", analyzerName)
				return false
			}

			return true
		})
	}

	return nil, nil
}
