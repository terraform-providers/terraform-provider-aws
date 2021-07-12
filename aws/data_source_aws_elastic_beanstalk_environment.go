package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsElasticBeanstalkEnvironment() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsElasticBeanstalkEnvironmentRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"application": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

// dataSourceAwsElasticBeanstalkEnvironmentRead performs the EB environment lookup.
func dataSourceAwsElasticBeanstalkEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elasticbeanstalkconn

	params := &elasticbeanstalk.DescribeEnvironmentsInput{
		ApplicationName: aws.String(d.Get("application").(string)),
	}

	log.Printf("[DEBUG] Reading Elastic Beanstalk Environments: %s", params)

	describeResp, err := conn.DescribeEnvironments(params)
	if err != nil {
		return fmt.Errorf("Error retrieving environments: %s", err)
	}

	// TODO: just return first for now
	env := describeResp.Environments[0]

	d.SetId(aws.StringValue(describeResp.Environments[0].EnvironmentId))

	arn := aws.StringValue(env.EnvironmentArn)
	d.Set("arn", arn)

	if err := d.Set("name", env.EnvironmentName); err != nil {
		return err
	}

	if err := d.Set("application", env.ApplicationName); err != nil {
		return err
	}

	return nil
}
