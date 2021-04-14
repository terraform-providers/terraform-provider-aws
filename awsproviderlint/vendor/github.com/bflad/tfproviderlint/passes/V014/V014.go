package V014

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

const Doc = `check for custom SchemaValidateFunc that implement validation.IntInSlice() or validation.IntNotInSlice()

The V014 analyzer reports when custom SchemaValidateFunc declarations can be
replaced with validation.IntInSlice() or validation.IntNotInSlice().`

const analyzerName = "V014"

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

		if !hasIfIntEquality(schemaValidateFunc.Body, pass.TypesInfo) {
			continue
		}

		pass.Reportf(schemaValidateFunc.Pos, "%s: custom SchemaValidateFunc should be replaced with validation.IntInSlice() or validation.IntNotInSlice()", analyzerName)
	}

	return nil, nil
}

func hasIfIntEquality(node ast.Node, info *types.Info) bool {
	result := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch n := n.(type) {
		default:
			return true
		case *ast.IfStmt:
			if !hasIntEquality(n, info) {
				return true
			}

			result = true

			return false
		}
	})

	return result
}

func hasIntEquality(node ast.Node, info *types.Info) bool {
	result := false

	ast.Inspect(node, func(n ast.Node) bool {
		binaryExpr, ok := n.(*ast.BinaryExpr)

		if !ok {
			return true
		}

		if !exprIsInt(binaryExpr.X, info) || !exprIsInt(binaryExpr.Y, info) {
			return true
		}

		if !tokenIsEquality(binaryExpr.Op) {
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

func exprIsInt(e ast.Expr, info *types.Info) bool {
	switch e := e.(type) {
	default:
		return false
	case *ast.BasicLit:
		return e.Kind == token.INT
	case *ast.Ident:
		switch t := info.TypeOf(e).Underlying().(type) {
		default:
			return false
		case *types.Basic:
			return t.Kind() == types.Int
		}
	}
}

func tokenIsEquality(t token.Token) bool {
	validTokens := []token.Token{
		token.EQL, // ==
		token.NEQ, // !=
	}

	for _, validToken := range validTokens {
		if t == validToken {
			return true
		}
	}

	return false
}
