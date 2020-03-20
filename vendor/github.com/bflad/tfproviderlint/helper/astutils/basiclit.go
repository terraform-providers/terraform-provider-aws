package astutils

import (
	"go/ast"
	"strconv"
	"strings"
)

// ExprBoolValue fetches a bool value from the Expr
// If the Expr cannot parse as a bool, returns nil.
func ExprBoolValue(e ast.Expr) *bool {
	switch v := e.(type) {
	case *ast.Ident:
		stringValue := v.Name
		boolValue, err := strconv.ParseBool(stringValue)

		if err != nil {
			return nil
		}

		return &boolValue
	}

	return nil
}

// ExprIntValue fetches an int value from the Expr
// If the Expr cannot parse as an int, returns nil.
func ExprIntValue(e ast.Expr) *int {
	switch v := e.(type) {
	case *ast.BasicLit:
		stringValue := strings.Trim(v.Value, `"`)
		intValue, err := strconv.Atoi(stringValue)

		if err != nil {
			return nil
		}

		return &intValue
	}

	return nil
}

// ExprStringValue fetches a string value from the Expr
// If the Expr is not BasicLit, returns an empty string.
func ExprStringValue(e ast.Expr) *string {
	switch v := e.(type) {
	case *ast.BasicLit:
		stringValue := strings.Trim(v.Value, `"`)

		return &stringValue
	}

	return nil
}

// ExprValue fetches a pointer to the Expr
// If the Expr is nil, returns nil
func ExprValue(e ast.Expr) *ast.Expr {
	switch v := e.(type) {
	case *ast.Ident:
		if v.Name == "nil" {
			return nil
		}

		return &e
	default:
		return &e
	}
}
