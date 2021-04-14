// Package V001 defines an Analyzer that checks for
// custom SchemaValidateFunc that implement validation.StringMatch()
package V001

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemavalidatefuncinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for custom SchemaValidateFunc that implement validation.StringMatch()

The V001 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with validation.StringMatch() or validation.StringDoesNotMatch().`

const analyzerName = "V001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		schemavalidatefuncinfo.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemaValidateFuncs := pass.ResultOf[schemavalidatefuncinfo.Analyzer].([]*schema.SchemaValidateFuncInfo)

	for _, schemaValidateFunc := range schemaValidateFuncs {
		if ignorer.ShouldIgnore(analyzerName, schemaValidateFunc.Node) {
			continue
		}

		ast.Inspect(schemaValidateFunc.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			if !astutils.IsPackageReceiverMethod(callExpr.Fun, pass.TypesInfo, "regexp", "Regexp", "MatchString") {
				return true
			}

			pass.Reportf(schemaValidateFunc.Pos, "%s: custom SchemaValidateFunc should be replaced with validation.StringMatch() or validation.StringDoesNotMatch()", analyzerName)
			return false
		})
	}

	return nil, nil
}
