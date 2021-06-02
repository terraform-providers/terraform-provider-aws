package nullable

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
)

func TestNullableBool(t *testing.T) {
	cases := []struct {
		val           string
		expectNull    bool
		expectedValue bool
		expectedErr   error
	}{
		{
			val:           "true",
			expectNull:    false,
			expectedValue: true,
		},
		{
			val:           "false",
			expectNull:    false,
			expectedValue: false,
		},
		{
			val:           "",
			expectNull:    true,
			expectedValue: false,
		},
		{
			val:           "A",
			expectNull:    false,
			expectedValue: false,
			expectedErr:   strconv.ErrSyntax,
		},
	}

	for i, tc := range cases {
		v := Bool(tc.val)

		if null := v.IsNull(); null != tc.expectNull {
			t.Fatalf("expected test case %d IsNull to return %t, got %t", i, null, tc.expectNull)
		}

		value, null, err := v.Value()
		if value != tc.expectedValue {
			t.Fatalf("expected test case %d Value to be %t, got %t", i, tc.expectedValue, value)
		}
		if null != tc.expectNull {
			t.Fatalf("expected test case %d Value null flag to be %t, got %t", i, tc.expectNull, null)
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

func TestValidationBool(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "true",
			f:   ValidateTypeStringNullableBool,
		},
		{
			val:         "A",
			f:           ValidateTypeStringNullableBool,
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as boolean: .*`),
		},
		{
			val:         1,
			f:           ValidateTypeStringNullableBool,
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}
