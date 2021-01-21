// Package AT003 defines an Analyzer that checks for
// acceptance test names missing an underscore
package AT003

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testaccfuncdecl"
)

const Doc = `check for acceptance test function names missing an underscore

The AT003 analyzer reports where an underscore is not
present in the function name, which could make per-resource testing harder to
execute in larger providers or those with overlapping resource names.`

const analyzerName = "AT003"

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

		if !strings.Contains(testAccFunc.Name.Name, "_") {
			pass.Reportf(testAccFunc.Name.NamePos, "%s: acceptance test function name should include underscore", analyzerName)
		}
	}

	return nil, nil
}
