package astutils

import (
	"go/ast"
)

func CompositeLitField(cl *ast.CompositeLit, fieldName string) *ast.KeyValueExpr {
	for _, elt := range cl.Elts {
		switch e := elt.(type) {
		case *ast.KeyValueExpr:
			if e.Key.(*ast.Ident).Name != fieldName {
				continue
			}

			return e
		}
	}

	return nil
}

func CompositeLitFields(cl *ast.CompositeLit) map[string]*ast.KeyValueExpr {
	result := make(map[string]*ast.KeyValueExpr, len(cl.Elts))

	for _, elt := range cl.Elts {
		switch e := elt.(type) {
		case *ast.KeyValueExpr:
			result[e.Key.(*ast.Ident).Name] = e
		}
	}

	return result
}

func CompositeLitFieldBoolValue(cl *ast.CompositeLit, fieldName string) *bool {
	kvExpr := CompositeLitField(cl, fieldName)

	if kvExpr == nil {
		return nil
	}

	return ExprBoolValue(kvExpr.Value)
}

func CompositeLitFieldExprValue(cl *ast.CompositeLit, fieldName string) *ast.Expr {
	kvExpr := CompositeLitField(cl, fieldName)

	if kvExpr == nil {
		return nil
	}

	return ExprValue(kvExpr.Value)
}

func CompositeLitFieldIntValue(cl *ast.CompositeLit, fieldName string) *int {
	kvExpr := CompositeLitField(cl, fieldName)

	if kvExpr == nil {
		return nil
	}

	return ExprIntValue(kvExpr.Value)
}

func CompositeLitContainsAnyField(cl *ast.CompositeLit, fieldNames ...string) bool {
	for _, elt := range cl.Elts {
		switch e := elt.(type) {
		case *ast.KeyValueExpr:
			name := e.Key.(*ast.Ident).Name

			for _, field := range fieldNames {
				if name == field {
					return true
				}
			}
		}
	}

	return false
}
