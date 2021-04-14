package R009

import (
	"go/ast"

	"github.com/bflad/gopaniccheck/passes/logpaniccallexpr"
	"github.com/bflad/gopaniccheck/passes/logpanicfcallexpr"
	"github.com/bflad/gopaniccheck/passes/logpaniclncallexpr"
	"github.com/bflad/gopaniccheck/passes/paniccallexpr"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for Go panic usage

The R009 analyzer reports usage of Go panics, which should be avoided.
Any errors should be surfaced to Terraform, which will display them in the
user interface and ensures any necessary state actions (e.g. cleanup) are
performed as expected.`

const analyzerName = "R009"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		logpaniccallexpr.Analyzer,
		logpanicfcallexpr.Analyzer,
		logpaniclncallexpr.Analyzer,
		paniccallexpr.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	logPanicCallExprs := pass.ResultOf[logpaniccallexpr.Analyzer].([]*ast.CallExpr)
	logPanicfCallExprs := pass.ResultOf[logpanicfcallexpr.Analyzer].([]*ast.CallExpr)
	logPaniclnCallExprs := pass.ResultOf[logpaniclncallexpr.Analyzer].([]*ast.CallExpr)
	panicCallExprs := pass.ResultOf[paniccallexpr.Analyzer].([]*ast.CallExpr)

	for _, logPanicCallExpr := range logPanicCallExprs {
		if ignorer.ShouldIgnore(analyzerName, logPanicCallExpr) {
			continue
		}

		pass.Reportf(logPanicCallExpr.Pos(), "%s: avoid log.Panic() usage", analyzerName)
	}

	for _, logPanicfCallExpr := range logPanicfCallExprs {
		if ignorer.ShouldIgnore(analyzerName, logPanicfCallExpr) {
			continue
		}

		pass.Reportf(logPanicfCallExpr.Pos(), "%s: avoid log.Panicf() usage", analyzerName)
	}

	for _, logPaniclnCallExpr := range logPaniclnCallExprs {
		if ignorer.ShouldIgnore(analyzerName, logPaniclnCallExpr) {
			continue
		}

		pass.Reportf(logPaniclnCallExpr.Pos(), "%s: avoid log.Panicln() usage", analyzerName)
	}

	for _, panicCallExpr := range panicCallExprs {
		if ignorer.ShouldIgnore(analyzerName, panicCallExpr) {
			continue
		}

		pass.Reportf(panicCallExpr.Pos(), "%s: avoid panic() usage", analyzerName)
	}

	return nil, nil
}
