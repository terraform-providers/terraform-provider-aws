package AWSR002

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatasetcallexpr"
	"github.com/terraform-providers/terraform-provider-aws/awsproviderlint/helper/awsprovidertype/keyvaluetags"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for d.Set() of tags attribute that should include IgnoreConfig()

The AWSR002 analyzer reports when a (schema.ResourceData).Set() call with the
tags key is missing a call to (keyvaluetags.KeyValueTags).IgnoreConfig() in the
value, which ensures any provider level ignore tags configuration is applied.
`

const analyzerName = "AWSR002"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourcedatasetcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	callExprs := pass.ResultOf[resourcedatasetcallexpr.Analyzer].([]*ast.CallExpr)
	commentIgnorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

	for _, callExpr := range callExprs {
		if commentIgnorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		if len(callExpr.Args) < 2 {
			continue
		}

		attributeName := astutils.ExprStringValue(callExpr.Args[0])

		if attributeName == nil || *attributeName != "tags" {
			continue
		}

		var ignoreConfigCallExprFound bool

		ast.Inspect(callExpr.Args[1], func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			if keyvaluetags.IsReceiverMethod(callExpr.Fun, pass.TypesInfo, keyvaluetags.TypeNameKeyValueTags, keyvaluetags.KeyValueTagsMethodNameIgnoreConfig) {
				ignoreConfigCallExprFound = true
				return false
			}

			return true
		})

		if !ignoreConfigCallExprFound {
			pass.Reportf(callExpr.Args[1].Pos(), "%s: missing (keyvaluetags.KeyValueTags).IgnoreConfig()", analyzerName)
		}

	}

	return nil, nil
}
