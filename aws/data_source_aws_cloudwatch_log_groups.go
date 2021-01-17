package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCloudwatchLogGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudwatchLogGroupsRead,
		Schema: map[string]*schema.Schema{
			"log_group_name_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arns": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"log_group_names": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsCloudwatchLogGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	logGroupNamePrefix := d.Get("log_group_name_prefix").(string)
	input := &cloudwatchlogs.DescribeLogGroupsInput{LogGroupNamePrefix: &logGroupNamePrefix}
	var logGroupNames = []string{}
	var arns = []string{}

	err := conn.DescribeLogGroupsPages(input,
		func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
			for _, group := range page.LogGroups {
				logGroupNames = append(logGroupNames, aws.StringValue(group.LogGroupName))
				arns = append(arns, aws.StringValue(group.Arn))
			}
			return !lastPage
		})
	if err != nil {
		return err
	}

	err = d.Set("log_group_names", logGroupNames)
	if err != nil {
		return fmt.Errorf("Error setting Log Group Names: %s", err)
	}

	err = d.Set("arns", arns)
	if err != nil {
		return fmt.Errorf("Error setting Log Group Arns: %s", err)
	}
	d.SetId(meta.(*AWSClient).region)
	return nil
}
