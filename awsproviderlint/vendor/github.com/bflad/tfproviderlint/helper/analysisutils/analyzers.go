package analysisutils

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// AvoidSelectorExprAnalyzer returns an Analyzer for *ast.SelectorExpr to avoid
func AvoidSelectorExprAnalyzer(analyzerName string, callExprAnalyzer, selectorExprAnalyzer *analysis.Analyzer, packagePath, typeName string) *analysis.Analyzer {
	doc := fmt.Sprintf(`check for %[2]s.%[3]s usage to avoid

The %[1]s analyzer reports usage:

%[2]s.%[3]s
`, analyzerName, packagePath, typeName)

	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  doc,
		Requires: []*analysis.Analyzer{
			callExprAnalyzer,
			commentignore.Analyzer,
			selectorExprAnalyzer,
		},
		Run: AvoidSelectorExprRunner(analyzerName, callExprAnalyzer, selectorExprAnalyzer, packagePath, typeName),
	}
}

// DeprecatedReceiverMethodSelectorExprAnalyzer returns an Analyzer for deprecated *ast.SelectorExpr
func DeprecatedReceiverMethodSelectorExprAnalyzer(analyzerName string, callExprAnalyzer, selectorExprAnalyzer *analysis.Analyzer, packagePath, typeName, methodName string) *analysis.Analyzer {
	doc := fmt.Sprintf(`check for deprecated %[2]s.%[3]s usage

The %[1]s analyzer reports usage of the deprecated:

%[2]s.%[3]s
`, analyzerName, packagePath, typeName, methodName)

	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  doc,
		Requires: []*analysis.Analyzer{
			callExprAnalyzer,
			commentignore.Analyzer,
			selectorExprAnalyzer,
		},
		Run: DeprecatedReceiverMethodSelectorExprRunner(analyzerName, callExprAnalyzer, selectorExprAnalyzer, packagePath, typeName, methodName),
	}
}

// DeprecatedEmptyCallExprWithReplacementSelectorExprAnalyzer returns an Analyzer for deprecated *ast.SelectorExpr with replacement
func DeprecatedEmptyCallExprWithReplacementSelectorExprAnalyzer(analyzerName string, callExprAnalyzer, selectorExprAnalyzer *analysis.Analyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName string) *analysis.Analyzer {
	doc := fmt.Sprintf(`check for deprecated %[2]s.%[3]s usage

The %[1]s analyzer reports usage of the deprecated:

%[2]s.%[3]s

That should be replaced with:

%[4]s.%[5]s
`, analyzerName, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName)

	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  doc,
		Requires: []*analysis.Analyzer{
			callExprAnalyzer,
			commentignore.Analyzer,
			selectorExprAnalyzer,
		},
		Run: DeprecatedEmptyCallExprWithReplacementSelectorExprRunner(analyzerName, callExprAnalyzer, selectorExprAnalyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName),
	}
}

// DeprecatedWithReplacementPointerSelectorExprAnalyzer returns an Analyzer for deprecated *ast.SelectorExpr with replacement
func DeprecatedWithReplacementPointerSelectorExprAnalyzer(analyzerName string, selectorExprAnalyzer *analysis.Analyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName string) *analysis.Analyzer {
	doc := fmt.Sprintf(`check for deprecated %[2]s.%[3]s usage

The %[1]s analyzer reports usage of the deprecated:

%[2]s.%[3]s

That should be replaced with:

*%[4]s.%[5]s
`, analyzerName, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName)

	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  doc,
		Requires: []*analysis.Analyzer{
			commentignore.Analyzer,
			selectorExprAnalyzer,
		},
		Run: DeprecatedWithReplacementPointerSelectorExprRunner(analyzerName, selectorExprAnalyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName),
	}
}

// DeprecatedWithReplacementSelectorExprAnalyzer returns an Analyzer for deprecated *ast.SelectorExpr with replacement
func DeprecatedWithReplacementSelectorExprAnalyzer(analyzerName string, selectorExprAnalyzer *analysis.Analyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName string) *analysis.Analyzer {
	doc := fmt.Sprintf(`check for deprecated %[2]s.%[3]s usage

The %[1]s analyzer reports usage of the deprecated:

%[2]s.%[3]s

That should be replaced with:

%[4]s.%[5]s
`, analyzerName, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName)

	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  doc,
		Requires: []*analysis.Analyzer{
			commentignore.Analyzer,
			selectorExprAnalyzer,
		},
		Run: DeprecatedWithReplacementSelectorExprRunner(analyzerName, selectorExprAnalyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName),
	}
}

// FunctionCallExprAnalyzer returns an Analyzer for function *ast.CallExpr
func FunctionCallExprAnalyzer(analyzerName string, packageFunc func(ast.Expr, *types.Info, string) bool, packagePath string, functionName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find %s.%s calls for later passes", packagePath, functionName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        FunctionCallExprRunner(packageFunc, functionName),
		ResultType: reflect.TypeOf([]*ast.CallExpr{}),
	}
}

// ReceiverMethodAssignStmtAnalyzer returns an Analyzer for receiver method *ast.AssignStmt
func ReceiverMethodAssignStmtAnalyzer(analyzerName string, packageReceiverMethodFunc func(ast.Expr, *types.Info, string, string) bool, packagePath string, receiverName string, methodName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find (%s.%s).%s assignments for later passes", packagePath, receiverName, methodName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        ReceiverMethodAssignStmtRunner(packageReceiverMethodFunc, receiverName, methodName),
		ResultType: reflect.TypeOf([]*ast.AssignStmt{}),
	}
}

// ReceiverMethodCallExprAnalyzer returns an Analyzer for receiver method *ast.CallExpr
func ReceiverMethodCallExprAnalyzer(analyzerName string, packageReceiverMethodFunc func(ast.Expr, *types.Info, string, string) bool, packagePath string, receiverName string, methodName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find (%s.%s).%s calls for later passes", packagePath, receiverName, methodName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        ReceiverMethodCallExprRunner(packageReceiverMethodFunc, receiverName, methodName),
		ResultType: reflect.TypeOf([]*ast.CallExpr{}),
	}
}

// ReceiverMethodSelectorExprAnalyzer returns an Analyzer for receiver method *ast.SelectorExpr
func ReceiverMethodSelectorExprAnalyzer(analyzerName string, packageReceiverMethodFunc func(ast.Expr, *types.Info, string, string) bool, packagePath string, receiverName string, methodName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find (%s.%s).%s calls for later passes", packagePath, receiverName, methodName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        ReceiverMethodSelectorExprRunner(packageReceiverMethodFunc, receiverName, methodName),
		ResultType: reflect.TypeOf([]*ast.SelectorExpr{}),
	}
}

// SelectorExprAnalyzer returns an Analyzer for *ast.SelectorExpr
func SelectorExprAnalyzer(analyzerName string, packageFunc func(ast.Expr, *types.Info, string) bool, packagePath string, selectorName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find %s.%s usage for later passes", packagePath, selectorName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        SelectorExprRunner(packageFunc, selectorName),
		ResultType: reflect.TypeOf([]*ast.SelectorExpr{}),
	}
}

// TypeAssertExprRemovalAnalyzer returns an Analyzer for *ast.TypeAssertExpr
func TypeAssertExprRemovalAnalyzer(analyzerName string, typeAssertExprAnalyzer *analysis.Analyzer, packagePath string, selectorName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("remove %s.%s type assertions", packagePath, selectorName),
		Requires: []*analysis.Analyzer{
			typeAssertExprAnalyzer,
		},
		Run: TypeAssertExprRemovalRunner(analyzerName, typeAssertExprAnalyzer),
	}
}

// TypeAssertExprAnalyzer returns an Analyzer for *ast.TypeAssertExpr
func TypeAssertExprAnalyzer(analyzerName string, packageFunc func(ast.Expr, *types.Info, string) bool, packagePath string, selectorName string) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  fmt.Sprintf("find %s.%s type assertions for later passes", packagePath, selectorName),
		Requires: []*analysis.Analyzer{
			inspect.Analyzer,
		},
		Run:        TypeAssertExprRunner(packageFunc, selectorName),
		ResultType: reflect.TypeOf([]*ast.TypeAssertExpr{}),
	}
}
