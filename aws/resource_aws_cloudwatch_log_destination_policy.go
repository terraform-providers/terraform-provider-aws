package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsCloudWatchLogDestinationPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchLogDestinationPolicyPut,
		Update: resourceAwsCloudWatchLogDestinationPolicyPut,
		Read:   resourceAwsCloudWatchLogDestinationPolicyRead,
		Delete: resourceAwsCloudWatchLogDestinationPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("destination_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"destination_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"access_policy": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsCloudWatchLogDestinationPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn

	destination_name := d.Get("destination_name").(string)
	access_policy := d.Get("access_policy").(string)

	params := &cloudwatchlogs.PutDestinationPolicyInput{
		DestinationName: aws.String(destination_name),
		AccessPolicy:    aws.String(access_policy),
	}

	_, err := conn.PutDestinationPolicy(params)

	if err != nil {
		return fmt.Errorf("Error creating DestinationPolicy with destination_name %s: %#v", destination_name, err)
	}

	d.SetId(destination_name)
	return resourceAwsCloudWatchLogDestinationPolicyRead(d, meta)
}

func resourceAwsCloudWatchLogDestinationPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchlogsconn
	destination_name := d.Get("destination_name").(string)
	destination, exists, err := lookupCloudWatchLogDestination(conn, destination_name, nil)
	if err != nil {
		return err
	}

	if !exists || destination.AccessPolicy == nil {
		log.Printf("[WARN] CloudWatch Log Destination Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("access_policy", destination.AccessPolicy)

	return nil
}

func resourceAwsCloudWatchLogDestinationPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
