package analysisutils

import (
	"fmt"
	"go/ast"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// StdlibFunctionCallExprAnalyzer returns an Analyzer for standard library function *ast.CallExpr
func StdlibFunctionCallExprAnalyzer(analyzerName string, packagePath string, functionName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find %s.%s() calls for later passes", packagePath, functionName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        StdlibFunctionCallExprRunner(packagePath, functionName),
		ResultType: reflect.TypeOf([]*ast.CallExpr{}),
	}
}

// StdlibFunctionSelectorExprAnalyzer returns an Analyzer for standard library function *ast.SelectorExpr
func StdlibFunctionSelectorExprAnalyzer(analyzerName string, packagePath string, functionName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find %s.%s() selectors for later passes", packagePath, functionName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        StdlibFunctionSelectorExprRunner(packagePath, functionName),
		ResultType: reflect.TypeOf([]*ast.SelectorExpr{}),
	}
}
