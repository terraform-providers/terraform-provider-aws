package aws

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

func TestDiffBeanstalkTags(t *testing.T) {
	cases := []struct {
		Old, New map[string]interface{}
		Create   map[string]string
		Remove   []string
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
			Remove: []string{
				"foo",
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
			Remove: []string{},
		},
	}

	for i, tc := range cases {
		c, r := diffTagsBeanstalk(tagsFromMapBeanstalk(tc.Old), tagsFromMapBeanstalk(tc.New))
		cm := tagsToMapBeanstalk(c)
		rl := []string{}
		for _, tagName := range r {
			rl = append(rl, *tagName)
		}
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rl, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rl)
		}
	}
}

func TestIgnoringTagsBeanstalk(t *testing.T) {
	var ignoredTags []*elasticbeanstalk.Tag
	ignoredTags = append(ignoredTags, &elasticbeanstalk.Tag{
		Key:   aws.String("aws:cloudformation:logical-id"),
		Value: aws.String("foo"),
	})
	ignoredTags = append(ignoredTags, &elasticbeanstalk.Tag{
		Key:   aws.String("aws:foo:bar"),
		Value: aws.String("baz"),
	})
	for _, tag := range ignoredTags {
		if !tagIgnoredBeanstalk(tag) {
			t.Fatalf("Tag %v with value %v not ignored, but should be!", *tag.Key, *tag.Value)
		}
	}
}
