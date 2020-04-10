package passes

import (
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestValidateAllChecks(t *testing.T) {
	err := analysis.Validate(AllChecks)

	if err != nil {
		t.Fatal(err)
	}
}
