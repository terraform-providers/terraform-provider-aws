package astutils

import (
	"go/ast"
	"go/types"
	"strings"
)

// IsPackageReceiverMethod returns true if the package suffix (for vendoring), receiver name, and method name match
func IsPackageReceiverMethod(e ast.Expr, info *types.Info, packageSuffix string, receiverName, methodName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != methodName {
			return false
		}

		return IsPackageType(info.TypeOf(e.X), packageSuffix, receiverName)
	}

	return false
}

// IsPackageFunc returns true if the function package suffix (for vendoring) and name matches
func IsPackageFunc(e ast.Expr, info *types.Info, packageSuffix string, funcName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != funcName {
			return false
		}

		switch x := e.X.(type) {
		case *ast.Ident:
			return strings.HasSuffix(info.ObjectOf(x).(*types.PkgName).Imported().Path(), packageSuffix)
		}
	case *ast.StarExpr:
		return IsPackageFunc(e.X, info, packageSuffix, funcName)
	}

	return false
}

// IsPackageFunctionFieldListType returns true if the function parameter package suffix (for vendoring) and name matches
func IsPackageFunctionFieldListType(e ast.Expr, info *types.Info, packageSuffix string, typeName string) bool {
	switch e := e.(type) {
	case *ast.SelectorExpr:
		if e.Sel.Name != typeName {
			return false
		}

		switch x := e.X.(type) {
		case *ast.Ident:
			return strings.HasSuffix(info.ObjectOf(x).(*types.PkgName).Imported().Path(), packageSuffix)
		}
	case *ast.StarExpr:
		return IsPackageFunctionFieldListType(e.X, info, packageSuffix, typeName)
	}

	return false
}

// IsPackageNamedType returns if the type name matches and is from the package suffix
func IsPackageNamedType(t *types.Named, packageSuffix string, typeName string) bool {
	if t.Obj().Name() != typeName {
		return false
	}

	// HasSuffix here due to vendoring
	return strings.HasSuffix(t.Obj().Pkg().Path(), packageSuffix)
}

// IsPackageType returns true if the type name can be matched and is from the package suffix
func IsPackageType(t types.Type, packageSuffix string, typeName string) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsPackageNamedType(t, packageSuffix, typeName)
	case *types.Pointer:
		return IsPackageType(t.Elem(), packageSuffix, typeName)
	}

	return false
}
