package analysisutils

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"path/filepath"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// AvoidSelectorExprRunner returns an Analyzer runner for *ast.SelectorExpr to avoid
func AvoidSelectorExprRunner(analyzerName string, callExprAnalyzer, selectorExprAnalyzer *analysis.Analyzer, packagePath, typeName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		callExprs := pass.ResultOf[callExprAnalyzer].([]*ast.CallExpr)
		selectorExprs := pass.ResultOf[selectorExprAnalyzer].([]*ast.SelectorExpr)
		ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

		// CallExpr and SelectorExpr will overlap, so only perform one report/fix
		reported := make(map[token.Pos]struct{})

		for _, callExpr := range callExprs {
			if ignorer.ShouldIgnore(analyzerName, callExpr) {
				continue
			}

			pass.Report(analysis.Diagnostic{
				Pos:     callExpr.Pos(),
				End:     callExpr.End(),
				Message: fmt.Sprintf("%s: avoid %s.%s", analyzerName, packagePath, typeName),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Remove",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     callExpr.Pos(),
								End:     callExpr.End(),
								NewText: []byte{},
							},
						},
					},
				},
			})

			reported[callExpr.Pos()] = struct{}{}
		}

		for _, selectorExpr := range selectorExprs {
			if ignorer.ShouldIgnore(analyzerName, selectorExpr) {
				continue
			}

			if _, ok := reported[selectorExpr.Pos()]; ok {
				continue
			}

			pass.Report(analysis.Diagnostic{
				Pos:     selectorExpr.Pos(),
				End:     selectorExpr.End(),
				Message: fmt.Sprintf("%s: avoid %s.%s", analyzerName, packagePath, typeName),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Remove",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     selectorExpr.Pos(),
								End:     selectorExpr.End(),
								NewText: []byte{},
							},
						},
					},
				},
			})
		}

		return nil, nil
	}
}

// DeprecatedReceiverMethodSelectorExprRunner returns an Analyzer runner for deprecated *ast.SelectorExpr
func DeprecatedReceiverMethodSelectorExprRunner(analyzerName string, callExprAnalyzer, selectorExprAnalyzer *analysis.Analyzer, packagePath, typeName, methodName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		callExprs := pass.ResultOf[callExprAnalyzer].([]*ast.CallExpr)
		selectorExprs := pass.ResultOf[selectorExprAnalyzer].([]*ast.SelectorExpr)
		ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

		// CallExpr and SelectorExpr will overlap, so only perform one report/fix
		reported := make(map[token.Pos]struct{})

		for _, callExpr := range callExprs {
			if ignorer.ShouldIgnore(analyzerName, callExpr) {
				continue
			}

			var callExprBuf bytes.Buffer

			if err := format.Node(&callExprBuf, pass.Fset, callExpr); err != nil {
				return nil, fmt.Errorf("error formatting original: %s", err)
			}

			pass.Report(analysis.Diagnostic{
				Pos:     callExpr.Pos(),
				End:     callExpr.End(),
				Message: fmt.Sprintf("%s: deprecated %s", analyzerName, callExprBuf.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Remove",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     callExpr.Pos(),
								End:     callExpr.End(),
								NewText: []byte{},
							},
						},
					},
				},
			})

			reported[callExpr.Pos()] = struct{}{}
		}

		for _, selectorExpr := range selectorExprs {
			if ignorer.ShouldIgnore(analyzerName, selectorExpr) {
				continue
			}

			if _, ok := reported[selectorExpr.Pos()]; ok {
				continue
			}

			var selectorExprBuf bytes.Buffer

			if err := format.Node(&selectorExprBuf, pass.Fset, selectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting original: %s", err)
			}

			pass.Report(analysis.Diagnostic{
				Pos:     selectorExpr.Pos(),
				End:     selectorExpr.End(),
				Message: fmt.Sprintf("%s: deprecated %s", analyzerName, selectorExprBuf.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Remove",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     selectorExpr.Pos(),
								End:     selectorExpr.End(),
								NewText: []byte{},
							},
						},
					},
				},
			})
		}

		return nil, nil
	}
}

// DeprecatedEmptyCallExprWithReplacementSelectorExprRunner returns an Analyzer runner for deprecated *ast.SelectorExpr with replacement
func DeprecatedEmptyCallExprWithReplacementSelectorExprRunner(analyzerName string, callExprAnalyzer *analysis.Analyzer, selectorExprAnalyzer *analysis.Analyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		callExprs := pass.ResultOf[callExprAnalyzer].([]*ast.CallExpr)
		selectorExprs := pass.ResultOf[selectorExprAnalyzer].([]*ast.SelectorExpr)
		ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

		// CallExpr and SelectorExpr will overlap, so only perform one report/fix
		reported := make(map[token.Pos]struct{})

		for _, callExpr := range callExprs {
			if ignorer.ShouldIgnore(analyzerName, callExpr) {
				continue
			}

			if len(callExpr.Args) != 0 {
				continue
			}

			selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)

			if !ok {
				continue
			}

			newSelectorExpr := &ast.SelectorExpr{
				Sel: selectorExpr.Sel,
				X:   selectorExpr.X,
			}

			if oldPackagePath != newPackagePath {
				newSelectorExpr.X = &ast.Ident{
					Name: filepath.Base(newPackagePath),
				}
			}

			if oldSelectorName != newSelectorName {
				newSelectorExpr.Sel = &ast.Ident{
					Name: newSelectorName,
				}
			}

			var callExprBuf, newSelectorExprBuf bytes.Buffer

			if err := format.Node(&callExprBuf, pass.Fset, selectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting original: %s", err)
			}

			if err := format.Node(&newSelectorExprBuf, pass.Fset, newSelectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting new: %s", err)
			}

			pass.Report(analysis.Diagnostic{
				Pos:     callExpr.Pos(),
				End:     callExpr.End(),
				Message: fmt.Sprintf("%s: deprecated %s should be replaced with %s", analyzerName, callExprBuf.String(), newSelectorExprBuf.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Replace",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     callExpr.Pos(),
								End:     callExpr.End(),
								NewText: newSelectorExprBuf.Bytes(),
							},
						},
					},
				},
			})

			reported[callExpr.Pos()] = struct{}{}
		}

		for _, selectorExpr := range selectorExprs {
			if ignorer.ShouldIgnore(analyzerName, selectorExpr) {
				continue
			}

			if _, ok := reported[selectorExpr.Pos()]; ok {
				continue
			}

			newSelectorExpr := &ast.SelectorExpr{
				Sel: selectorExpr.Sel,
				X:   selectorExpr.X,
			}

			if oldPackagePath != newPackagePath {
				newSelectorExpr.X = &ast.Ident{
					Name: filepath.Base(newPackagePath),
				}
			}

			if oldSelectorName != newSelectorName {
				newSelectorExpr.Sel = &ast.Ident{
					Name: newSelectorName,
				}
			}

			var selectorExprBuf, newSelectorExprBuf bytes.Buffer

			if err := format.Node(&selectorExprBuf, pass.Fset, selectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting original: %s", err)
			}

			if err := format.Node(&newSelectorExprBuf, pass.Fset, newSelectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting new: %s", err)
			}

			pass.Report(analysis.Diagnostic{
				Pos:     selectorExpr.Pos(),
				End:     selectorExpr.End(),
				Message: fmt.Sprintf("%s: deprecated %s should be replaced with %s", analyzerName, selectorExprBuf.String(), newSelectorExprBuf.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Replace",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     selectorExpr.Pos(),
								End:     selectorExpr.End(),
								NewText: newSelectorExprBuf.Bytes(),
							},
						},
					},
				},
			})
		}

		return nil, nil
	}
}

// DeprecatedWithReplacementPointerSelectorExprRunner returns an Analyzer runner for deprecated *ast.SelectorExpr with replacement
func DeprecatedWithReplacementPointerSelectorExprRunner(analyzerName string, selectorExprAnalyzer *analysis.Analyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
		selectorExprs := pass.ResultOf[selectorExprAnalyzer].([]*ast.SelectorExpr)

		for _, selectorExpr := range selectorExprs {
			if ignorer.ShouldIgnore(analyzerName, selectorExpr) {
				continue
			}

			newSelectorExpr := &ast.SelectorExpr{
				Sel: selectorExpr.Sel,
				X:   selectorExpr.X,
			}

			if oldPackagePath != newPackagePath {
				newSelectorExpr.X = &ast.Ident{
					Name: filepath.Base(newPackagePath),
				}
			}

			if oldSelectorName != newSelectorName {
				newSelectorExpr.Sel = &ast.Ident{
					Name: newSelectorName,
				}
			}

			newStarExpr := &ast.StarExpr{
				X: newSelectorExpr,
			}

			var selectorExprBuf, newStarExprBuf bytes.Buffer

			if err := format.Node(&selectorExprBuf, pass.Fset, selectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting original: %s", err)
			}

			if err := format.Node(&newStarExprBuf, pass.Fset, newStarExpr); err != nil {
				return nil, fmt.Errorf("error formatting new: %s", err)
			}

			pass.Report(analysis.Diagnostic{
				Pos:     selectorExpr.Pos(),
				End:     selectorExpr.End(),
				Message: fmt.Sprintf("%s: deprecated %s should be replaced with %s", analyzerName, selectorExprBuf.String(), newStarExprBuf.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Replace",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     selectorExpr.Pos(),
								End:     selectorExpr.End(),
								NewText: newStarExprBuf.Bytes(),
							},
						},
					},
				},
			})
		}

		return nil, nil
	}
}

// DeprecatedWithReplacementSelectorExprRunner returns an Analyzer runner for deprecated *ast.SelectorExpr with replacement
func DeprecatedWithReplacementSelectorExprRunner(analyzerName string, selectorExprAnalyzer *analysis.Analyzer, oldPackagePath, oldSelectorName, newPackagePath, newSelectorName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		selectorExprs := pass.ResultOf[selectorExprAnalyzer].([]*ast.SelectorExpr)
		ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)

		for _, selectorExpr := range selectorExprs {
			if ignorer.ShouldIgnore(analyzerName, selectorExpr) {
				continue
			}

			newSelectorExpr := &ast.SelectorExpr{
				Sel: selectorExpr.Sel,
				X:   selectorExpr.X,
			}

			if oldPackagePath != newPackagePath {
				newSelectorExpr.X = &ast.Ident{
					Name: filepath.Base(newPackagePath),
				}
			}

			if oldSelectorName != newSelectorName {
				newSelectorExpr.Sel = &ast.Ident{
					Name: newSelectorName,
				}
			}

			var selectorExprBuf, newSelectorExprBuf bytes.Buffer

			if err := format.Node(&selectorExprBuf, pass.Fset, selectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting original: %s", err)
			}

			if err := format.Node(&newSelectorExprBuf, pass.Fset, newSelectorExpr); err != nil {
				return nil, fmt.Errorf("error formatting new: %s", err)
			}

			pass.Report(analysis.Diagnostic{
				Pos:     selectorExpr.Pos(),
				End:     selectorExpr.End(),
				Message: fmt.Sprintf("%s: deprecated %s should be replaced with %s", analyzerName, selectorExprBuf.String(), newSelectorExprBuf.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Replace",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     selectorExpr.Pos(),
								End:     selectorExpr.End(),
								NewText: newSelectorExprBuf.Bytes(),
							},
						},
					},
				},
			})
		}

		return nil, nil
	}
}

// FunctionCallExprRunner returns an Analyzer runner for function *ast.CallExpr
func FunctionCallExprRunner(packageFunc func(ast.Expr, *types.Info, string) bool, functionName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.CallExpr)(nil),
		}
		var result []*ast.CallExpr

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			callExpr := n.(*ast.CallExpr)

			if !packageFunc(callExpr.Fun, pass.TypesInfo, functionName) {
				return
			}

			result = append(result, callExpr)
		})

		return result, nil
	}
}

// ReceiverMethodAssignStmtRunner returns an Analyzer runner for receiver method *ast.AssignStmt
func ReceiverMethodAssignStmtRunner(packageReceiverMethodFunc func(ast.Expr, *types.Info, string, string) bool, receiverName string, methodName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.AssignStmt)(nil),
		}
		var result []*ast.AssignStmt

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			assignStmt := n.(*ast.AssignStmt)

			if len(assignStmt.Rhs) != 1 {
				return
			}

			callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr)

			if !ok {
				return
			}

			if !packageReceiverMethodFunc(callExpr.Fun, pass.TypesInfo, receiverName, methodName) {
				return
			}

			result = append(result, assignStmt)
		})

		return result, nil
	}
}

// ReceiverMethodCallExprRunner returns an Analyzer runner for receiver method *ast.CallExpr
func ReceiverMethodCallExprRunner(packageReceiverMethodFunc func(ast.Expr, *types.Info, string, string) bool, receiverName string, methodName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.CallExpr)(nil),
		}
		var result []*ast.CallExpr

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			callExpr := n.(*ast.CallExpr)

			if !packageReceiverMethodFunc(callExpr.Fun, pass.TypesInfo, receiverName, methodName) {
				return
			}

			result = append(result, callExpr)
		})

		return result, nil
	}
}

// ReceiverMethodSelectorExprRunner returns an Analyzer runner for receiver method *ast.SelectorExpr
func ReceiverMethodSelectorExprRunner(packageReceiverMethodFunc func(ast.Expr, *types.Info, string, string) bool, receiverName string, methodName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.SelectorExpr)(nil),
		}
		var result []*ast.SelectorExpr

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			selectorExpr := n.(*ast.SelectorExpr)

			if !packageReceiverMethodFunc(selectorExpr, pass.TypesInfo, receiverName, methodName) {
				return
			}

			result = append(result, selectorExpr)
		})

		return result, nil
	}
}

// SelectorExprRunner returns an Analyzer runner for *ast.SelectorExpr
func SelectorExprRunner(packageFunc func(ast.Expr, *types.Info, string) bool, selectorName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.SelectorExpr)(nil),
		}
		var result []*ast.SelectorExpr

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			selectorExpr := n.(*ast.SelectorExpr)

			if !packageFunc(selectorExpr, pass.TypesInfo, selectorName) {
				return
			}

			result = append(result, selectorExpr)
		})

		return result, nil
	}
}

// TypeAssertExprRemovalRunner returns an Analyzer runner for removing *ast.TypeAssertExpr
func TypeAssertExprRemovalRunner(analyzerName string, typeAssertExprAnalyzer *analysis.Analyzer) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		typeAssertExprs := pass.ResultOf[typeAssertExprAnalyzer].([]*ast.TypeAssertExpr)

		for _, typeAssertExpr := range typeAssertExprs {
			var typeAssertExprBuf, xBuf bytes.Buffer

			if err := format.Node(&typeAssertExprBuf, pass.Fset, typeAssertExpr); err != nil {
				return nil, fmt.Errorf("error formatting original: %s", err)
			}

			if err := format.Node(&xBuf, pass.Fset, typeAssertExpr.X); err != nil {
				return nil, fmt.Errorf("error formatting new: %s", err)
			}

			pass.Report(analysis.Diagnostic{
				Pos:     typeAssertExpr.Pos(),
				End:     typeAssertExpr.End(),
				Message: fmt.Sprintf("%s: %s type assertion should be removed", analyzerName, typeAssertExprBuf.String()),
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Remove",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     typeAssertExpr.Pos(),
								End:     typeAssertExpr.End(),
								NewText: xBuf.Bytes(),
							},
						},
					},
				},
			})
		}

		return nil, nil
	}
}

// TypeAssertExprRunner returns an Analyzer runner for *ast.TypeAssertExpr
func TypeAssertExprRunner(packageFunc func(ast.Expr, *types.Info, string) bool, selectorName string) func(*analysis.Pass) (interface{}, error) {
	return func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		nodeFilter := []ast.Node{
			(*ast.TypeAssertExpr)(nil),
		}
		var result []*ast.TypeAssertExpr

		inspect.Preorder(nodeFilter, func(n ast.Node) {
			typeAssertExpr := n.(*ast.TypeAssertExpr)

			if !packageFunc(typeAssertExpr.Type, pass.TypesInfo, selectorName) {
				return
			}

			result = append(result, typeAssertExpr)
		})

		return result, nil
	}
}
