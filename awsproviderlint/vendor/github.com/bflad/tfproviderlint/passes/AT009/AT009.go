package AT009

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/acctest"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/acctest/randstringfromcharsetcallexpr"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for acctest.RandStringFromCharSet() calls that can be acctest.RandString()

The AT009 analyzer reports where the second parameter of a
RandStringFromCharSet call is acctest.CharSetAlpha, which is equivalent to
calling RandString.`

const analyzerName = "AT009"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		randstringfromcharsetcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	callExprs := pass.ResultOf[randstringfromcharsetcallexpr.Analyzer].([]*ast.CallExpr)
	for _, callExpr := range callExprs {
		if ignorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		if len(callExpr.Args) < 2 {
			continue
		}

		if !acctest.IsConst(callExpr.Args[1], pass.TypesInfo, acctest.ConstNameCharSetAlphaNum) {
			continue
		}

		pass.Reportf(callExpr.Pos(), "%s: should use RandString call instead", analyzerName)
	}

	return nil, nil
}
