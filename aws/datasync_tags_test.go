package aws

import (
	"reflect"
	"testing"
)

func TestDataSyncTagsDiff(t *testing.T) {
	cases := []struct {
		Old, New       map[string]interface{}
		Create, Remove map[string]string
	}{
		// Basic add/remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"bar": "baz",
			},
			Create: map[string]string{
				"bar": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
		},

		// Modify
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "baz",
			},
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
		},
	}

	for i, tc := range cases {
		create, remove := dataSyncTagsDiff(expandDataSyncTagListEntry(tc.Old), expandDataSyncTagListEntry(tc.New))
		createMap := flattenDataSyncTagListEntry(create)
		removeMap := flattenDataSyncTagListEntry(remove)
		if !reflect.DeepEqual(createMap, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, createMap)
		}
		if !reflect.DeepEqual(removeMap, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, removeMap)
		}
	}
}
