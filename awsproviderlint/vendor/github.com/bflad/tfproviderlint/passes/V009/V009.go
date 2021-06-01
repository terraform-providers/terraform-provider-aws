package V009

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/validation/stringmatchcallexpr"
)

const Doc = `check for validation.StringMatch() calls with empty message argument

The V009 analyzer reports when the second argument for a validation.StringMatch()
call is an empty string. It is preferred to provide a friendly validation
message, rather than allowing the function to return the raw regular expression
as the message, since not all practitioners may be familiar with regular
expression syntax.`

const analyzerName = "V009"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		stringmatchcallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	sets := pass.ResultOf[stringmatchcallexpr.Analyzer].([]*ast.CallExpr)
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
				pass.Reportf(v.Pos(), "%s: validation.StringMatch() message argument should be non-empty", analyzerName)
			}
		}
	}

	return nil, nil
}
