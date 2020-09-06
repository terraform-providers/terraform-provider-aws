package resource

import (
	"go/ast"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	RetryErrorFieldErr       = `Err`
	RetryErrorFieldRetryable = `Retryable`

	TypeNameRetryError = `RetryError`
)

// retryErrorType is an internal representation of the SDK helper/resource.RetryError type
//
// This is used to prevent importing the real type since the project supports
// multiple versions of the Terraform Plugin SDK, while allowing passes to
// access the data in a familiar manner.
type retryErrorType struct {
	Err       error
	Retryable bool
}

// RetryErrorInfo represents all gathered RetryError data for easier access
type RetryErrorInfo struct {
	AstCompositeLit *ast.CompositeLit
	Fields          map[string]*ast.KeyValueExpr
	RetryError      *retryErrorType
	TypesInfo       *types.Info
}

// NewRetryErrorInfo instantiates a RetryErrorInfo
func NewRetryErrorInfo(cl *ast.CompositeLit, info *types.Info) *RetryErrorInfo {
	result := &RetryErrorInfo{
		AstCompositeLit: cl,
		Fields:          astutils.CompositeLitFields(cl),
		RetryError:      &retryErrorType{},
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
