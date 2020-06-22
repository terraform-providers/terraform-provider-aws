package aws

import (
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/namevaluesfilters"
)

func dataSourceAwsAutoscalingGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsAutoscalingGroupsRead,

		Schema: map[string]*schema.Schema{
			"names": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"arns": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": namevaluesfilters.Schema(),
		},
	}
}

func dataSourceAwsAutoscalingGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	log.Printf("[DEBUG] Reading Autoscaling Groups.")

	var rawName []string
	var rawArn []string
	var err error

	if v := d.Get("filter").(*schema.Set); v.Len() > 0 {
		input := &autoscaling.DescribeTagsInput{
			Filters: namevaluesfilters.New(v).AutoscalingFilters(),
		}

		err = conn.DescribeTagsPages(input, func(resp *autoscaling.DescribeTagsOutput, lastPage bool) bool {
			for _, v := range resp.Tags {
				rawName = append(rawName, aws.StringValue(v.ResourceId))
			}
			return !lastPage
		})

		maxAutoScalingGroupNames := 1600
		for i := 0; i < len(rawName); i += maxAutoScalingGroupNames {
			end := i + maxAutoScalingGroupNames

			if end > len(rawName) {
				end = len(rawName)
			}

			nameInput := &autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: aws.StringSlice(rawName[i:end]),
				MaxRecords:            aws.Int64(100),
			}

			err = conn.DescribeAutoScalingGroupsPages(nameInput, func(resp *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
				for _, group := range resp.AutoScalingGroups {
					rawArn = append(rawArn, aws.StringValue(group.AutoScalingGroupARN))
				}
				return !lastPage
			})
		}
	} else {
		err = conn.DescribeAutoScalingGroupsPages(&autoscaling.DescribeAutoScalingGroupsInput{}, func(resp *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			for _, group := range resp.AutoScalingGroups {
				rawName = append(rawName, aws.StringValue(group.AutoScalingGroupName))
				rawArn = append(rawArn, aws.StringValue(group.AutoScalingGroupARN))
			}
			return !lastPage
		})
	}
	if err != nil {
		return fmt.Errorf("Error fetching Autoscaling Groups: %w", err)
	}

	d.SetId(meta.(*AWSClient).region)

	sort.Strings(rawName)
	sort.Strings(rawArn)

	if err := d.Set("names", rawName); err != nil {
		return fmt.Errorf("[WARN] Error setting Autoscaling Group Names: %w", err)
	}

	if err := d.Set("arns", rawArn); err != nil {
		return fmt.Errorf("[WARN] Error setting Autoscaling Group ARNs: %w", err)
	}

	return nil

}
