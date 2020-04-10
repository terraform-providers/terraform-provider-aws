package schema

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	TypeNameCustomizeDiffFunc = `CustomizeDiffFunc`
)

// IsFuncTypeCustomizeDiffFunc returns true if the FuncType matches expected parameters and results types
func IsFuncTypeCustomizeDiffFunc(node ast.Node, info *types.Info) bool {
	funcType := astutils.FuncTypeFromNode(node)

	if funcType == nil {
		return false
	}

	if !astutils.HasFieldListLength(funcType.Params, 2) {
		return false
	}

	if !astutils.IsFieldListTypePackageType(funcType.Params, 0, info, PackagePath, TypeNameResourceDiff) {
		return false
	}

	if !astutils.IsFieldListType(funcType.Params, 1, astutils.IsExprTypeInterface) {
		return false
	}

	if !astutils.HasFieldListLength(funcType.Results, 1) {
		return false
	}

	return astutils.IsFieldListType(funcType.Results, 0, astutils.IsExprTypeError)
}

// IsTypeCustomizeDiffFunc returns if the type is CustomizeDiffFunc from the customdiff package
func IsTypeCustomizeDiffFunc(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameCustomizeDiffFunc)
	case *types.Pointer:
		return IsTypeCustomizeDiffFunc(t.Elem())
	default:
		return false
	}
}

// CustomizeDiffFuncInfo represents all gathered CustomizeDiffFunc data for easier access
type CustomizeDiffFuncInfo struct {
	AstFuncDecl *ast.FuncDecl
	AstFuncLit  *ast.FuncLit
	Body        *ast.BlockStmt
	Node        ast.Node
	Pos         token.Pos
	Type        *ast.FuncType
	TypesInfo   *types.Info
}

// NewCustomizeDiffFuncInfo instantiates a CustomizeDiffFuncInfo
func NewCustomizeDiffFuncInfo(node ast.Node, info *types.Info) *CustomizeDiffFuncInfo {
	result := &CustomizeDiffFuncInfo{
		TypesInfo: info,
	}

	switch node := node.(type) {
	case *ast.FuncDecl:
		result.AstFuncDecl = node
		result.Body = node.Body
		result.Node = node
		result.Pos = node.Pos()
		result.Type = node.Type
	case *ast.FuncLit:
		result.AstFuncLit = node
		result.Body = node.Body
		result.Node = node
		result.Pos = node.Pos()
		result.Type = node.Type
	}

	return result
}
