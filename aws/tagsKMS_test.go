package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

// go test -v -run="TestDiffKMSTags"
func TestDiffKMSTags(t *testing.T) {
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
		c, r := diffTagsKMS(tagsFromMapKMS(tc.Old), tagsFromMapKMS(tc.New))
		cm := tagsToMapKMS(c)
		rm := tagsToMapKMS(r)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

// go test -v -run="TestIgnoringTagsKMS"
func TestIgnoringTagsKMS(t *testing.T) {
	var ignoredTags []*kms.Tag
	ignoredTags = append(ignoredTags, &kms.Tag{
		TagKey:   aws.String("aws:cloudformation:logical-id"),
		TagValue: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &kms.Tag{
		TagKey:   aws.String("aws:foo:bar"),
		TagValue: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnoredKMS(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.TagKey, *tag.TagValue)
		}
	}
}
