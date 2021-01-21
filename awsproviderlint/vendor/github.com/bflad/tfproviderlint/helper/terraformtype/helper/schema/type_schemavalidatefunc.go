package schema

import (
	"go/ast"
	"go/token"
	"go/types"
)

// SchemaValidateFuncInfo represents all gathered SchemaValidateFunc data for easier access
type SchemaValidateFuncInfo struct {
	AstFuncDecl *ast.FuncDecl
	AstFuncLit  *ast.FuncLit
	Body        *ast.BlockStmt
	Node        ast.Node
	Pos         token.Pos
	Type        *ast.FuncType
	TypesInfo   *types.Info
}

// NewSchemaValidateFuncInfo instantiates a SchemaValidateFuncInfo
func NewSchemaValidateFuncInfo(node ast.Node, info *types.Info) *SchemaValidateFuncInfo {
	result := &SchemaValidateFuncInfo{
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
