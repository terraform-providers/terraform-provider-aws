package nullable

import (
	"errors"
	"regexp"
	"strconv"
	"testing"
)

func TestNullableInt(t *testing.T) {
	cases := []struct {
		val           string
		expectedNull  bool
		expectedValue int64
		expectedErr   error
	}{
		{
			val:           "1",
			expectedNull:  false,
			expectedValue: 1,
		},
		{
			val:           "",
			expectedNull:  true,
			expectedValue: 0,
		},
		{
			val:           "A",
			expectedNull:  false,
			expectedValue: 0,
			expectedErr:   strconv.ErrSyntax,
		},
	}

	for i, tc := range cases {
		v := Int(tc.val)

		if null := v.IsNull(); null != tc.expectedNull {
			t.Fatalf("expected test case %d IsNull to return %t, got %t", i, null, tc.expectedNull)
		}

		value, null, err := v.Value()
		if value != tc.expectedValue {
			t.Fatalf("expected test case %d Value to be %d, got %d", i, tc.expectedValue, value)
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

func TestValidationInt(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateTypeStringNullableInt,
		},
		{
			val:         "A",
			f:           ValidateTypeStringNullableInt,
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as int: .*`),
		},
		{
			val:         1,
			f:           ValidateTypeStringNullableInt,
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationIntAtLeast(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateTypeStringNullableIntAtLeast(1),
		},
		{
			val: "1",
			f:   ValidateTypeStringNullableIntAtLeast(0),
		},
		{
			val:         "1",
			f:           ValidateTypeStringNullableIntAtLeast(2),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at least \(2\), got 1`),
		},
		{
			val:         1,
			f:           ValidateTypeStringNullableIntAtLeast(2),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}
