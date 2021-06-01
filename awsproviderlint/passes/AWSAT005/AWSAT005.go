// Package AWSAT005 defines an Analyzer that checks for
// hardcoded AWS partitions in ARNs
package AWSAT005

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

const Doc = `check for hardcoded AWS partitions in ARNs

The AWSAT005 analyzer reports hardcoded AWS partitions in ARNs. For tests to
work across AWS partitions, the partitions should not be hardcoded.
`

const analyzerName = "AWSAT005"

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

	var partitions []string
	for _, p := range endpoints.DefaultPartitions() {
		partitions = append(partitions, p.ID())
	}

	re := regexp.MustCompile(`arn:(` + strings.Join(partitions, "|") + `):`)
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

		pass.Reportf(x.ValuePos, "%s: avoid hardcoded ARN AWS partitions, use aws_partition data source", analyzerName)
	})
	return nil, nil
}
