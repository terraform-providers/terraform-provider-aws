package resource

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	tfresource "github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	RetryErrorFieldErr       = `Err`
	RetryErrorFieldRetryable = `Retryable`

	TypeNameRetryError = `RetryError`
)

// RetryErrorInfo represents all gathered RetryError data for easier access
type RetryErrorInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	RetryError      *tfresource.RetryError
	TypesInfo       *types.Info
}

// NewRetryErrorInfo instantiates a RetryErrorInfo
func NewRetryErrorInfo(cl *ast.CompositeLit, info *types.Info) *RetryErrorInfo {
	result := &RetryErrorInfo{
		AstCompositeLit: cl,
		Fields:          astutils.CompositeLitFields(cl),
		RetryError:      &tfresource.RetryError{},
		TypesInfo:       info,
	}

	return result
}

// DeclaresField returns true if the field name is present in the AST
func (info *RetryErrorInfo) DeclaresField(fieldName string) bool {
	return info.Fields[fieldName] != nil
}

// IsTypeRetryError returns if the type is RetryError from the helper/resource package
func IsTypeRetryError(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameRetryError)
	case *types.Pointer:
		return IsTypeRetryError(t.Elem())
	default:
		return false
	}
}
