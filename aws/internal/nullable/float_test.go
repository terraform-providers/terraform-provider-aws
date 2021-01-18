package nullable

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
)

func TestNullableFloat(t *testing.T) {
	cases := []struct {
		val           string
		expectedValue float64
		expectedNull  bool
		expectedErr   error
	}{
		{
			val:           "1.5",
			expectedValue: 1.5,
			expectedNull:  false,
		},
		{
			val:           "",
			expectedValue: 0,
			expectedNull:  true,
		},
		{
			val:           "A",
			expectedValue: 0,
			expectedNull:  false,
			expectedErr:   strconv.ErrSyntax,
		},
	}

	for i, tc := range cases {
		v := Float(tc.val)

		if null := v.IsNull(); null != tc.expectedNull {
			t.Fatalf("expected test case %d IsNull to return %t, got %t", i, null, tc.expectedNull)
		}

		value, null, err := v.Value()
		if value != tc.expectedValue {
			t.Fatalf("expected test case %d Value to be %f, got %f", i, tc.expectedValue, value)
		}
		if null != tc.expectedNull {
			t.Fatalf("expected test case %d Value null flag to be %t, got %t", i, tc.expectedNull, null)
		}
		if tc.expectedErr == nil && err != nil {
			t.Fatalf("expected test case %d to succeed, got error %s", i, err)
		}
		if tc.expectedErr != nil {
			if !errors.Is(err, tc.expectedErr) {
				t.Fatalf("expected test case %d to have error matching \"%s\", got %s", i, tc.expectedErr, err)
			}
		}
	}
}

func TestValidationFloat(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "1.5",
			f:   ValidateTypeStringNullableFloat,
		},
		{
			val: "",
			f:   ValidateTypeStringNullableFloat,
		},
		{
			val:         "A",
			f:           ValidateTypeStringNullableFloat,
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as float: .*`),
		},
		{
			val:         1.5,
			f:           ValidateTypeStringNullableFloat,
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationFloatBetween(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "1.5",
			f:   ValidateFloatBetween(1.0, 2.0),
		},
		{
			// Test inclusive upper bound
			val: "1.0",
			f:   ValidateFloatBetween(0.0, 1.0),
		},
		{
			// Test inclusive lower bound
			val: "0.0",
			f:   ValidateFloatBetween(0.0, 1.0),
		},
		{
			val: "",
			f:   ValidateFloatBetween(1.0, 2.0),
		},
		{
			val:         "0.5",
			f:           ValidateFloatBetween(1.0, 2.0),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be in the range \(1\.0\d+ - 2\.0\d+\), got 0\.5\d+`),
		},
		{
			val:         "2.5",
			f:           ValidateFloatBetween(1.0, 2.0),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be in the range \(1\.0\d+ - 2\.0\d+\), got 2\.5\d+`),
		},
		{
			val:         "A",
			f:           ValidateFloatBetween(1.0, 2.0),
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as float: .*`),
		},
		{
			val:         1,
			f:           ValidateFloatBetween(1.0, 2.0),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationFloatAtLeast(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "2.5",
			f:   ValidateFloatAtLeast(1.5),
		},
		{
			val: "-1.0",
			f:   ValidateFloatAtLeast(-1.5),
		},
		{
			val: "",
			f:   ValidateFloatAtLeast(2.5),
		},
		{
			val:         "1.5",
			f:           ValidateFloatAtLeast(2.5),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at least \(2\.5\d*\), got 1\.5\d*`),
		},
		{
			val:         "A",
			f:           ValidateFloatAtLeast(2),
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as float: .*`),
		},
		{
			val:         2.5,
			f:           ValidateFloatAtLeast(1.5),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationFloatAtMost(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "2.5",
			f:   ValidateFloatAtMost(3.5),
		},
		{
			val: "-1.0",
			f:   ValidateFloatAtMost(-0.5),
		},
		{
			val: "",
			f:   ValidateFloatAtMost(1.0),
		},
		{
			val:         "2.5",
			f:           ValidateFloatAtMost(1.5),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at most \(1\.5\d+\), got 2\.5\d+`),
		},
		{
			val:         "A",
			f:           ValidateFloatAtMost(0),
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as float: .*`),
		},
		{
			val:         2.5,
			f:           ValidateFloatAtMost(3.5),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}
