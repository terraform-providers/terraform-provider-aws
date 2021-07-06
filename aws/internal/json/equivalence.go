package json

import (
	"bytes"
	"encoding/json"
	"reflect"
)

// BytesEqual compares two arrays of JSON bytes and returns true if the unmarshaled objects represented by the bytes
// are equal according to `reflect.DeepEqual`.
func BytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func StringsEquivalent(s1, s2 string) bool {
	b1 := bytes.NewBufferString("")
	if err := json.Compact(b1, []byte(s1)); err != nil {
		return false
	}

	b2 := bytes.NewBufferString("")
	if err := json.Compact(b2, []byte(s2)); err != nil {
		return false
	}

	return BytesEqual(b1.Bytes(), b2.Bytes())
}
