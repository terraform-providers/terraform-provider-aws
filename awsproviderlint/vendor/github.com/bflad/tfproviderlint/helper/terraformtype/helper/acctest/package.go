package acctest

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype"
)

const (
	PackageModule     = terraformtype.ModuleTerraformPluginSdk
	PackageModulePath = `helper/acctest`
	PackageName       = `acctest`
	PackagePath       = PackageModule + `/` + PackageModulePath
)

// IsConst returns if the expr is a constant in the acctest package
func IsConst(e ast.Expr, info *types.Info, constName string) bool {
	// IsModulePackageFunc can handle any SelectorExpr name
	return astutils.IsModulePackageFunc(e, info, PackageModule, PackageModulePath, constName)
}

// IsFunc returns if the function call is in the acctest package
func IsFunc(e ast.Expr, info *types.Info, funcName string) bool {
	return astutils.IsModulePackageFunc(e, info, PackageModule, PackageModulePath, funcName)
}

// IsNamedType returns if the type name matches and is from the helper/acctest package
func IsNamedType(t *types.Named, typeName string) bool {
	return astutils.IsModulePackageNamedType(t, PackageModule, PackageModulePath, typeName)
}

// PackagePathVersion returns the import path for a module version
func PackagePathVersion(moduleVersion int) string {
	switch moduleVersion {
	case 0, 1:
		return PackagePath
	default:
		return fmt.Sprintf("%s/v%d/%s", PackageModule, moduleVersion, PackageModulePath)
	}
}
