package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsRamResourceShare() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRamResourceShareRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ram.ResourceOwnerOtherAccounts,
					ram.ResourceOwnerSelf,
				}, false),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"owning_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsRamResourceShareRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	owner := d.Get("resource_owner").(string)

	filters, filtersOk := d.GetOk("filter")

	params := &ram.GetResourceSharesInput{
		Name:          aws.String(name),
		ResourceOwner: aws.String(owner),
	}

	if filtersOk {
		params.TagFilters = buildRAMTagFilters(filters.(*schema.Set))
	}

	for {
		resp, err := conn.GetResourceShares(params)

		if err != nil {
			return fmt.Errorf("Error retrieving resource share: empty response for: %s", params)
		}

		if len(resp.ResourceShares) > 1 {
			return fmt.Errorf("Multiple resource shares found for: %s", name)
		}

		if resp == nil || len(resp.ResourceShares) == 0 {
			return fmt.Errorf("No matching resource found: %s", err)
		}

		for _, r := range resp.ResourceShares {
			if aws.StringValue(r.Name) == name {
				d.SetId(aws.StringValue(r.ResourceShareArn))
				d.Set("arn", aws.StringValue(r.ResourceShareArn))
				d.Set("owning_account_id", aws.StringValue(r.OwningAccountId))
				d.Set("status", aws.StringValue(r.Status))

				if err := d.Set("tags", keyvaluetags.RamKeyValueTags(r.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
					return fmt.Errorf("error setting tags: %s", err)
				}

				break
			}
		}

		if resp.NextToken == nil {
			break
		}

		params.NextToken = resp.NextToken
	}

	return nil
}

func buildRAMTagFilters(set *schema.Set) []*ram.TagFilter {
	var filters []*ram.TagFilter

	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}
		filters = append(filters, &ram.TagFilter{
			TagKey:    aws.String(m["name"].(string)),
			TagValues: filterValues,
		})
	}

	return filters
}
