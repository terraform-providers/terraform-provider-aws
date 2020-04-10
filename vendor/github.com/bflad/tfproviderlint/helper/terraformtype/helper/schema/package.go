package schema

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	PackageName = `schema`
	PackagePath = `github.com/hashicorp/terraform-plugin-sdk/helper/schema`
)

// IsFunc returns if the function call is in the package
func IsFunc(e ast.Expr, info *types.Info, funcName string) bool {
	return astutils.IsPackageFunc(e, info, PackagePath, funcName)
}

// IsNamedType returns if the type name matches and is from the package
func IsNamedType(t *types.Named, typeName string) bool {
	return astutils.IsPackageNamedType(t, PackagePath, typeName)
}

// IsReceiverMethod returns if the receiver method call is in the package
func IsReceiverMethod(e ast.Expr, info *types.Info, receiverName string, methodName string) bool {
	return astutils.IsPackageReceiverMethod(e, info, PackagePath, receiverName, methodName)
}
