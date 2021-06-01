// Package AT002 defines an Analyzer that checks for
// acceptance test names including the word import
package AT002

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testaccfuncdecl"
)

const Doc = `check for acceptance test function names including the word import

The AT002 analyzer reports where the word import or Import is used
in an acceptance test function name, which generally means there is an extraneous
acceptance test. ImportState testing should be included as a TestStep with each
applicable acceptance test, rather than a separate test that only verifies import
of a single test configuration.`

const analyzerName = "AT002"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		testaccfuncdecl.Analyzer,
		commentignore.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	testAccFuncs := pass.ResultOf[testaccfuncdecl.Analyzer].([]*ast.FuncDecl)
	for _, testAccFunc := range testAccFuncs {
		if ignorer.ShouldIgnore(analyzerName, testAccFunc) {
			continue
		}

		funcName := testAccFunc.Name.Name

		if strings.Contains(funcName, "_import") || strings.Contains(funcName, "_Import") {
			pass.Reportf(testAccFunc.Name.NamePos, "%s: acceptance test function name should not include import", analyzerName)
		}
	}

	return nil, nil
}
