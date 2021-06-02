package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// tagsSchema returns the schema to use for tags.
//
func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
}

func tagsSchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
}

func tagsSchemaForceNew() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
	}
}

func tagsSchemaConflictsWith(conflictsWith []string) *schema.Schema {
	return &schema.Schema{
		ConflictsWith: conflictsWith,
		Type:          schema.TypeMap,
		Optional:      true,
		Elem:          &schema.Schema{Type: schema.TypeString},
	}
}

// ec2TagsFromTagDescriptions returns the tags from the given tag descriptions.
// No attempt is made to remove duplicates.
func ec2TagsFromTagDescriptions(tds []*ec2.TagDescription) []*ec2.Tag {
	if len(tds) == 0 {
		return nil
	}

	tags := []*ec2.Tag{}
	for _, td := range tds {
		tags = append(tags, &ec2.Tag{
			Key:   td.Key,
			Value: td.Value,
		})
	}

	return tags
}

// ec2TagSpecificationsFromMap returns the tag specifications for the given tag key/value map and resource type.
func ec2TagSpecificationsFromMap(m map[string]interface{}, t string) []*ec2.TagSpecification {
	if len(m) == 0 {
		return nil
	}

	return []*ec2.TagSpecification{
		{
			ResourceType: aws.String(t),
			Tags:         keyvaluetags.New(m).IgnoreAws().Ec2Tags(),
		},
	}
}

// ec2TagSpecificationsFromKeyValueTags returns the tag specifications for the given KeyValueTags object and resource type.
func ec2TagSpecificationsFromKeyValueTags(tags keyvaluetags.KeyValueTags, t string) []*ec2.TagSpecification {
	if len(tags) == 0 {
		return nil
	}

	return []*ec2.TagSpecification{
		{
			ResourceType: aws.String(t),
			Tags:         tags.IgnoreAws().Ec2Tags(),
		},
	}
}

// SetTagsDiff sets the new plan difference with the result of
// merging resource tags on to those defined at the provider-level;
// returns an error if unsuccessful or if the resource tags are identical
// to those configured at the provider-level to avoid non-empty plans
// after resource READ operations as resource and provider-level tags
// will be indistinguishable when returned from an AWS API.
func SetTagsDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resourceTags := keyvaluetags.New(diff.Get("tags").(map[string]interface{}))

	if defaultTagsConfig.TagsEqual(resourceTags) {
		return fmt.Errorf(`"tags" are identical to those in the "default_tags" configuration block of the provider: please de-duplicate and try again`)
	}

	allTags := defaultTagsConfig.MergeTags(resourceTags).IgnoreConfig(ignoreTagsConfig)

	// To ensure "tags_all" is correctly computed, we explicitly set the attribute diff
	// when the merger of resource-level tags onto provider-level tags results in n > 0 tags,
	// otherwise we mark the attribute as "Computed" only when their is a known diff (excluding an empty map)
	// or a change for "tags_all".
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18366
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19005
	if len(allTags) > 0 {
		if err := diff.SetNew("tags_all", allTags.Map()); err != nil {
			return fmt.Errorf("error setting new tags_all diff: %w", err)
		}
	} else if len(diff.Get("tags_all").(map[string]interface{})) > 0 {
		if err := diff.SetNewComputed("tags_all"); err != nil {
			return fmt.Errorf("error setting tags_all to computed: %w", err)
		}
	} else if diff.HasChange("tags_all") {
		if err := diff.SetNewComputed("tags_all"); err != nil {
			return fmt.Errorf("error setting tags_all to computed: %w", err)
		}
	}

	return nil
}
