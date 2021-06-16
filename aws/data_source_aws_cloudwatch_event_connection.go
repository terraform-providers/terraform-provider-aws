package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsCloudwatchEventConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudwatchEventConnectionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCloudwatchEventConnectionRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("name").(string))

	conn := meta.(*AWSClient).cloudwatcheventsconn

	input := &events.DescribeConnectionInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading CloudWatchEvent connection (%s)", d.Id())
	output, err := conn.DescribeConnection(input)
	if err != nil {
		return fmt.Errorf("error getting CloudWatchEvent connection (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting CloudWatchEvent connection (%s): empty response", d.Id())
	}

	log.Printf("[DEBUG] Found CloudWatchEvent connection: %#v", *output)
	d.Set("arn", output.ConnectionArn)
	d.Set("secret_arn", output.SecretArn)
	d.Set("name", output.Name)
	d.Set("authorization_type", output.AuthorizationType)
	return nil
}
