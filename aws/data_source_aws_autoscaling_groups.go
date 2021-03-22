package aws

import (
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsAutoscalingGroupsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingconn

	log.Printf("[DEBUG] Reading Autoscaling Groups.")

	var rawName []string
	var rawArn []string
	var err error

	tf := d.Get("filter").(*schema.Set)
	if tf.Len() > 0 {
		input := &autoscaling.DescribeTagsInput{
			Filters: expandAsgTagFilters(tf.List()),
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

func expandAsgTagFilters(in []interface{}) []*autoscaling.Filter {
	out := make([]*autoscaling.Filter, len(in))
	for i, filter := range in {
		m := filter.(map[string]interface{})
		values := expandStringSet(m["values"].(*schema.Set))

		out[i] = &autoscaling.Filter{
			Name:   aws.String(m["name"].(string)),
			Values: values,
		}
	}
	return out
}
