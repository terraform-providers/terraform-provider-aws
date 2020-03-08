package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsQuickSightGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsQuickSightGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aws_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsQuickSightGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).quicksightconn
	groupName := d.Get("group_name").(string)
	req := &quicksight.DescribeGroupInput{
		GroupName:    aws.String(groupName),
		AwsAccountId: aws.String(d.Get("aws_account_id").(string)),
		Namespace:    aws.String(d.Get("namespace").(string)),
	}

	log.Printf("[DEBUG] Reading quicksight group: %s", req)
	resp, err := conn.DescribeGroup(req)
	if err != nil {
		return fmt.Errorf("error getting group: %s", err)
	}

	group := resp.Group
	d.SetId(aws.StringValue(group.PrincipalId))
	d.Set("arn", group.Arn)
	d.Set("description", group.Description)
	d.Set("group_id", group.PrincipalId)

	return nil
}
