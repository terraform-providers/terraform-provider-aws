// Package AWSAT006 defines an Analyzer that checks for
// hardcoded AWS partition DNS suffixes
package AWSAT006

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const Doc = `check for hardcoded AWS partition DNS suffixes

The AWSAT006 analyzer reports hardcoded AWS partition DNS suffixes. For tests
to work across AWS partitions, the partition DNS suffixes should not be
hardcoded.
`

const analyzerName = "AWSAT006"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		inspect.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.BasicLit)(nil),
	}

	var suffixes []string
	for _, p := range endpoints.DefaultPartitions() {
		suffixes = append(suffixes, p.DNSSuffix())
	}

	re := regexp.MustCompile(strings.Join(suffixes, "|"))
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		x := n.(*ast.BasicLit)

		if ignorer.ShouldIgnore(analyzerName, x) {
			return
		}

		if x.Kind != token.STRING {
			return
		}

		if !re.MatchString(x.Value) {
			return
		}

		pass.Reportf(x.ValuePos, "%s: avoid hardcoding AWS partition DNS suffixes, instead use the aws_partition data source", analyzerName)
	})
	return nil, nil
}
