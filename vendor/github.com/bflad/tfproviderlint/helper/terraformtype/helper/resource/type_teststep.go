package resource

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	tfresource "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	TestStepFieldCheck                     = `Check`
	TestStepFieldConfig                    = `Config`
	TestStepFieldDestroy                   = `Destroy`
	TestStepFieldExpectError               = `ExpectError`
	TestStepFieldExpectNonEmptyPlan        = `ExpectNonEmptyPlan`
	TestStepFieldImportState               = `ImportState`
	TestStepFieldImportStateId             = `ImportStateId`
	TestStepFieldImportStateIdFunc         = `ImportStateIdFunc`
	TestStepFieldImportStateIdPrefix       = `ImportStateIdPrefix`
	TestStepFieldImportStateCheck          = `ImportStateCheck`
	TestStepFieldImportStateVerify         = `ImportStateVerify`
	TestStepFieldImportStateVerifyIgnore   = `ImportStateVerifyIgnore`
	TestStepFieldPlanOnly                  = `PlanOnly`
	TestStepFieldPreConfig                 = `PreConfig`
	TestStepFieldPreventDiskCleanup        = `PreventDiskCleanup`
	TestStepFieldPreventPostDestroyRefresh = `PreventPostDestroyRefresh`
	TestStepFieldResourceName              = `ResourceName`
	TestStepFieldSkipFunc                  = `SkipFunc`
	TestStepFieldTaint                     = `Taint`

	TypeNameTestStep = `TestStep`
)

// TestStepInfo represents all gathered TestStep data for easier access
type TestStepInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	TestStep        *tfresource.TestStep
	TypesInfo       *types.Info
}

// NewTestStepInfo instantiates a TestStepInfo
func NewTestStepInfo(cl *ast.CompositeLit, info *types.Info) *TestStepInfo {
	result := &TestStepInfo{
		AstCompositeLit: cl,
		Fields:          astutils.CompositeLitFields(cl),
		TestStep:        &tfresource.TestStep{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *TestStepInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsTypeTestStep returns if the type is TestStep from the helper/schema package
func IsTypeTestStep(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameTestStep)
	case *types.Pointer:
		return IsTypeTestStep(t.Elem())
	default:
		return false
	}
}
