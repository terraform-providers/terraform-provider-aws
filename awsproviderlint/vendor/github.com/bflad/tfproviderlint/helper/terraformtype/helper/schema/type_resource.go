package schema

import (
	"go/ast"
	"go/types"
	"time"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	ResourceFieldCreate               = `Create`
	ResourceFieldCreateContext        = `CreateContext`
	ResourceFieldCreateWithoutTimeout = `CreateWithoutTimeout`
	ResourceFieldCustomizeDiff        = `CustomizeDiff`
	ResourceFieldDelete               = `Delete`
	ResourceFieldDeleteContext        = `DeleteContext`
	ResourceFieldDeleteWithoutTimeout = `DeleteWithoutTimeout`
	ResourceFieldDeprecationMessage   = `DeprecationMessage`
	ResourceFieldDescription          = `Description`
	ResourceFieldExists               = `Exists`
	ResourceFieldImporter             = `Importer`
	ResourceFieldMigrateState         = `MigrateState`
	ResourceFieldRead                 = `Read`
	ResourceFieldReadContext          = `ReadContext`
	ResourceFieldReadWithoutTimeout   = `ReadWithoutTimeout`
	ResourceFieldSchema               = `Schema`
	ResourceFieldSchemaVersion        = `SchemaVersion`
	ResourceFieldStateUpgraders       = `StateUpgraders`
	ResourceFieldTimeouts             = `Timeouts`
	ResourceFieldUpdate               = `Update`
	ResourceFieldUpdateContext        = `UpdateContext`
	ResourceFieldUpdateWithoutTimeout = `UpdateWithoutTimeout`

	TypeNameResource = `Resource`
)

// resourceType is an internal representation of the SDK helper/schema.Resource type
//
// This is used to prevent importing the real type since the project supports
// multiple versions of the Terraform Plugin SDK, while allowing passes to
// access the data in a familiar manner.
type resourceType struct {
	Description  string
	MigrateState func(int, interface{}, interface{}) (interface{}, error)
	Schema       map[string]*schemaType
	Timeouts     resourceTimeoutType
}

// ResourceInfo represents all gathered Resource data for easier access
type ResourceInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	Resource        *resourceType
	TypesInfo       *types.Info
}

// NewResourceInfo instantiates a ResourceInfo
func NewResourceInfo(cl *ast.CompositeLit, info *types.Info) *ResourceInfo {
	result := &ResourceInfo{
		AstCompositeLit: cl,
		Fields:          astutils.CompositeLitFields(cl),
		Resource:        &resourceType{},
		TypesInfo:       info,
	}

	if kvExpr := result.Fields[ResourceFieldDescription]; kvExpr != nil && astutils.ExprStringValue(kvExpr.Value) != nil {
		result.Resource.Description = *astutils.ExprStringValue(kvExpr.Value)
	}

	if kvExpr := result.Fields[ResourceFieldMigrateState]; kvExpr != nil && astutils.ExprValue(kvExpr.Value) != nil {
		result.Resource.MigrateState = func(int, interface{}, interface{}) (interface{}, error) { return nil, nil }
	}

	if kvExpr := result.Fields[ResourceFieldTimeouts]; kvExpr != nil && astutils.ExprValue(kvExpr.Value) != nil {
		if timeoutUexpr, ok := kvExpr.Value.(*ast.UnaryExpr); ok {
			if timeoutClit, ok := timeoutUexpr.X.(*ast.CompositeLit); ok {
				for _, expr := range timeoutClit.Elts {
					switch elt := expr.(type) {
					case *ast.KeyValueExpr:
						var key string

						switch keyExpr := elt.Key.(type) {
						case *ast.Ident:
							key = keyExpr.Name
						}

						if key == "" {
							continue
						}

						if astutils.ExprValue(elt.Value) != nil {
							switch key {
							case "Create":
								result.Resource.Timeouts.Create = new(time.Duration)
							case "Read":
								result.Resource.Timeouts.Read = new(time.Duration)
							case "Update":
								result.Resource.Timeouts.Update = new(time.Duration)
							case "Delete":
								result.Resource.Timeouts.Delete = new(time.Duration)
							case "Default":
								result.Resource.Timeouts.Default = new(time.Duration)
							}
						}
					}
				}
			}
		}
	}

	if kvExpr := result.Fields[ResourceFieldSchema]; kvExpr != nil && astutils.ExprValue(kvExpr.Value) != nil {
		result.Resource.Schema = map[string]*schemaType{}
		if smap, ok := kvExpr.Value.(*ast.CompositeLit); ok {
			for _, expr := range smap.Elts {
				switch elt := expr.(type) {
				case *ast.KeyValueExpr:
					var key string

					switch keyExpr := elt.Key.(type) {
					case *ast.BasicLit:
						keyPtr := astutils.ExprStringValue(keyExpr)

						if keyPtr == nil {
							continue
						}

						key = *keyPtr
					}

					if key == "" {
						continue
					}

					switch valueExpr := elt.Value.(type) {
					case *ast.CompositeLit:
						result.Resource.Schema[key] = NewSchemaInfo(valueExpr, info).Schema
					}
				}
			}
		}
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
