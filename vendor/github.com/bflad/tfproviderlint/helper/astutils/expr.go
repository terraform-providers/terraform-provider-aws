package astutils

import (
	"go/ast"
)

// IsExprTypeArrayString returns true if the expression matches []string
func IsExprTypeArrayString(e ast.Expr) bool {
	arrayType, ok := e.(*ast.ArrayType)

	return ok && IsExprTypeString(arrayType.Elt)
}

// IsExprTypeArrayError returns true if the expression matches []error
func IsExprTypeArrayError(e ast.Expr) bool {
	arrayType, ok := e.(*ast.ArrayType)

	return ok && IsExprTypeError(arrayType.Elt)
}

// IsExprTypeError returns true if the expression matches string
func IsExprTypeError(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)

	return ok && ident.Name == "error"
}

// IsExprTypeInterface returns true if the expression matches interface{}
func IsExprTypeInterface(e ast.Expr) bool {
	_, ok := e.(*ast.InterfaceType)

	return ok
}

// IsExprTypeString returns true if the expression matches string
func IsExprTypeString(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)

	return ok && ident.Name == "string"
}
