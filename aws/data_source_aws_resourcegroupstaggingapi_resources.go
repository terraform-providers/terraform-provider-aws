package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsResourceGroupsTaggingAPIResources() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsResourceGroupsTaggingAPIResourcesRead,

		Schema: map[string]*schema.Schema{
			"exclude_compliant_resources": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"include_compliance_details": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"resource_arn_list": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"tag_filter"},
			},
			"resource_type_filters": {
				Type:          schema.TypeSet,
				Optional:      true,
				MaxItems:      100,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"resource_arn_list"},
			},
			"tag_filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 20,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"resource_tag_mapping_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"compliance_details": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compliance_status": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"keys_with_noncompliant_values": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"non_compliant_keys": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"tags": tagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceAwsResourceGroupsTaggingAPIResourcesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).resourcegroupstaggingapiconn

	input := &resourcegroupstaggingapi.GetResourcesInput{}

	if v, ok := d.GetOk("include_compliance_details"); ok {
		input.IncludeComplianceDetails = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("exclude_compliant_resources"); ok {
		input.ExcludeCompliantResources = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("resource_arn_list"); ok && v.(*schema.Set).Len() > 0 {
		input.ResourceARNList = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_filter"); ok {
		input.TagFilters = expandAwsResourceGroupsTaggingAPITagFilters(v.([]interface{}))
	}

	if v, ok := d.GetOk("resource_type_filters"); ok && v.(*schema.Set).Len() > 0 {
		input.ResourceTypeFilters = expandStringSet(v.(*schema.Set))
	}

	var taggings []*resourcegroupstaggingapi.ResourceTagMapping

	err := conn.GetResourcesPages(input, func(page *resourcegroupstaggingapi.GetResourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		taggings = append(taggings, page.ResourceTagMappingList...)
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error getting Resource Groups Tags API Resources: %w", err)
	}

	d.SetId(meta.(*AWSClient).partition)

	if err := d.Set("resource_tag_mapping_list", flattenAwsResourceGroupsTaggingAPIResourcesTagMappingList(taggings)); err != nil {
		return fmt.Errorf("error setting resource tag mapping list: %w", err)
	}

	return nil
}

func expandAwsResourceGroupsTaggingAPITagFilters(filters []interface{}) []*resourcegroupstaggingapi.TagFilter {
	result := make([]*resourcegroupstaggingapi.TagFilter, len(filters))

	for i, filter := range filters {
		m := filter.(map[string]interface{})

		result[i] = &resourcegroupstaggingapi.TagFilter{
			Key: aws.String(m["key"].(string)),
		}

		if v, ok := m["values"]; ok && v.(*schema.Set).Len() > 0 {
			result[i].Values = expandStringSet(v.(*schema.Set))
		}
	}

	return result
}

func flattenAwsResourceGroupsTaggingAPIResourcesTagMappingList(list []*resourcegroupstaggingapi.ResourceTagMapping) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		l := map[string]interface{}{
			"resource_arn": aws.StringValue(i.ResourceARN),
			"tags":         keyvaluetags.ResourcegroupstaggingapiKeyValueTags(i.Tags).Map(),
		}

		if i.ComplianceDetails != nil {
			l["compliance_details"] = flattenAwsResourceGroupsTaggingAPIComplianceDetails(i.ComplianceDetails)
		}

		result = append(result, l)
	}

	return result
}

func flattenAwsResourceGroupsTaggingAPIComplianceDetails(details *resourcegroupstaggingapi.ComplianceDetails) []map[string]interface{} {
	if details == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"compliance_status":             aws.BoolValue(details.ComplianceStatus),
		"keys_with_noncompliant_values": flattenStringSet(details.KeysWithNoncompliantValues),
		"non_compliant_keys":            flattenStringSet(details.NoncompliantKeys),
	}

	return []map[string]interface{}{m}
}
