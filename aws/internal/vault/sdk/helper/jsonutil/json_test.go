package jsonutil

import (
	"bytes"
	"reflect"
	"testing"
)

func TestJSONUtil_DecodeJSONFromReader(t *testing.T) {
	input := `{"test":"data","validation":"process"}`

	var actual map[string]interface{}

	err := DecodeJSONFromReader(bytes.NewReader([]byte(input)), &actual)
	if err != nil {
		t.Errorf("decoding err: %v", err)
	}

	expected := map[string]interface{}{
		"test":       "data",
		"validation": "process",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: expected:%#v\nactual:%#v", expected, actual)
	}
}
