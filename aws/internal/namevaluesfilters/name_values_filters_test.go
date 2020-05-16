package namevaluesfilters

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestNameValuesFiltersMap(t *testing.T) {
	testCases := []struct {
		name    string
		filters NameValuesFilters
		want    map[string][]string
	}{
		{
			name:    "empty",
			filters: New(map[string][]string{}),
			want:    map[string][]string{},
		},
		{
			name: "empty_strings",
			filters: New(map[string][]string{
				"name1": {""},
				"name2": {"", ""},
			}),
			want: map[string][]string{},
		},
		{
			name: "duplicates",
			filters: New(map[string][]string{
				"name1": {"value1"},
				"name2": {"value2a", "value2b", "", "value2a", "value2c", "value2c"},
			}),
			want: map[string][]string{
				"name1": {"value1"},
				"name2": {"value2a", "value2b", "value2c"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.filters.Map()

			testNameValuesFiltersVerifyMap(t, got, testCase.want)
		})
	}
}

func TestNameValuesFiltersMerge(t *testing.T) {
	testCases := []struct {
		name         string
		filters      NameValuesFilters
		mergeFilters NameValuesFilters
		want         map[string][]string
	}{
		{
			name:         "empty",
			filters:      New(map[string][]string{}),
			mergeFilters: New(map[string]string{}),
			want:         map[string][]string{},
		},
		{
			name: "add_all",
			filters: New(map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			}),
			mergeFilters: New(map[string][]string{
				"name4": {"value4a", "value4b"},
				"name5": {"value5"},
				"name6": {"value6a", "value6b", "value6c"},
			}),
			want: map[string][]string{
				"name1": {"value1"},
				"name2": {"value2"},
				"name3": {"value3"},
				"name4": {"value4a", "value4b"},
				"name5": {"value5"},
				"name6": {"value6a", "value6b", "value6c"},
			},
		},
		{
			name: "mixed",
			filters: New(map[string][]string{
				"name1": {"value1a"},
				"name2": {"value2a", "value2b"},
			}),
			mergeFilters: map[string][]string{
				"name1": {"value1b"},
				"name3": {"value3"},
			},
			want: map[string][]string{
				"name1": {"value1a", "value1b"},
				"name2": {"value2a", "value2b"},
				"name3": {"value3"},
			},
		},
		{
			name: "from_set",
			filters: New(schema.NewSet(testNameValuesFiltersHashSet, []interface{}{
				map[string]interface{}{
					"name": "name1",
					"values": schema.NewSet(schema.HashString, []interface{}{
						"value1",
					}),
				},
				map[string]interface{}{
					"name": "name2",
					"values": schema.NewSet(schema.HashString, []interface{}{
						"value2a",
						"value2b",
					}),
				},
				map[string]interface{}{
					"name": "name3",
					"values": schema.NewSet(schema.HashString, []interface{}{
						"value3",
					}),
				},
			})),
			mergeFilters: map[string][]string{
				"name1": {"value1"},
				"name2": {"value2c"},
			},
			want: map[string][]string{
				"name1": {"value1"},
				"name2": {"value2a", "value2b", "value2c"},
				"name3": {"value3"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.filters.Merge(testCase.mergeFilters)

			testNameValuesFiltersVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func testNameValuesFiltersVerifyMap(t *testing.T, got map[string][]string, want map[string][]string) {
	for k, wantV := range want {
		gotV, ok := got[k]

		if !ok {
			t.Errorf("want missing name: %s", k)
			continue
		}

		if !reflect.DeepEqual(gotV, wantV) {
			t.Errorf("got name (%s) values %s; want values %s", k, gotV, wantV)
		}
	}

	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("got extra name: %s", k)
		}
	}
}

func testNameValuesFiltersHashSet(v interface{}) int {
	m := v.(map[string]interface{})
	return hashcode.String(m["name"].(string))
}
