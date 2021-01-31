package schema

import (
	"go/ast"
	"go/token"
	"go/types"
)

// CRUDFuncInfo represents all gathered CreateFunc, ReadFunc, UpdateFunc, and DeleteFunc data for easier access
// Since Create, Delete, Read, and Update functions all have the same function
// signature, we cannot differentiate them in AST (except by potentially by
// function declaration naming heuristics later on).
type CRUDFuncInfo struct {
	AstFuncDecl *ast.FuncDecl
	AstFuncLit  *ast.FuncLit
	Body        *ast.BlockStmt
	Node        ast.Node
	Pos         token.Pos
	Type        *ast.FuncType
	TypesInfo   *types.Info
}

// NewCRUDFuncInfo instantiates a CRUDFuncInfo
func NewCRUDFuncInfo(node ast.Node, info *types.Info) *CRUDFuncInfo {
	result := &CRUDFuncInfo{
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
