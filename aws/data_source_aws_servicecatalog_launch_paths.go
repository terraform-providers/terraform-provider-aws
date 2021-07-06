package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
)

func dataSourceAwsServiceCatalogLaunchPaths() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceCatalogLaunchPathsRead,

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"summaries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"constraint_summaries": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"description": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"path_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": tagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceAwsServiceCatalogLaunchPathsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	summaries, err := waiter.LaunchPathsReady(conn, d.Get("accept_language").(string), d.Get("product_id").(string))

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Launch Paths: %w", err)
	}

	if err := d.Set("summaries", flattenServiceCatalogLaunchPathSummaries(summaries, ignoreTagsConfig)); err != nil {
		return fmt.Errorf("error setting summaries: %w", err)
	}

	d.SetId(d.Get("product_id").(string))

	return nil
}

func flattenServiceCatalogLaunchPathSummary(apiObject *servicecatalog.LaunchPathSummary, ignoreTagsConfig *keyvaluetags.IgnoreConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if len(apiObject.ConstraintSummaries) > 0 {
		tfMap["constraint_summaries"] = flattenServiceCatalogConstraintSummaries(apiObject.ConstraintSummaries)
	}

	if apiObject.Id != nil {
		tfMap["path_id"] = aws.StringValue(apiObject.Id)
	}

	if apiObject.Name != nil {
		tfMap["name"] = aws.StringValue(apiObject.Name)
	}

	tags := keyvaluetags.ServicecatalogKeyValueTags(apiObject.Tags)

	tfMap["tags"] = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()

	return tfMap
}

func flattenServiceCatalogLaunchPathSummaries(apiObjects []*servicecatalog.LaunchPathSummary, ignoreTagsConfig *keyvaluetags.IgnoreConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenServiceCatalogLaunchPathSummary(apiObject, ignoreTagsConfig))
	}

	return tfList
}

func flattenServiceCatalogConstraintSummary(apiObject *servicecatalog.ConstraintSummary) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Description != nil {
		tfMap["description"] = aws.StringValue(apiObject.Description)
	}

	if apiObject.Type != nil {
		tfMap["type"] = aws.StringValue(apiObject.Type)
	}

	return tfMap
}

func flattenServiceCatalogConstraintSummaries(apiObjects []*servicecatalog.ConstraintSummary) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenServiceCatalogConstraintSummary(apiObject))
	}

	return tfList
}
