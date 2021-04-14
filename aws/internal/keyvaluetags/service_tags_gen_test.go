package keyvaluetags

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

// map[string]*string handling

func TestAmplifyKeyValueTags(t *testing.T) {
	testCases := []struct {
		name string
		tags map[string]*string
		want map[string]string
	}{
		{
			name: "empty",
			tags: map[string]*string{},
			want: map[string]string{},
		},
		{
			name: "non_empty",
			tags: map[string]*string{
				"key1": aws.String("value1"),
				"key2": aws.String("value2"),
				"key3": aws.String("value3"),
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := AmplifyKeyValueTags(testCase.tags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsAmplifyTags(t *testing.T) {
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
			name: "non_empty",
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
			got := testCase.tags.AmplifyTags()

			testKeyValueTagsVerifyMap(t, aws.StringValueMap(got), testCase.want)
		})
	}
}

// []*SERVICE.Tag handling

func TestAthenaKeyValueTags(t *testing.T) {
	testCases := []struct {
		name string
		tags []*athena.Tag
		want map[string]string
	}{
		{
			name: "empty",
			tags: []*athena.Tag{},
			want: map[string]string{},
		},
		{
			name: "non_empty",
			tags: []*athena.Tag{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := AthenaKeyValueTags(testCase.tags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsAthenaTags(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want []*athena.Tag
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: []*athena.Tag{},
		},
		{
			name: "non_empty",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []*athena.Tag{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.AthenaTags()

			gotMap := make(map[string]string, len(got))
			for _, tag := range got {
				gotMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			wantMap := make(map[string]string, len(testCase.want))
			for _, tag := range testCase.want {
				wantMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			testKeyValueTagsVerifyMap(t, gotMap, wantMap)
		})
	}
}

// []*SERVICE.Tag (with TagKey and TagValue fields) handling

func TestKmsKeyValueTags(t *testing.T) {
	testCases := []struct {
		name string
		tags []*kms.Tag
		want map[string]string
	}{
		{
			name: "empty",
			tags: []*kms.Tag{},
			want: map[string]string{},
		},
		{
			name: "non_empty",
			tags: []*kms.Tag{
				{
					TagKey:   aws.String("key1"),
					TagValue: aws.String("value1"),
				},
				{
					TagKey:   aws.String("key2"),
					TagValue: aws.String("value2"),
				},
				{
					TagKey:   aws.String("key3"),
					TagValue: aws.String("value3"),
				},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := KmsKeyValueTags(testCase.tags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsKmsTags(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want []*kms.Tag
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: []*kms.Tag{},
		},
		{
			name: "non_empty",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []*kms.Tag{
				{
					TagKey:   aws.String("key1"),
					TagValue: aws.String("value1"),
				},
				{
					TagKey:   aws.String("key2"),
					TagValue: aws.String("value2"),
				},
				{
					TagKey:   aws.String("key3"),
					TagValue: aws.String("value3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.KmsTags()

			gotMap := make(map[string]string, len(got))
			for _, tag := range got {
				gotMap[aws.StringValue(tag.TagKey)] = aws.StringValue(tag.TagValue)
			}

			wantMap := make(map[string]string, len(testCase.want))
			for _, tag := range testCase.want {
				wantMap[aws.StringValue(tag.TagKey)] = aws.StringValue(tag.TagValue)
			}

			testKeyValueTagsVerifyMap(t, gotMap, wantMap)
		})
	}
}

// []*SERVICE.TagListEntry handling

func TestDatasyncKeyValueTags(t *testing.T) {
	testCases := []struct {
		name string
		tags []*datasync.TagListEntry
		want map[string]string
	}{
		{
			name: "empty",
			tags: []*datasync.TagListEntry{},
			want: map[string]string{},
		},
		{
			name: "non_empty",
			tags: []*datasync.TagListEntry{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := DatasyncKeyValueTags(testCase.tags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsDatasyncTags(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want []*datasync.TagListEntry
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: []*datasync.TagListEntry{},
		},
		{
			name: "non_empty",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []*datasync.TagListEntry{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.DatasyncTags()

			gotMap := make(map[string]string, len(got))
			for _, tag := range got {
				gotMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			wantMap := make(map[string]string, len(testCase.want))
			for _, tag := range testCase.want {
				wantMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			testKeyValueTagsVerifyMap(t, gotMap, wantMap)
		})
	}
}

// []*SERVICE.TagRef handling

func TestAppmeshKeyValueTags(t *testing.T) {
	testCases := []struct {
		name string
		tags []*appmesh.TagRef
		want map[string]string
	}{
		{
			name: "empty",
			tags: []*appmesh.TagRef{},
			want: map[string]string{},
		},
		{
			name: "non_empty",
			tags: []*appmesh.TagRef{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := AppmeshKeyValueTags(testCase.tags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsAppmeshTags(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want []*appmesh.TagRef
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: []*appmesh.TagRef{},
		},
		{
			name: "non_empty",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []*appmesh.TagRef{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.AppmeshTags()

			gotMap := make(map[string]string, len(got))
			for _, tag := range got {
				gotMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			wantMap := make(map[string]string, len(testCase.want))
			for _, tag := range testCase.want {
				wantMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			testKeyValueTagsVerifyMap(t, gotMap, wantMap)
		})
	}
}

func TestQuicksightKeyValueTags(t *testing.T) {
	testCases := []struct {
		name string
		tags []*quicksight.Tag
		want map[string]string
	}{
		{
			name: "empty",
			tags: []*quicksight.Tag{},
			want: map[string]string{},
		},
		{
			name: "non_empty",
			tags: []*quicksight.Tag{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := QuicksightKeyValueTags(testCase.tags)

			testKeyValueTagsVerifyMap(t, got.Map(), testCase.want)
		})
	}
}

func TestKeyValueTagsQuicksightTags(t *testing.T) {
	testCases := []struct {
		name string
		tags KeyValueTags
		want []*quicksight.Tag
	}{
		{
			name: "empty",
			tags: New(map[string]string{}),
			want: []*quicksight.Tag{},
		},
		{
			name: "non_empty",
			tags: New(map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}),
			want: []*quicksight.Tag{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
				{
					Key:   aws.String("key3"),
					Value: aws.String("value3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.tags.QuicksightTags()

			gotMap := make(map[string]string, len(got))
			for _, tag := range got {
				gotMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			wantMap := make(map[string]string, len(testCase.want))
			for _, tag := range testCase.want {
				wantMap[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
			}

			testKeyValueTagsVerifyMap(t, gotMap, wantMap)
		})
	}
}
