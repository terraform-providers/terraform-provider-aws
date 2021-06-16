package V011

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemavalidatefuncinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for custom SchemaValidateFunc that implement validation.StringLenBetween()

The V011 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with validation.StringLenBetween().`

const analyzerName = "V011"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		schemavalidatefuncinfo.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	schemaValidateFuncs := pass.ResultOf[schemavalidatefuncinfo.Analyzer].([]*schema.SchemaValidateFuncInfo)

	for _, schemaValidateFunc := range schemaValidateFuncs {
		if ignorer.ShouldIgnore(analyzerName, schemaValidateFunc.Node) {
			continue
		}

		if !hasIfStringLenCheck(schemaValidateFunc.Body, pass.TypesInfo) {
			continue
		}

		pass.Reportf(schemaValidateFunc.Pos, "%s: custom SchemaValidateFunc should be replaced with validation.StringLenBetween()", analyzerName)
	}

	return nil, nil
}

func hasIfStringLenCheck(node ast.Node, info *types.Info) bool {
	result := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		default:
			return true
		case *ast.IfStmt:
			if !hasStringLenCheck(n, info) {
				return true
			}

			result = true

			return false
		}
	})

	return result
}

func hasStringLenCheck(node ast.Node, info *types.Info) bool {
	result := false

	ast.Inspect(node, func(n ast.Node) bool {
		binaryExpr, ok := n.(*ast.BinaryExpr)

		if !ok {
			return true
		}

		if !exprIsStringLenCallExpr(binaryExpr.X, info) && !exprIsStringLenCallExpr(binaryExpr.Y, info) {
			return true
		}

		if !tokenIsLenCheck(binaryExpr.Op) {
			return true
		}

		result = true

		return false
	})

	return result
}

func exprIsStringLenCallExpr(e ast.Expr, info *types.Info) bool {
	switch e := e.(type) {
	default:
		return false
	case *ast.CallExpr:
		switch fun := e.Fun.(type) {
		default:
			return false
		case *ast.Ident:
			if fun.Name != "len" {
				return false
			}
		}

		if len(e.Args) != 1 {
			return false
		}

		switch arg := info.TypeOf(e.Args[0]).Underlying().(type) {
		default:
			return false
		case *types.Basic:
			return arg.Kind() == types.String
		}
	}
}

func tokenIsLenCheck(t token.Token) bool {
	validTokens := []token.Token{
		token.GEQ, // >=
		token.GTR, // >
		token.LEQ, // <=
		token.LSS, // <
	}

	for _, validToken := range validTokens {
		if t == validToken {
			return true
		}
	}

	return false
}
