package ruleguard

import (
	"go/ast"
)

type nodeCategory int

const (
	nodeUnknown nodeCategory = iota

	nodeArrayType
	nodeAssignStmt
	nodeBasicLit
	nodeBinaryExpr
	nodeBlockStmt
	nodeBranchStmt
	nodeCallExpr
	nodeCaseClause
	nodeChanType
	nodeCommClause
	nodeCompositeLit
	nodeDeclStmt
	nodeDeferStmt
	nodeEllipsis
	nodeEmptyStmt
	nodeExprStmt
	nodeForStmt
	nodeFuncDecl
	nodeFuncLit
	nodeFuncType
	nodeGenDecl
	nodeGoStmt
	nodeIdent
	nodeIfStmt
	nodeImportSpec
	nodeIncDecStmt
	nodeIndexExpr
	nodeInterfaceType
	nodeKeyValueExpr
	nodeLabeledStmt
	nodeMapType
	nodeParenExpr
	nodeRangeStmt
	nodeReturnStmt
	nodeSelectStmt
	nodeSelectorExpr
	nodeSendStmt
	nodeSliceExpr
	nodeStarExpr
	nodeStructType
	nodeSwitchStmt
	nodeTypeAssertExpr
	nodeTypeSpec
	nodeTypeSwitchStmt
	nodeUnaryExpr
	nodeValueSpec

	nodeCategoriesCount
)

func categorizeNode(n ast.Node) nodeCategory {
	switch n.(type) {
	case *ast.ArrayType:
		return nodeArrayType
	case *ast.AssignStmt:
		return nodeAssignStmt
	case *ast.BasicLit:
		return nodeBasicLit
	case *ast.BinaryExpr:
		return nodeBinaryExpr
	case *ast.BlockStmt:
		return nodeBlockStmt
	case *ast.BranchStmt:
		return nodeBranchStmt
	case *ast.CallExpr:
		return nodeCallExpr
	case *ast.CaseClause:
		return nodeCaseClause
	case *ast.ChanType:
		return nodeChanType
	case *ast.CommClause:
		return nodeCommClause
	case *ast.CompositeLit:
		return nodeCompositeLit
	case *ast.DeclStmt:
		return nodeDeclStmt
	case *ast.DeferStmt:
		return nodeDeferStmt
	case *ast.Ellipsis:
		return nodeEllipsis
	case *ast.EmptyStmt:
		return nodeEmptyStmt
	case *ast.ExprStmt:
		return nodeExprStmt
	case *ast.ForStmt:
		return nodeForStmt
	case *ast.FuncDecl:
		return nodeFuncDecl
	case *ast.FuncLit:
		return nodeFuncLit
	case *ast.FuncType:
		return nodeFuncType
	case *ast.GenDecl:
		return nodeGenDecl
	case *ast.GoStmt:
		return nodeGoStmt
	case *ast.Ident:
		return nodeIdent
	case *ast.IfStmt:
		return nodeIfStmt
	case *ast.ImportSpec:
		return nodeImportSpec
	case *ast.IncDecStmt:
		return nodeIncDecStmt
	case *ast.IndexExpr:
		return nodeIndexExpr
	case *ast.InterfaceType:
		return nodeInterfaceType
	case *ast.KeyValueExpr:
		return nodeKeyValueExpr
	case *ast.LabeledStmt:
		return nodeLabeledStmt
	case *ast.MapType:
		return nodeMapType
	case *ast.ParenExpr:
		return nodeParenExpr
	case *ast.RangeStmt:
		return nodeRangeStmt
	case *ast.ReturnStmt:
		return nodeReturnStmt
	case *ast.SelectStmt:
		return nodeSelectStmt
	case *ast.SelectorExpr:
		return nodeSelectorExpr
	case *ast.SendStmt:
		return nodeSendStmt
	case *ast.SliceExpr:
		return nodeSliceExpr
	case *ast.StarExpr:
		return nodeStarExpr
	case *ast.StructType:
		return nodeStructType
	case *ast.SwitchStmt:
		return nodeSwitchStmt
	case *ast.TypeAssertExpr:
		return nodeTypeAssertExpr
	case *ast.TypeSpec:
		return nodeTypeSpec
	case *ast.TypeSwitchStmt:
		return nodeTypeSwitchStmt
	case *ast.UnaryExpr:
		return nodeUnaryExpr
	case *ast.ValueSpec:
		return nodeValueSpec
	default:
		return nodeUnknown
	}
}
