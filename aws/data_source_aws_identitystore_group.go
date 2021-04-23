package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsIdentityStoreGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIdentityStoreGroupRead,

		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"filter": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"attribute_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},
		},
	}
}

func dataSourceAwsIdentityStoreGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).identitystoreconn

	input := &identitystore.ListGroupsInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		Filters:         expandIdentityStoreFilters(d.Get("filter").(*schema.Set).List()),
	}

	var results []*identitystore.Group

	err := conn.ListGroupsPages(input, func(page *identitystore.ListGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, group := range page.Groups {
			if group == nil {
				continue
			}

			if v, ok := d.GetOk("group_id"); ok && v.(string) != aws.StringValue(group.GroupId) {
				continue
			}

			results = append(results, group)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Identity Store Groups: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no Identity Store Group found matching criteria; try different search")
	}

	if len(results) > 1 {
		return fmt.Errorf("multiple Identity Store Groups found matching criteria; try different search")
	}

	group := results[0]

	d.SetId(aws.StringValue(group.GroupId))
	d.Set("display_name", group.DisplayName)
	d.Set("group_id", group.GroupId)

	return nil
}

func expandIdentityStoreFilters(l []interface{}) []*identitystore.Filter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	filters := make([]*identitystore.Filter, 0, len(l))
	for _, v := range l {
		tfMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		filter := &identitystore.Filter{}

		if v, ok := tfMap["attribute_path"].(string); ok && v != "" {
			filter.AttributePath = aws.String(v)
		}

		if v, ok := tfMap["attribute_value"].(string); ok && v != "" {
			filter.AttributeValue = aws.String(v)
		}

		filters = append(filters, filter)
	}

	return filters
}
