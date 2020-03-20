package keyvaluetags

import (
	"testing"
)

func TestKeyValueTagsIgnoreAws(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want map[string]string
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(map[string]string{
				"aws:cloudformation:key1": "value1",
				"aws:cloudformation:key2": "value2",
				"aws:cloudformation:key3": "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(map[string]string{
				"aws:cloudformation:key1": "value1",
				"key2":                    "value2",
				"key3":                    "value3",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "none",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.IgnoreAws()

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnoreElasticbeanstalk(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want map[string]string
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(map[string]string{
				"aws:cloudformation:key1": "value1",
				"elasticbeanstalk:key2":   "value2",
				"Name":                    "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(map[string]string{
				"aws:cloudformation:key1": "value1",
				"key2":                    "value2",
				"elasticbeanstalk:key3":   "value3",
				"key4":                    "value4",
				"Name":                    "value5",
			}),
			want: map[string]string{
				"key2": "value2",
				"key4": "value4",
			},
		},
		{
			name: "none",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.IgnoreElasticbeanstalk()

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnorePrefixes(t *testing.T) {
	testCases := []struct {
		name              string
		tags              KeyValueTags
		ignoreTagPrefixes KeyValueTags
		want              map[string]string
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			ignoreTagPrefixes: New([]string{
				"key1",
				"key2",
				"key3",
			}),
			want: map[string]string{},
		},
		{
			name: "all_exact",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New([]string{
				"key1",
				"key2",
				"key3",
			}),
			want: map[string]string{},
		},
		{
			name: "all_prefix",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New([]string{
				"key",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New([]string{
				"key1",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "none",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTagPrefixes: New([]string{
				"key4",
				"key5",
				"key6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.IgnorePrefixes(testCase.ignoreTagPrefixes)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnoreRds(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want map[string]string
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(map[string]string{
				"aws:cloudformation:key1": "value1",
				"rds:key2":                "value2",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(map[string]string{
				"aws:cloudformation:key1": "value1",
				"key2":                    "value2",
				"rds:key3":                "value3",
				"key4":                    "value4",
			}),
			want: map[string]string{
				"key2": "value2",
				"key4": "value4",
			},
		},
		{
			name: "none",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.IgnoreRds()

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsIgnore(t *testing.T) {
	testCases := []struct {
		name       string
		tags       KeyValueTags
		ignoreTags KeyValueTags
		want       map[string]string
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			ignoreTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "all",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
		{
			name: "mixed",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTags: New(map[string]string{
				"key1": "value1",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "none",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			ignoreTags: New(map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.Ignore(testCase.ignoreTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsKeys(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want []string
	}{
		{
			name: "empty_map_string_interface",
			tags: New(map[string]interface{}{}),
			want: []string{},
		},
		{
			name: "empty_map_string_stringPointer",
			tags: New(map[string]*string{}),
			want: []string{},
		},
		{
			name: "empty_map_string_string",
			tags: New(map[string]string{}),
			want: []string{},
		},
		{
			name: "empty_slice_interface",
			tags: New(map[string]interface{}{}),
			want: []string{},
		},
		{
			name: "empty_slice_string",
			tags: New(map[string]string{}),
			want: []string{},
		},
		{
			name: "non_empty_map_string_interface",
			tags: New(map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_map_string_string",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_map_string_stringPointer",
			tags: New(map[string]*string{
				"key1": testStringPtr("value1"),
				"key2": testStringPtr("value2"),
				"key3": testStringPtr("value3"),
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_slice_interface",
			tags: New([]interface{}{
				"key1",
				"key2",
				"key3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
		{
			name: "non_empty_slice_string",
			tags: New([]string{
				"key1",
				"key2",
				"key3",
			}),
			want: []string{
				"key1",
				"key2",
				"key3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.Keys()

			testKeyValueTagsVerifyKeys(t, got, testCase.want)
		})
	}
}

func TestKeyValueTagsMap(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want map[string]string
	}{
		{
			name: "empty_map_string_interface",
			tags: New(map[string]interface{}{}),
			want: map[string]string{},
		},
		{
			name: "empty_map_string_string",
			tags: New(map[string]string{}),
			want: map[string]string{},
		},
		{
			name: "empty_map_string_stringPointer",
			tags: New(map[string]*string{}),
			want: map[string]string{},
		},
		{
			name: "non_empty_map_string_interface",
			tags: New(map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_map_string_string",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "non_empty_map_string_stringPointer",
			tags: New(map[string]*string{
				"key1": testStringPtr("value1"),
				"key2": testStringPtr("value2"),
				"key3": testStringPtr("value3"),
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.Map()

			testKeyValueTagsVerifyMap(t, got, testCase.want)
		})
	}
}

func TestKeyValueTagsMerge(t *testing.T) {
	testCases := []struct {
		name      string
		tags      KeyValueTags
		mergeTags KeyValueTags
		want      map[string]string
	}{
		{
			name:      "empty",
			tags:      New(map[string]string{}),
			mergeTags: New(map[string]string{}),
			want:      map[string]string{},
		},
		{
			name: "add_all",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			mergeTags: New(map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			},
		},
		{
			name: "mixed",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			mergeTags: New(map[string]string{
				"key1": "value1updated",
				"key4": "value4",
			}),
			want: map[string]string{
				"key1": "value1updated",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			},
		},
		{
			name: "update_all",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			mergeTags: New(map[string]string{
				"key1": "value1updated",
				"key2": "value2updated",
				"key3": "value3updated",
			}),
			want: map[string]string{
				"key1": "value1updated",
				"key2": "value2updated",
				"key3": "value3updated",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.Merge(testCase.mergeTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsRemoved(t *testing.T) {
	testCases := []struct {
		name    string
		oldTags KeyValueTags
		newTags KeyValueTags
		want    map[string]string
	}{
		{
			name:    "empty",
			oldTags: New(map[string]string{}),
			newTags: New(map[string]string{}),
			want:    map[string]string{},
		},
		{
			name: "all_new",
			oldTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "mixed",
			oldTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(map[string]string{
				"key1": "value1",
			}),
			want: map[string]string{
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "no_changes",
			oldTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.oldTags.Removed(testCase.newTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsUpdated(t *testing.T) {
	testCases := []struct {
		name    string
		oldTags KeyValueTags
		newTags KeyValueTags
		want    map[string]string
	}{
		{
			name:    "empty",
			oldTags: New(map[string]string{}),
			newTags: New(map[string]string{}),
			want:    map[string]string{},
		},
		{
			name: "all_new",
			oldTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			}),
			want: map[string]string{
				"key4": "value4",
				"key5": "value5",
				"key6": "value6",
			},
		},
		{
			name: "mixed",
			oldTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(map[string]string{
				"key1": "value1updated",
				"key4": "value4",
			}),
			want: map[string]string{
				"key1": "value1updated",
				"key4": "value4",
			},
		},
		{
			name: "no_changes",
			oldTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			newTags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: map[string]string{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.oldTags.Updated(testCase.newTags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsChunks(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		size int
		want []int
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			size: 10,
			want: []int{},
		},
		{
			name: "chunk_1",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 1,
			want: []int{1, 1, 1, 1},
		},
		{
			name: "chunk_2",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 2,
			want: []int{2, 2},
		},
		{
			name: "chunk_3",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 3,
			want: []int{3, 1},
		},
		{
			name: "chunk_4",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 4,
			want: []int{4},
		},
		{
			name: "chunk_5",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			size: 5,
			want: []int{4},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.Chunks(testCase.size)

			if len(got) != len(testCase.want) {
				t.Errorf("unexpected number of chunks: %d", len(got))
			}

			for i, n := range testCase.want {
				if len(got[i]) != n {
					t.Errorf("chunk (%d) length %d; want length %d", i, len(got[i]), n)
				}
			}
		})
	}
}

func TestKeyValueTagsContainsAll(t *testing.T) {
	testCases := []struct {
		name   string
		source KeyValueTags
		target KeyValueTags
		want   bool
	}{
		{
			name:   "empty",
			source: New(map[string]string{}),
			target: New(map[string]string{}),
			want:   true,
		},
		{
			name:   "source_empty",
			source: New(map[string]string{}),
			target: New(map[string]string{
				"key1": "value1",
			}),
			want: false,
		},
		{
			name: "target_empty",
			source: New(map[string]string{
				"key1": "value1",
			}),
			target: New(map[string]string{}),
			want:   true,
		},
		{
			name: "exact_match",
			source: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
			target: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
			want: true,
		},
		{
			name: "source_contains_all",
			source: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			target: New(map[string]string{
				"key1": "value1",
				"key3": "value3",
			}),
			want: true,
		},
		{
			name: "source_does_not_contain_all",
			source: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			target: New(map[string]string{
				"key1": "value1",
				"key4": "value4",
			}),
			want: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.source.ContainsAll(testCase.target)

			if got != testCase.want {
				t.Errorf("unexpected ContainsAll: %t", got)
			}
		})
	}
}

func TestKeyValueTagsHash(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		zero bool
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			zero: true,
		},
		{
			name: "not_empty",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			}),
			zero: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.Hash()

			if (got == 0 && !testCase.zero) || (got != 0 && testCase.zero) {
				t.Errorf("unexpected hash code: %d", got)
			}
		})
	}
}

func TestKeyValueTagsUrlEncode(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want string
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: "",
		},
		{
			name: "single",
			tags: New(map[string]string{
				"key1": "value1",
			}),
			want: "key1=value1",
		},
		{
			name: "multiple",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: "key1=value1&key2=value2&key3=value3",
		},
		{
			name: "multiple_with_encoded",
			tags: New(map[string]string{
				"key1":  "value 1",
				"key@2": "value+:2",
				"key3":  "value3",
			}),
			want: "key1=value+1&key3=value3&key%402=value%2B%3A2",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.UrlEncode()

			if got != testCase.want {
				t.Errorf("unexpected URL encoded value: %q", got)
			}
		})
	}
}

func testKeyValueTagsVerifyKeys(t *testing.T, got []string, want []string) {
	for _, g := range got {
		found := false

		for _, w := range want {
			if w == g {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("got extra key: %s", g)
		}
	}

	for _, w := range want {
		found := false

		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("want missing key: %s", w)
		}
	}
}

func testKeyValueTagsVerifyMap(t *testing.T, got map[string]string, want map[string]string) {
	for k, wantV := range want {
		gotV, ok := got[k]

		if !ok {
			t.Errorf("want missing key: %s", k)
			continue
		}

		if gotV != wantV {
			t.Errorf("got key (%s) value %s; want value %s", k, gotV, wantV)
		}
	}

	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("got extra key: %s", k)
		}
	}
}

func testStringPtr(str string) *string {
	return &str
}
