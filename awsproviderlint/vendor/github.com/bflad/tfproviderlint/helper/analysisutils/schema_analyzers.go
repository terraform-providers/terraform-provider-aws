package analysisutils

import (
	"fmt"

	"github.com/bflad/tfproviderlint/passes/commentignore"
	"github.com/bflad/tfproviderlint/passes/helper/schema/schemainfo"
	"golang.org/x/tools/go/analysis"
)

// SchemaAttributeReferencesAnalyzer returns an Analyzer for fields that use schema attribute references
func SchemaAttributeReferencesAnalyzer(analyzerName string, fieldName string) *analysis.Analyzer {
	doc := fmt.Sprintf(`check for Schema with invalid %[2]s references

The %[1]s analyzer ensures schema attribute references in the Schema %[2]s
field use valid syntax. The Terraform Plugin SDK can unit test attribute
references to verify the references against the full schema.
`, analyzerName, fieldName)

	return &analysis.Analyzer{
		Name: analyzerName,
		Doc:  doc,
		Requires: []*analysis.Analyzer{
			commentignore.Analyzer,
			schemainfo.Analyzer,
		},
		Run: SchemaAttributeReferencesRunner(analyzerName, fieldName),
	}
}
