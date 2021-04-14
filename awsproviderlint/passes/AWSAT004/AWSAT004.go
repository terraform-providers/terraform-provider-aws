// Package AWSAT004 defines an Analyzer that checks for
// TestCheckResourceAttr() calls with hardcoded TypeSet state hashes
package AWSAT004

import (
	"go/ast"
	"regexp"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/resource/testcheckresourceattrcallexpr"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for TestCheckResourceAttr() calls with hardcoded TypeSet state hashes

The AWSAT004 analyzer reports TestCheckResourceAttr() calls with hardcoded
TypeSet state hashes. Hardcoded state hashes are an unreliable way to
specifically address state values since hashes may change over time, be
inconsistent across partitions, and can be inadvertently changed by modifying
configurations.
`

const analyzerName = "AWSAT004"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testcheckresourceattrcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	callExprs := pass.ResultOf[testcheckresourceattrcallexpr.Analyzer].([]*ast.CallExpr)

	re := regexp.MustCompile(`[a-z0-9_]+\.\d{7,20}`)
	for _, callExpr := range callExprs {
		if ignorer.ShouldIgnore(analyzerName, callExpr) {
			continue
		}

		attributeName := astutils.ExprStringValue(callExpr.Args[1])

		if attributeName == nil {
			continue
		}

		if !re.MatchString(*attributeName) {
			continue
		}

		pass.Reportf(callExpr.Args[1].Pos(), "%s: avoid hardcoded state hashes, instead use the TestCheckTypeSetElemNestedAttrs function", analyzerName)

	}
	return nil, nil
}
