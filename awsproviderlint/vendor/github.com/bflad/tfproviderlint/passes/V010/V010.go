package V010

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/validation/stringdoesnotmatchcallexpr"
)

const Doc = `check for validation.StringDoesNotMatch() calls with empty message argument

The V010 analyzer reports when the second argument for a validation.StringDoesNotMatch()
call is an empty string. It is preferred to provide a friendly validation
message, rather than allowing the function to return the raw regular expression
as the message, since not all practitioners may be familiar with regular
expression syntax.`

const analyzerName = "V010"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		stringdoesnotmatchcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	sets := pass.ResultOf[stringdoesnotmatchcallexpr.Analyzer].([]*ast.CallExpr)
	for _, set := range sets {
		if ignorer.ShouldIgnore(analyzerName, set) {
			continue
		}

		if len(set.Args) < 2 {
			continue
		}

		switch v := set.Args[1].(type) {
		default:
			continue
		case *ast.BasicLit:
			if value := astutils.ExprStringValue(v); value != nil && *value == "" {
				pass.Reportf(v.Pos(), "%s: validation.StringDoesNotMatch() message argument should be non-empty", analyzerName)
			}
		}
	}

	return nil, nil
}
