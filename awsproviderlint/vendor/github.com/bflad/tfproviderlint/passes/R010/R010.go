package R010

import (
	"go/ast"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/resourcedatagetchangeassignstmt"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for (schema.ResourceData).GetChange() usage that should prefer (schema.ResourceData).Get()

The R010 analyzer reports when (schema.ResourceData).GetChange() assignments
are not using the first return value (assigned to _), which should be
replaced with (schema.ResourceData).Get() instead.`

const analyzerName = "R010"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		resourcedatagetchangeassignstmt.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	assignStmts := pass.ResultOf[resourcedatagetchangeassignstmt.Analyzer].([]*ast.AssignStmt)

	for _, assignStmt := range assignStmts {
		if ignorer.ShouldIgnore(analyzerName, assignStmt) {
			continue
		}

		ident, ok := assignStmt.Lhs[0].(*ast.Ident)

		if !ok || ident.Name != "_" {
			continue
		}

		pass.Reportf(assignStmt.Pos(), "%s: prefer d.Get() over d.GetChange() when only using second return value", analyzerName)
	}

	return nil, nil
}
