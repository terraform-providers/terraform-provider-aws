package validation

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	PackageName = `validation`
	PackagePath = `github.com/hashicorp/terraform-plugin-sdk/helper/validation`
)

// IsFunc returns if the function call is in the package
func IsFunc(e ast.Expr, info *types.Info, funcName string) bool {
	return astutils.IsPackageFunc(e, info, PackagePath, funcName)
}
