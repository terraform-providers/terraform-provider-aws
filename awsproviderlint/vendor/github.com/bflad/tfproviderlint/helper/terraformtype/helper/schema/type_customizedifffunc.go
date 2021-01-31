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

	return isFuncTypeCustomizeDiffFuncV1(funcType, info) || isFuncTypeCustomizeDiffFuncV2(funcType, info)
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

// isFuncTypeCustomizeDiffFuncV1 returns true if the FuncType matches expected parameters and results types of V1
func isFuncTypeCustomizeDiffFuncV1(funcType *ast.FuncType, info *types.Info) bool {
	if !astutils.HasFieldListLength(funcType.Params, 2) {
		return false
	}

	if !astutils.IsFieldListTypePackageType(funcType.Params, 0, info, PackagePathVersion(1), TypeNameResourceDiff) {
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

// isFuncTypeCustomizeDiffFuncV2 returns true if the FuncType matches expected parameters and results types of V2
func isFuncTypeCustomizeDiffFuncV2(funcType *ast.FuncType, info *types.Info) bool {
	if !astutils.HasFieldListLength(funcType.Params, 3) {
		return false
	}

	if !astutils.IsFieldListTypePackageType(funcType.Params, 0, info, "context", "Context") {
		return false
	}

	if !astutils.IsFieldListTypePackageType(funcType.Params, 1, info, PackagePathVersion(2), TypeNameResourceDiff) {
		return false
	}

	if !astutils.IsFieldListType(funcType.Params, 2, astutils.IsExprTypeInterface) {
		return false
	}

	if !astutils.HasFieldListLength(funcType.Results, 1) {
		return false
	}

	return astutils.IsFieldListType(funcType.Results, 0, astutils.IsExprTypeError)
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
