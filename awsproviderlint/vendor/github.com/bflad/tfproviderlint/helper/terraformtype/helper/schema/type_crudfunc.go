package schema

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/diag"
)

// IsFuncTypeCRUDFunc returns true if the FuncType matches expected parameters and results types
func IsFuncTypeCRUDFunc(node ast.Node, info *types.Info) bool {
	funcType := astutils.FuncTypeFromNode(node)

	if funcType == nil {
		return false
	}

	return isFuncTypeCRUDFunc(funcType, info) || isFuncTypeCRUDContextFunc(funcType, info)
}

// isFuncTypeCRUDFunc returns true if the FuncType matches expected parameters and results types of V1 or V2 without a context.
func isFuncTypeCRUDFunc(funcType *ast.FuncType, info *types.Info) bool {
	if !astutils.HasFieldListLength(funcType.Params, 2) {
		return false
	}

	if !astutils.IsFieldListTypeModulePackageType(funcType.Params, 0, info, PackageModule, PackageModulePath, TypeNameResourceData) {
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

// isFuncTypeCRUDContextFunc returns true if the FuncType matches expected parameters and results types of V2 with a context.
func isFuncTypeCRUDContextFunc(funcType *ast.FuncType, info *types.Info) bool {
	if !astutils.HasFieldListLength(funcType.Params, 3) {
		return false
	}

	if !astutils.IsFieldListTypePackageType(funcType.Params, 0, info, "context", "Context") {
		return false
	}

	if !astutils.IsFieldListTypeModulePackageType(funcType.Params, 1, info, PackageModule, PackageModulePath, TypeNameResourceData) {
		return false
	}

	if !astutils.IsFieldListType(funcType.Params, 2, astutils.IsExprTypeInterface) {
		return false
	}

	if !astutils.HasFieldListLength(funcType.Results, 1) {
		return false
	}

	if !astutils.IsFieldListTypeModulePackageType(funcType.Results, 0, info, diag.PackageModule, diag.PackageModulePath, diag.TypeNameDiagnostics) {
		return false
	}

	return true
}

// CRUDFuncInfo represents all gathered CreateContext, ReadContext, UpdateContext, and DeleteContext data for easier access
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
