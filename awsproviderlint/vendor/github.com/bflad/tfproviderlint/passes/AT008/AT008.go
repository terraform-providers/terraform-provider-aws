package AT008

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/testaccfuncdecl"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for acceptance test function declaration *testing.T parameter naming

The AT008 analyzer reports where the *testing.T parameter of an acceptance test
declaration is not named t, which is a standard convention.`

const analyzerName = "AT008"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testaccfuncdecl.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	funcDecls := pass.ResultOf[testaccfuncdecl.Analyzer].([]*ast.FuncDecl)
	for _, funcDecl := range funcDecls {
		if ignorer.ShouldIgnore(analyzerName, funcDecl) {
			continue
		}

		params := funcDecl.Type.Params

		if params == nil || len(params.List) != 1 {
			continue
		}

		firstParam := params.List[0]

		if firstParam == nil || len(firstParam.Names) != 1 {
			continue
		}

		name := firstParam.Names[0]

		if name == nil || name.Name == "t" {
			continue
		}

		pass.Reportf(name.Pos(), "%s: acceptance test function declaration *testing.T parameter should be named t", analyzerName)
	}

	return nil, nil
}
