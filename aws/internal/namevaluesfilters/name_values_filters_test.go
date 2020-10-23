package namevaluesfilters

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
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

func TestNameValuesFiltersAdd(t *testing.T) {
	testCases := []struct {
		name    string
		filters NameValuesFilters
		add     interface{}
		want    map[string][]string
	}{
		{
			name:    "empty",
			filters: New(map[string][]string{}),
			add:     nil,
			want:    map[string][]string{},
		},
		{
			name: "add_all",
			filters: New(map[string]string{
				"name1": "value1",
				"name2": "value2",
				"name3": "value3",
			}),
			add: New(map[string][]string{
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
			add: map[string]string{
				"name1": "value1b",
				"name3": "value3",
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
			add: map[string][]string{
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
			got := testCase.filters.Add(testCase.add)

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
