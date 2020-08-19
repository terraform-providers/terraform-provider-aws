package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

func diffTagsQuickSight(oldTags, newTags []*quicksight.Tag) ([]*quicksight.Tag, []*quicksight.Tag) {
	create := make(map[string]interface{})
	for _, t := range newTags {
		create[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	var remove []*quicksight.Tag
	for _, t := range oldTags {
		old, ok := create[aws.StringValue(t.Key)]
		if !ok || old != aws.StringValue(t.Value) {
			remove = append(remove, t)
		} else if ok {
			delete(create, aws.StringValue(t.Key))
		}
	}

	return tagsFromMapQuickSight(create), remove
}

func tagsFromMapQuickSight(m map[string]interface{}) []*quicksight.Tag {
	result := make([]*quicksight.Tag, 0, len(m))
	for k, v := range m {
		t := &quicksight.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		result = append(result, t)
	}

	return result
}

func tagKeysQuickSight(ts []*quicksight.Tag) []*string {
	result := make([]*string, 0, len(ts))
	for _, t := range ts {
		result = append(result, t.Key)
	}
	return result
}
