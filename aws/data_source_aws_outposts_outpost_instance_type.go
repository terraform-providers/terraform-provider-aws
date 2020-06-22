package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsOutpostsOutpostInstanceType() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOutpostsOutpostInstanceTypeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"instance_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"preferred_instance_types"},
			},
			"preferred_instance_types": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"instance_type"},
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsOutpostsOutpostInstanceTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).outpostsconn

	input := &outposts.GetOutpostInstanceTypesInput{
		OutpostId: aws.String(d.Get("arn").(string)), // Accepts both ARN and ID; prefer ARN which is more common
	}

	var outpostID string
	var foundInstanceTypes []string

	for {
		output, err := conn.GetOutpostInstanceTypes(input)

		if err != nil {
			return fmt.Errorf("error getting Outpost Instance Types: %w", err)
		}

		if output == nil {
			break
		}

		outpostID = aws.StringValue(output.OutpostId)

		for _, outputInstanceType := range output.InstanceTypes {
			foundInstanceTypes = append(foundInstanceTypes, aws.StringValue(outputInstanceType.InstanceType))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if len(foundInstanceTypes) == 0 {
		return fmt.Errorf("no Outpost Instance Types found matching criteria; try different search")
	}

	var resultInstanceType string

	// Check requested instance type
	if v, ok := d.GetOk("instance_type"); ok {
		for _, foundInstanceType := range foundInstanceTypes {
			if foundInstanceType == v.(string) {
				resultInstanceType = v.(string)
				break
			}
		}
	}

	// Search preferred instance types in their given order and set result
	// instance type for first match found
	if l := d.Get("preferred_instance_types").([]interface{}); len(l) > 0 {
		for _, elem := range l {
			preferredInstanceType, ok := elem.(string)

			if !ok {
				continue
			}

			for _, foundInstanceType := range foundInstanceTypes {
				if foundInstanceType == preferredInstanceType {
					resultInstanceType = preferredInstanceType
					break
				}
			}

			if resultInstanceType != "" {
				break
			}
		}
	}

	if resultInstanceType == "" && len(foundInstanceTypes) > 1 {
		return fmt.Errorf("multiple Outpost Instance Types found matching criteria; try different search")
	}

	if resultInstanceType == "" && len(foundInstanceTypes) == 1 {
		resultInstanceType = foundInstanceTypes[0]
	}

	if resultInstanceType == "" {
		return fmt.Errorf("no Outpost Instance Types found matching criteria; try different search")
	}

	d.Set("instance_type", resultInstanceType)

	d.SetId(outpostID)

	return nil
}
