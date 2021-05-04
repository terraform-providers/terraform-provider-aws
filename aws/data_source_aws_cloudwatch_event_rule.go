package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsCloudwatchEventRule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudwatchEventRuleRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsCloudwatchEventRuleRead(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input := cloudwatchevents.DescribeRuleInput{
		Name: aws.String(name),
	}

	resp, err := conn.DescribeRule(&input)
	if err != nil {
		return fmt.Errorf("error describing Cloudwatch Event Rule (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set("arn", aws.StringValue(resp.Arn))
	d.Set("description", aws.StringValue(resp.Description))
	d.Set("event_pattern", aws.StringValue(resp.EventPattern))
	d.Set("managed_by", aws.StringValue(resp.ManagedBy))
	d.Set("role_arn", aws.StringValue(resp.RoleArn))
	d.Set("schedule_expression", aws.StringValue(resp.ScheduleExpression))
	d.Set("state", aws.StringValue(resp.State))

	return nil
}
