package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsOrganizationsAccounts() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsAccountsRead,

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"children": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsOrganizationsAccountsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	parent_id := d.Get("parent_id").(string)

	params := &organizations.ListAccountsForParentInput{
		ParentId: aws.String(parent_id),
	}

	var children []*organizations.Account

	err := conn.ListAccountsForParentPages(params,
		func(page *organizations.ListAccountsForParentOutput, lastPage bool) bool {
			children = append(children, page.Accounts...)

			return !lastPage
		})

	if err != nil {
		return fmt.Errorf("error listing Organizations Organization Accounts for parent (%s): %w", parent_id, err)
	}

	d.SetId(parent_id)

	if err := d.Set("children", flattenOrganizationsAccounts(children)); err != nil {
		return fmt.Errorf("error setting children: %w", err)
	}

	return nil
}
