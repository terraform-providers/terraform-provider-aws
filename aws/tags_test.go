package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestDiffTags(t *testing.T) {
	cases := []struct {
		Old, New       map[string]interface{}
		Create, Remove map[string]string
	}{
		// Add
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Create: map[string]string{
				"bar": "baz",
			},
			Remove: map[string]string{},
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

		// Overlap
		{
			Old: map[string]interface{}{
				"foo":   "bar",
				"hello": "world",
			},
			New: map[string]interface{}{
				"foo":   "baz",
				"hello": "world",
			},
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
		},

		// Remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			New: map[string]interface{}{
				"foo": "bar",
			},
			Create: map[string]string{},
			Remove: map[string]string{
				"bar": "baz",
			},
		},
	}

	for i, tc := range cases {
		c, r := diffTags(tagsFromMap(tc.Old), tagsFromMap(tc.New))
		cm := tagsToMap(c)
		rm := tagsToMap(r)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

func TestIgnoringTags(t *testing.T) {
	var ignoredTags []*ec2.Tag
	ignoredTags = append(ignoredTags, &ec2.Tag{

		Key:   aws.String("aws:cloudformation:logical-id"),
		Value: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &ec2.Tag{
		Key:   aws.String("aws:foo:bar"),
		Value: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnored(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.Key, *tag.Value)
		}
	}
}

func TestTagsMapToHash(t *testing.T) {
	cases := []struct {
		Left, Right map[string]interface{}
		MustBeEqual bool
	}{
		{
			Left:        map[string]interface{}{},
			Right:       map[string]interface{}{},
			MustBeEqual: true,
		},
		{
			Left: map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
			Right: map[string]interface{}{
				"bar": "baz",
				"foo": "bar",
			},
			MustBeEqual: true,
		},
		{
			Left: map[string]interface{}{
				"foo": "bar",
			},
			Right: map[string]interface{}{
				"bar": "baz",
			},
			MustBeEqual: false,
		},
	}

	for i, tc := range cases {
		l := tagsMapToHash(tc.Left)
		r := tagsMapToHash(tc.Right)
		if tc.MustBeEqual && (l != r) {
			t.Fatalf("%d: Hashes don't match", i)
		}
		if !tc.MustBeEqual && (l == r) {
			t.Logf("%d: Hashes match", i)
		}
	}
}

// testAccCheckTags can be used to check the tags on a resource.
func testAccCheckTags(
	ts *[]*ec2.Tag, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := tagsToMap(*ts)
		v, ok := m[key]
		if value != "" && !ok {
			return fmt.Errorf("Missing tag: %s", key)
		} else if value == "" && ok {
			return fmt.Errorf("Extra tag: %s", key)
		}
		if value == "" {
			return nil
		}

		if v != value {
			return fmt.Errorf("%s: bad value: %s", key, v)
		}

		return nil
	}
}
