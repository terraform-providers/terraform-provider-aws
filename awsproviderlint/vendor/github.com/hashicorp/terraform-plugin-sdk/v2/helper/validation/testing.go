package validation

import (
	"regexp"

	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testCase struct {
	val         interface{}
	f           schema.SchemaValidateFunc
	expectedErr *regexp.Regexp
}

type diagTestCase struct {
	val         interface{}
	f           schema.SchemaValidateDiagFunc
	expectedErr *regexp.Regexp
}

func runTestCases(t testing.T, cases []testCase) {
	t.Helper()

	for i, tc := range cases {
		_, errs := tc.f(tc.val, "test_property")

		if len(errs) == 0 && tc.expectedErr == nil {
			continue
		}

		if len(errs) != 0 && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, errs)
		}

		if !matchAnyError(errs, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, errs)
		}
	}
}

func matchAnyError(errs []error, r *regexp.Regexp) bool {
	// err must match one provided
	for _, err := range errs {
		if r.MatchString(err.Error()) {
			return true
		}
	}
	return false
}

func runDiagTestCases(t testing.T, cases []diagTestCase) {
	t.Helper()

	for i, tc := range cases {
		p := cty.Path{
			cty.GetAttrStep{Name: "test_property"},
		}
		diags := tc.f(tc.val, p)

		if !diags.HasError() && tc.expectedErr == nil {
			continue
		}

		if diags.HasError() && tc.expectedErr == nil {
			t.Fatalf("expected test case %d to produce no errors, got %v", i, diags)
		}

		if !matchAnyDiagSummary(diags, tc.expectedErr) {
			t.Fatalf("expected test case %d to produce error matching \"%s\", got %v", i, tc.expectedErr, diags)
		}
	}
}

func matchAnyDiagSummary(ds diag.Diagnostics, r *regexp.Regexp) bool {
	for _, d := range ds {
		if r.MatchString(d.Summary) {
			return true
		}
	}
	return false
}
