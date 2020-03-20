package resource

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	PackageName = `resource`
	PackagePath = `github.com/hashicorp/terraform-plugin-sdk/helper/resource`
)

// IsFunc returns if the function call is in the resource package
func IsFunc(e ast.Expr, info *types.Info, funcName string) bool {
	return astutils.IsPackageFunc(e, info, PackagePath, funcName)
}

// IsNamedType returns if the type name matches and is from the helper/resource package
func IsNamedType(t *types.Named, typeName string) bool {
	return astutils.IsPackageNamedType(t, PackagePath, typeName)
}
