package XAT001

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/resource/testcaseinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for TestCase missing ErrorCheck

The XAT001 analyzer reports uses of TestCase which do not define an ErrorCheck
function. ErrorCheck can be used to skip tests for known environmental issues.`

const analyzerName = "XAT001"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testcaseinfo.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	testCases := pass.ResultOf[testcaseinfo.Analyzer].([]*resource.TestCaseInfo)

	for _, testCase := range testCases {
		if ignorer.ShouldIgnore(analyzerName, testCase.AstCompositeLit) {
			continue
		}

		if testCase.DeclaresField(resource.TestCaseFieldErrorCheck) {
			continue
		}

		pass.Reportf(testCase.AstCompositeLit.Type.(*ast.SelectorExpr).Sel.Pos(), "%s: missing ErrorCheck", analyzerName)
	}

	return nil, nil
}
