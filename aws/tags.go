package aws

import (
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// tagsSchema returns the schema to use for tags.
//
func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
	}
}

func tagsSchemaComputed() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Computed: true,
	}
}

func tagsSchemaForceNew() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		ForceNew: true,
	}
}

// tagsFromMap returns the tags for the given map of data.
func tagsFromMap(m map[string]interface{}) []*ec2.Tag {
	result := make([]*ec2.Tag, 0, len(m))
	for k, v := range m {
		t := &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v.(string)),
		}
		if !tagIgnored(t) {
			result = append(result, t)
		}
	}

	return result
}

// tagsToMap turns the list of tags into a map.
func tagsToMap(ts []*ec2.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range ts {
		if !tagIgnored(t) {
			result[*t.Key] = *t.Value
		}
	}

	return result
}

// tagIgnored compares a tag against a list of strings and checks if it should
// be ignored or not
func tagIgnored(t *ec2.Tag) bool {
	filter := []string{"^aws:"}
	for _, v := range filter {
		log.Printf("[DEBUG] Matching %v with %v\n", v, *t.Key)
		r, _ := regexp.MatchString(v, *t.Key)
		if r {
			log.Printf("[DEBUG] Found AWS specific tag %s (val: %s), ignoring.\n", *t.Key, *t.Value)
			return true
		}
	}
	return false
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
