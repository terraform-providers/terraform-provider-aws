package schema

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	tfschema "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	ResourceFieldCreate             = `Create`
	ResourceFieldCustomizeDiff      = `CustomizeDiff`
	ResourceFieldDelete             = `Delete`
	ResourceFieldDeprecationMessage = `DeprecationMessage`
	ResourceFieldExists             = `Exists`
	ResourceFieldImporter           = `Importer`
	ResourceFieldMigrateState       = `MigrateState`
	ResourceFieldRead               = `Read`
	ResourceFieldSchema             = `Schema`
	ResourceFieldSchemaVersion      = `SchemaVersion`
	ResourceFieldStateUpgraders     = `StateUpgraders`
	ResourceFieldTimeouts           = `Timeouts`
	ResourceFieldUpdate             = `Update`

	TypeNameResource = `Resource`
)

// ResourceInfo represents all gathered Resource data for easier access
type ResourceInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	Resource        *tfschema.Resource
	TypesInfo       *types.Info
}

// NewResourceInfo instantiates a ResourceInfo
func NewResourceInfo(cl *ast.CompositeLit, info *types.Info) *ResourceInfo {
	result := &ResourceInfo{
		AstCompositeLit: cl,
		Fields:          astutils.CompositeLitFields(cl),
		Resource:        &tfschema.Resource{},
		TypesInfo:       info,
	}

	if kvExpr := result.Fields[ResourceFieldMigrateState]; kvExpr != nil && astutils.ExprValue(kvExpr.Value) != nil {
		result.Resource.MigrateState = func(int, *terraform.InstanceState, interface{}) (*terraform.InstanceState, error) { return nil, nil }
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *ResourceInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsDataSource returns true if the Resource type matches a Terraform Data Source declaration
func (info *ResourceInfo) IsDataSource() bool {
	if info.DeclaresField(ResourceFieldCreate) {
		return false
	}

	return info.DeclaresField(ResourceFieldRead)
}

// IsResource returns true if the Resource type matches a Terraform Resource declaration
func (info *ResourceInfo) IsResource() bool {
	return info.DeclaresField(ResourceFieldCreate)
}

// GetResourceMapResourceNames returns all resource names held in a map[string]*schema.Resource
func GetResourceMapResourceNames(cl *ast.CompositeLit) []ast.Expr {
	var result []ast.Expr

	for _, elt := range cl.Elts {
		switch v := elt.(type) {
		case *ast.KeyValueExpr:
			result = append(result, v.Key)
		}
	}

	return result
}

// IsMapStringResource returns if the type is map[string]*Resource from the helper/schema package
func IsMapStringResource(cl *ast.CompositeLit, info *types.Info) bool {
	switch v := cl.Type.(type) {
	case *ast.MapType:
		switch k := v.Key.(type) {
		case *ast.Ident:
			if k.Name != "string" {
				return false
			}
		}

		return IsTypeResource(info.TypeOf(v.Value))
	}

	return false
}

// IsTypeResource returns if the type is Resource from the helper/schema package
func IsTypeResource(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameResource)
	case *types.Pointer:
		return IsTypeResource(t.Elem())
	default:
		return false
	}
}
