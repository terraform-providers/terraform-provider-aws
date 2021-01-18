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
		expectedValue int64
		expectedNull  bool
		expectedErr   error
	}{
		{
			val:           "1",
			expectedValue: 1,
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
			val: "",
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

func TestValidationIntBetween(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateIntBetween(1, 1),
		},
		{
			val: "1",
			f:   ValidateIntBetween(0, 2),
		},
		{
			val: "",
			f:   ValidateIntBetween(2, 3),
		},
		{
			val:         "1",
			f:           ValidateIntBetween(2, 3),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be in the range \(2 - 3\), got 1`),
		},
		{
			val:         "A",
			f:           ValidateIntBetween(2, 3),
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as int: .*`),
		},
		{
			val:         1,
			f:           ValidateIntBetween(2, 3),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationIntAtLeast(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateIntAtLeast(1),
		},
		{
			val: "1",
			f:   ValidateIntAtLeast(0),
		},
		{
			val: "",
			f:   ValidateIntAtLeast(2),
		},
		{
			val:         "1",
			f:           ValidateIntAtLeast(2),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at least \(2\), got 1`),
		},
		{
			val:         "A",
			f:           ValidateIntAtLeast(2),
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as int: .*`),
		},
		{
			val:         1,
			f:           ValidateIntAtLeast(2),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}

func TestValidationIntAtMost(t *testing.T) {
	runTestCases(t, []testCase{
		{
			val: "1",
			f:   ValidateIntAtMost(1),
		},
		{
			val: "1",
			f:   ValidateIntAtMost(2),
		},
		{
			val: "",
			f:   ValidateIntAtMost(0),
		},
		{
			val:         "1",
			f:           ValidateIntAtMost(0),
			expectedErr: regexp.MustCompile(`expected [\w]+ to be at most \(0\), got 1`),
		},
		{
			val:         "A",
			f:           ValidateIntAtMost(0),
			expectedErr: regexp.MustCompile(`[\w]+: cannot parse 'A' as int: .*`),
		},
		{
			val:         1,
			f:           ValidateIntAtMost(0),
			expectedErr: regexp.MustCompile(`expected type of [\w]+ to be string`),
		},
	})
}
