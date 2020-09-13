package jsonutil

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestJSONUtil_DecodeJSONFromReader(t *testing.T) {
	input := `{"test":"data","validation":"process"}`

	var actual map[string]interface{}

	err := DecodeJSONFromReader(bytes.NewReader([]byte(input)), &actual)
	if err != nil {
		fmt.Printf("decoding err: %v\n", err)
	}

	expected := map[string]interface{}{
		"test":       "data",
		"validation": "process",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: expected:%#v\nactual:%#v", expected, actual)
	}
}
