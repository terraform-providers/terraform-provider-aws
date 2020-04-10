package resource

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	tfresource "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	TestCaseFieldCheckDestroy              = `CheckDestroy`
	TestCaseFieldIDRefreshName             = `IDRefreshName`
	TestCaseFieldIDRefreshIgnore           = `IDRefreshIgnore`
	TestCaseFieldIsUnitTest                = `IsUnitTest`
	TestCaseFieldPreCheck                  = `PreCheck`
	TestCaseFieldPreventPostDestroyRefresh = `PreventPostDestroyRefresh`
	TestCaseFieldProviders                 = `Providers`
	TestCaseFieldProviderFactories         = `ProviderFactories`
	TestCaseFieldSteps                     = `Steps`

	TypeNameTestCase = `TestCase`
)

// TestCaseInfo represents all gathered TestCase data for easier access
type TestCaseInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	TestCase        *tfresource.TestCase
	TypesInfo       *types.Info
}

// NewTestCaseInfo instantiates a TestCaseInfo
func NewTestCaseInfo(cl *ast.CompositeLit, info *types.Info) *TestCaseInfo {
	result := &TestCaseInfo{
		AstCompositeLit: cl,
		Fields:          astutils.CompositeLitFields(cl),
		TestCase:        &tfresource.TestCase{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *TestCaseInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsTypeTestCase returns if the type is TestCase from the helper/schema package
func IsTypeTestCase(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameTestCase)
	case *types.Pointer:
		return IsTypeTestCase(t.Elem())
	default:
		return false
	}
}
