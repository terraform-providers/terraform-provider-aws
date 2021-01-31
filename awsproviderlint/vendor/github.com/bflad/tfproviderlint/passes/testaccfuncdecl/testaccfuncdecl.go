package testaccfuncdecl

import (
	"go/ast"
	"reflect"
	"strings"

	"github.com/bflad/tfproviderlint/passes/testfuncdecl"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "testaccfuncdecl",
	Doc:  "find *ast.FuncDecl with TestAcc prefixed names for later passes",
	Requires: []*analysis.Analyzer{
		testfuncdecl.Analyzer,
	},
	Run:        run,
	ResultType: reflect.TypeOf([]*ast.FuncDecl{}),
}

func run(pass *analysis.Pass) (interface{}, error) {
	testFuncs := pass.ResultOf[testfuncdecl.Analyzer].([]*ast.FuncDecl)

	var result []*ast.FuncDecl

	for _, testFunc := range testFuncs {
		if strings.HasPrefix(testFunc.Name.Name, "TestAcc") {
			result = append(result, testFunc)
		}
	}

	return result, nil
}
