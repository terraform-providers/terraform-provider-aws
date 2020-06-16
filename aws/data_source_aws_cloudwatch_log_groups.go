package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsCloudWatchLogGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudWatchLogGroupsRead,

		Schema: map[string]*schema.Schema{
			"prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"creation_times": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
				Set:      schema.HashInt,
			},
			"retention_in_days": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"metric_filter_counts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeInt},
			},
			"kms_key_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func getAllCloudWatchLogGroups(conn *cloudwatchlogs.CloudWatchLogs, input *cloudwatchlogs.DescribeLogGroupsInput) ([]*cloudwatchlogs.LogGroup, error) {
	var logGroups []*cloudwatchlogs.LogGroup
	var nextToken string

	for {
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}
		resp, err := conn.DescribeLogGroups(input)
		if err != nil {
			return nil, fmt.Errorf("Error describing Cloud Watch Log Groups: %s", err)
		}
		logGroups = append(logGroups, resp.LogGroups...)
		if resp.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(resp.NextToken)
	}
	return logGroups, nil
}

func dataSourceAwsCloudWatchLogGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	d.SetId(resource.UniqueId())

	logGroups, err := getAllCloudWatchLogGroups(conn, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(d.Get("prefix").(string)),
	})
	if err != nil {
		return err
	}

	var names, arns, creationTimes, retentionInDays, metricFilterCount, kmsKeyIDs []interface{}
	for _, v := range logGroups {
		names = append(names, aws.StringValue(v.LogGroupName))
		arns = append(arns, aws.StringValue(v.Arn))
		creationTimes = append(creationTimes, aws.Int64Value(v.CreationTime))
		retentionInDays = append(retentionInDays, int(aws.Int64Value(v.RetentionInDays)))
		metricFilterCount = append(metricFilterCount, int(aws.Int64Value(v.MetricFilterCount)))
		kmsKeyIDs = append(kmsKeyIDs, aws.StringValue(v.KmsKeyId))
	}

	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("Error setting Cloud Watch Log Group names: %s", err)
	}
	if err := d.Set("arns", arns); err != nil {
		return fmt.Errorf("Error setting Cloud Watch Log Group arns: %s", err)
	}
	if err := d.Set("creation_times", creationTimes); err != nil {
		return fmt.Errorf("Error setting Cloud Watch Log Group creation_times: %s", err)
	}
	if err := d.Set("retention_in_days", retentionInDays); err != nil {
		return fmt.Errorf("Error setting Cloud Watch Log Group retention_in_days: %s", err)
	}
	if err := d.Set("metric_filter_counts", metricFilterCount); err != nil {
		return fmt.Errorf("Error setting Cloud Watch Log Group metric_filter_counts: %s", err)
	}
	if err := d.Set("kms_key_ids", kmsKeyIDs); err != nil {
		return fmt.Errorf("Error setting Cloud Watch Log Group kms_key_ids: %s", err)
	}

	return nil
}
