package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsOrganizationsOrganizationalUnits() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOrganizationsOrganizationalUnitsRead,

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

func dataSourceAwsOrganizationsOrganizationalUnitsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).organizationsconn

	parent_id := d.Get("parent_id").(string)

	params := &organizations.ListOrganizationalUnitsForParentInput{
		ParentId: aws.String(parent_id),
	}

	var children []*organizations.OrganizationalUnit

	err := conn.ListOrganizationalUnitsForParentPages(params,
		func(page *organizations.ListOrganizationalUnitsForParentOutput, lastPage bool) bool {
			children = append(children, page.OrganizationalUnits...)

			return !lastPage
		})

	if err != nil {
		return fmt.Errorf("error listing Organizations Organization Units for parent (%s): %w", parent_id, err)
	}

	d.SetId(parent_id)

	if err := d.Set("children", flattenOrganizationsOrganizationalUnits(children)); err != nil {
		return fmt.Errorf("error setting children: %w", err)
	}

	return nil
}

func flattenOrganizationsOrganizationalUnits(ous []*organizations.OrganizationalUnit) []map[string]interface{} {
	if len(ous) == 0 {
		return nil
	}
	var result []map[string]interface{}
	for _, ou := range ous {
		result = append(result, map[string]interface{}{
			"arn":  aws.StringValue(ou.Arn),
			"id":   aws.StringValue(ou.Id),
			"name": aws.StringValue(ou.Name),
		})
	}
	return result
}
