package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsOutpostsOutpostInstanceTypes() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsOutpostsOutpostInstanceTypesRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsOutpostsOutpostInstanceTypesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).outpostsconn

	input := &outposts.GetOutpostInstanceTypesInput{
		OutpostId: aws.String(d.Get("arn").(string)), // Accepts both ARN and ID; prefer ARN which is more common
	}

	var outpostID string
	var instanceTypes []string

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
			instanceTypes = append(instanceTypes, aws.StringValue(outputInstanceType.InstanceType))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := d.Set("instance_types", instanceTypes); err != nil {
		return fmt.Errorf("error setting instance_types: %w", err)
	}

	d.SetId(outpostID)

	return nil
}
