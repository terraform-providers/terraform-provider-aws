// Package R006 defines an Analyzer that checks for
// RetryFunc that omit retryable errors
package R006

import (
	"flag"
	"go/ast"
	"go/types"
	"strings"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/resource/retryfuncinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for RetryFunc that omit retryable errors

The R006 analyzer reports when RetryFunc declarations are missing
retryable errors and should not be used as RetryFunc.

Optional parameters:
  - package-aliases Comma-separated list of additional Go import paths to consider as aliases for helper/resource, defaults to none.
`

const analyzerName = "R006"

var (
	packageAliases string
)

func parseFlags() flag.FlagSet {
	var flags = flag.NewFlagSet(analyzerName, flag.ExitOnError)
	flags.StringVar(&packageAliases, "package-aliases", "", "Comma-separated list of additional Go import paths to consider as aliases for helper/resource")
	return *flags
}

var Analyzer = &analysis.Analyzer{
	Name:  analyzerName,
	Doc:   Doc,
	Flags: parseFlags(),
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		retryfuncinfo.Analyzer,
	},
	Run: run,
}

func isPackageAliasIgnored(e ast.Expr, info *types.Info, packageAliasesList string) bool {
	packageAliases := strings.Split(packageAliasesList, ",")

	for _, packageAlias := range packageAliases {
		if astutils.IsModulePackageFunc(e, info, packageAlias, "", resource.FuncNameRetryableError) {
			return true
		}
	}

	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	retryFuncs := pass.ResultOf[retryfuncinfo.Analyzer].([]*resource.RetryFuncInfo)

	for _, retryFunc := range retryFuncs {
		if ignorer.ShouldIgnore(analyzerName, retryFunc.Node) {
			continue
		}

		var retryableErrorFound bool

		ast.Inspect(retryFunc.Body, func(n ast.Node) bool {
			callExpr, ok := n.(*ast.CallExpr)

			if !ok {
				return true
			}

			if resource.IsFunc(callExpr.Fun, pass.TypesInfo, resource.FuncNameRetryableError) {
				retryableErrorFound = true
				return false
			}

			if packageAliases != "" && isPackageAliasIgnored(callExpr.Fun, pass.TypesInfo, packageAliases) {
				retryableErrorFound = true
				return false
			}

			return true
		})

		if !retryableErrorFound {
			pass.Reportf(retryFunc.Pos, "%s: RetryFunc should include RetryableError() handling or be removed", analyzerName)
		}
	}

	return nil, nil
}
