package V012

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/bflad/tfproviderlint/helper/astutils"
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/schema"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemavalidatefuncinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for custom SchemaValidateFunc that implement validation.IntAtLeast(), validation.IntAtMost(), or validation.IntBetween()

The V012 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with validation.IntAtLeast(), validation.IntAtMost(), or
validation.IntBetween().`

const analyzerName = "V012"

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

		if hasStrconvAtoiCallExpr(schemaValidateFunc.Body, pass.TypesInfo) {
			continue
		}

		if !hasIfIntCheck(schemaValidateFunc.Body, pass.TypesInfo) {
			continue
		}

		pass.Reportf(schemaValidateFunc.Pos, "%s: custom SchemaValidateFunc should be replaced with validation.IntAtLeast(), validation.IntAtMost(), or validation.IntBetween()", analyzerName)
	}

	return nil, nil
}

func hasIfIntCheck(node ast.Node, info *types.Info) bool {
	result := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		default:
			return true
		case *ast.IfStmt:
			if !hasIntCheck(n, info) {
				return true
			}

			result = true

			return false
		}
	})

	return result
}

func hasIntCheck(node ast.Node, info *types.Info) bool {
	result := false

	ast.Inspect(node, func(n ast.Node) bool {
		binaryExpr, ok := n.(*ast.BinaryExpr)

		if !ok {
			return true
		}

		if !exprIsIntIdent(binaryExpr.X, info) && !exprIsIntIdent(binaryExpr.Y, info) {
			return true
		}

		if !tokenIsIntCheck(binaryExpr.Op) {
			return true
		}

		result = true

		return false
	})

	return result
}

func hasStrconvAtoiCallExpr(node ast.Node, info *types.Info) bool {
	result := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		default:
			return true
		case *ast.CallExpr:
			if !astutils.IsStdlibPackageFunc(n.Fun, info, "strconv", "Atoi") {
				return true
			}

			result = true

			return false
		}
	})

	return result
}

func exprIsIntIdent(e ast.Expr, info *types.Info) bool {
	switch e := e.(type) {
	default:
		return false
	case *ast.Ident:
		switch t := info.TypeOf(e).Underlying().(type) {
		default:
			return false
		case *types.Basic:
			return t.Kind() == types.Int
		}
	}
}

func tokenIsIntCheck(t token.Token) bool {
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
