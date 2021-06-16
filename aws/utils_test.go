package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

var base64encodingTests = []struct {
	in  []byte
	out string
}{
	// normal encoding case
	{[]byte("data should be encoded"), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
	// base64 encoded input should result in no change of output
	{[]byte("ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="), "ZGF0YSBzaG91bGQgYmUgZW5jb2RlZA=="},
}

func TestBase64Encode(t *testing.T) {
	for _, tt := range base64encodingTests {
		out := base64Encode(tt.in)
		if out != tt.out {
			t.Errorf("base64Encode(%s) => %s, want %s", tt.in, out, tt.out)
		}
	}
}

func TestLooksLikeJsonString(t *testing.T) {
	looksLikeJson := ` {"abc":"1"} `
	doesNotLookLikeJson := `abc: 1`

	if !looksLikeJsonString(looksLikeJson) {
		t.Errorf("Expected looksLikeJson to return true for %s", looksLikeJson)
	}
	if looksLikeJsonString(doesNotLookLikeJson) {
		t.Errorf("Expected looksLikeJson to return false for %s", doesNotLookLikeJson)
	}
}

func TestJsonBytesEqualQuotedAndUnquoted(t *testing.T) {
	unquoted := `{"test": "test"}`
	quoted := "{\"test\": \"test\"}"

	if !jsonBytesEqual([]byte(unquoted), []byte(quoted)) {
		t.Errorf("Expected jsonBytesEqual to return true for %s == %s", unquoted, quoted)
	}

	unquotedDiff := `{"test": "test"}`
	quotedDiff := "{\"test\": \"tested\"}"

	if jsonBytesEqual([]byte(unquotedDiff), []byte(quotedDiff)) {
		t.Errorf("Expected jsonBytesEqual to return false for %s == %s", unquotedDiff, quotedDiff)
	}
}

func TestJsonBytesEqualWhitespaceAndNoWhitespace(t *testing.T) {
	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !jsonBytesEqual([]byte(noWhitespace), []byte(whitespace)) {
		t.Errorf("Expected jsonBytesEqual to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if jsonBytesEqual([]byte(noWhitespaceDiff), []byte(whitespaceDiff)) {
		t.Errorf("Expected jsonBytesEqual to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
	}
}

func TestStringSlicesEqualIgnoreOrder(t *testing.T) {
	equal := []interface{}{
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"b", "a", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"tomato", "apple", "carrot"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Barrier", "Application", "Donut", "Chilly"},
		},
	}
	for _, v := range equal {
		if !StringSlicesEqualIgnoreOrder(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
			t.Fatalf("%v should be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}

	notEqual := []interface{}{
		[]interface{}{
			[]string{"c", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"b", "a", "c"},
			[]string{"a", "bread", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"tomato", "apple"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Barrier", "Applications", "Donut", "Chilly"},
		},
		[]interface{}{
			[]string{},
			[]string{"Barrier", "Applications", "Donut", "Chilly"},
		},
	}
	for _, v := range notEqual {
		if StringSlicesEqualIgnoreOrder(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
			t.Fatalf("%v should not be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}
}

func TestStringSlicesEqual(t *testing.T) {
	equal := []interface{}{
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
		},
		[]interface{}{
			[]string{"b", "a", "c"},
			[]string{"b", "a", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"apple", "carrot", "tomato"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Application", "Barrier", "Chilly", "Donut"},
		},
		[]interface{}{
			[]string{},
			[]string{},
		},
	}
	for _, v := range equal {
		if !StringSlicesEqual(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
			t.Fatalf("%v should be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}

	notEqual := []interface{}{
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"a", "b"},
		},
		[]interface{}{
			[]string{"a", "b", "c"},
			[]string{"b", "a", "c"},
		},
		[]interface{}{
			[]string{"apple", "carrot", "tomato"},
			[]string{"apple", "carrot", "tomato", "zucchini"},
		},
		[]interface{}{
			[]string{"Application", "Barrier", "Chilly", "Donut"},
			[]string{"Application", "Barrier", "Chilly", "Donuts"},
		},
		[]interface{}{
			[]string{},
			[]string{"Application", "Barrier", "Chilly", "Donuts"},
		},
	}
	for _, v := range notEqual {
		if StringSlicesEqual(aws.StringSlice(v.([]interface{})[0].([]string)), aws.StringSlice(v.([]interface{})[1].([]string))) {
			t.Fatalf("%v should not be equal: %v", v.([]interface{})[0].([]string), v.([]interface{})[1].([]string))
		}
	}
}
