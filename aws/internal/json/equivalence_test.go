package json

import (
	"testing"
)

func TestBytesEqualQuotedAndUnquoted(t *testing.T) {
	unquoted := `{"test": "test"}`
	quoted := "{\"test\": \"test\"}"

	if !BytesEqual([]byte(unquoted), []byte(quoted)) {
		t.Errorf("Expected BytesEqual to return true for %s == %s", unquoted, quoted)
	}

	unquotedDiff := `{"test": "test"}`
	quotedDiff := "{\"test\": \"tested\"}"

	if BytesEqual([]byte(unquotedDiff), []byte(quotedDiff)) {
		t.Errorf("Expected BytesEqual to return false for %s == %s", unquotedDiff, quotedDiff)
	}
}

func TestBytesEqualWhitespaceAndNoWhitespace(t *testing.T) {
	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !BytesEqual([]byte(noWhitespace), []byte(whitespace)) {
		t.Errorf("Expected BytesEqual to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if BytesEqual([]byte(noWhitespaceDiff), []byte(whitespaceDiff)) {
		t.Errorf("Expected BytesEqual to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
	}
}

func TestStringsEquivalentWhitespaceAndNoWhitespace(t *testing.T) {
	noWhitespace := `{"test":"test"}`
	whitespace := `
{
  "test": "test"
}`

	if !StringsEquivalent(noWhitespace, whitespace) {
		t.Errorf("Expected StringsEquivalent to return true for %s == %s", noWhitespace, whitespace)
	}

	noWhitespaceDiff := `{"test":"test"}`
	whitespaceDiff := `
{
  "test": "tested"
}`

	if StringsEquivalent(noWhitespaceDiff, whitespaceDiff) {
		t.Errorf("Expected StringsEquivalent to return false for %s == %s", noWhitespaceDiff, whitespaceDiff)
	}
}
