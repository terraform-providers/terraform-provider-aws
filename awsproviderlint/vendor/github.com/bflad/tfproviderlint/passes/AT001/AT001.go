// Package AT001 defines an Analyzer that checks for
// TestCase missing CheckDestroy
package AT001

import (
	"flag"
	"go/ast"
	"path/filepath"
	"strings"

	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/resource/testcaseinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for TestCase missing CheckDestroy

The AT001 analyzer reports likely incorrect uses of TestCase
which do not define a CheckDestroy function. CheckDestroy is used to verify
that test infrastructure has been removed at the end of an acceptance test.

Optional parameters:
  - ignored-filename-prefixes Comma-separated list of filename prefixes to ignore, defaults to 'data_source_'.
  - ignored-filename-suffixes Comma-separated list of filename suffixes to ignore, defaults to none.

More information can be found at:
https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy`

const analyzerName = "AT001"

var (
	ignoredSuffixes string
	ignoredPrefixes string
)

func parseFlags() flag.FlagSet {
	var flags = flag.NewFlagSet(analyzerName, flag.ExitOnError)
	flags.StringVar(&ignoredPrefixes, "ignored-filename-prefixes", "data_source_", "Comma-separated list of file name prefixes to ignore")
	flags.StringVar(&ignoredSuffixes, "ignored-filename-suffixes", "", "Comma-separated list of file name suffixes to ignore")
	return *flags
}


var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Flags: parseFlags(),
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testcaseinfo.Analyzer,
	},
	Run: run,
}

func isSuffixIgnored(fileName string, suffixesList string) bool {
	suffixes := strings.Split(suffixesList, ",")

	for _, suffix := range suffixes {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}
	return false
}

func isPrefixIgnored(fileName string, prefixesList string) bool {
	prefixes := strings.Split(prefixesList, ",")

	for _, prefix := range prefixes {
		if strings.HasPrefix(fileName, prefix) {
			return true
		}
	}
	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	testCases := pass.ResultOf[testcaseinfo.Analyzer].([]*resource.TestCaseInfo)

	for _, testCase := range testCases {
		fileName := filepath.Base(pass.Fset.File(testCase.AstCompositeLit.Pos()).Name())

		if ignoredPrefixes != "" && isPrefixIgnored(fileName, ignoredPrefixes) {
			continue
		}

		if ignoredSuffixes != "" && isSuffixIgnored(fileName, ignoredSuffixes) {
			continue
		}

		if ignorer.ShouldIgnore(analyzerName, testCase.AstCompositeLit) {
			continue
		}

		if testCase.DeclaresField(resource.TestCaseFieldCheckDestroy) {
			continue
		}

		pass.Reportf(testCase.AstCompositeLit.Type.(*ast.SelectorExpr).Sel.Pos(), "%s: missing CheckDestroy", analyzerName)
	}

	return nil, nil
}
