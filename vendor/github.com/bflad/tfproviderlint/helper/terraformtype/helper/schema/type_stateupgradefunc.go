package schema

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
)

const (
	TypeNameStateUpgradeFunc = `StateUpgradeFunc`
)

// IsFuncTypeStateUpgradeFunc returns true if the FuncType matches expected parameters and results types
func IsFuncTypeStateUpgradeFunc(node ast.Node, info *types.Info) bool {
	funcType := astutils.FuncTypeFromNode(node)

	if funcType == nil {
		return false
	}

	if !astutils.HasFieldListLength(funcType.Params, 2) {
		return false
	}

	if !astutils.IsFieldListType(funcType.Params, 0, astutils.IsExprTypeMapStringInterface) {
		return false
	}

	if !astutils.IsFieldListType(funcType.Params, 1, astutils.IsExprTypeInterface) {
		return false
	}

	if !astutils.HasFieldListLength(funcType.Results, 2) {
		return false
	}

	if !astutils.IsFieldListType(funcType.Results, 0, astutils.IsExprTypeMapStringInterface) {
		return false
	}

	return astutils.IsFieldListType(funcType.Results, 1, astutils.IsExprTypeError)
}

// IsTypeStateUpgradeFunc returns if the type is StateUpgradeFunc from the schema package
func IsTypeStateUpgradeFunc(t types.Type) bool {
	switch t := t.(type) {
	case *types.Named:
		return IsNamedType(t, TypeNameStateUpgradeFunc)
	case *types.Pointer:
		return IsTypeStateUpgradeFunc(t.Elem())
	default:
		return false
	}
}

// StateUpgradeFuncInfo represents all gathered StateUpgradeFunc data for easier access
type StateUpgradeFuncInfo struct {
	AstFuncDecl *ast.FuncDecl
	AstFuncLit  *ast.FuncLit
	Body        *ast.BlockStmt
	Node        ast.Node
	Pos         token.Pos
	Type        *ast.FuncType
	TypesInfo   *types.Info
}

// NewStateUpgradeFuncInfo instantiates a StateUpgradeFuncInfo
func NewStateUpgradeFuncInfo(node ast.Node, info *types.Info) *StateUpgradeFuncInfo {
	result := &StateUpgradeFuncInfo{
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
