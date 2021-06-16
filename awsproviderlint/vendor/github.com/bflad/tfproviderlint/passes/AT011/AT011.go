package AT011

import (
	"github.com/bflad/tfproviderlint/helper/terraformtype/helper/resource"
	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/resource/testcaseinfo"
	"golang.org/x/tools/go/analysis"
)

const Doc = `check for TestCase including IDRefreshIgnore implementation without IDRefreshName

The AT011 analyzer reports likely extraneous use of ID-only refresh testing.
Most resources should prefer to include a TestStep with ImportState instead
since it will cover the same testing functionality along with verifying
resource import support.

However for cases where IDRefreshName is being already being used, the
IDRefreshIgnore field is considered valid. If IDRefreshName is not being used,
then this analyzer will return a report.
`

const analyzerName = "AT011"

var Analyzer = &analysis.Analyzer{
	Name: analyzerName,
	Doc:  Doc,
	Requires: []*analysis.Analyzer{
		commentignore.Analyzer,
		testcaseinfo.Analyzer,
	},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	ignorer := pass.ResultOf[commentignore.Analyzer].(*commentignore.Ignorer)
	testCases := pass.ResultOf[testcaseinfo.Analyzer].([]*resource.TestCaseInfo)

	for _, testCase := range testCases {
		field, ok := testCase.Fields[resource.TestCaseFieldIDRefreshIgnore]

		if !ok || field == nil || testCase.DeclaresField(resource.TestCaseFieldIDRefreshName) {
			continue
		}

		if ignorer.ShouldIgnore(analyzerName, field) {
			continue
		}

		pass.Reportf(field.Pos(), "%s: extraneous TestCase IDRefreshIgnore without IDRefreshName", analyzerName)
	}

	return nil, nil
}
