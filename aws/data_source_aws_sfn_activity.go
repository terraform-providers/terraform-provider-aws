package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsSfnActivity() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSfnActivityRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ExactlyOneOf: []string{
					"arn",
					"name",
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ExactlyOneOf: []string{
					"arn",
					"name",
				},
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsSfnActivityRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient)
	conn := client.sfnconn
	log.Print("[DEBUG] Reading Step Function Activity")

	if nm, ok := d.GetOk("name"); ok {
		name := nm.(string)
		var acts []*sfn.ActivityListItem

		err := conn.ListActivitiesPages(&sfn.ListActivitiesInput{}, func(page *sfn.ListActivitiesOutput, b bool) bool {
			for _, a := range page.Activities {
				if name == aws.StringValue(a.Name) {
					acts = append(acts, a)
				}
			}
			return true
		})

		if err != nil {
			return fmt.Errorf("Error listing activities: %s", err)
		}

		if len(acts) == 0 {
			return fmt.Errorf("No activity found with name %s in this region", name)
		}

		if len(acts) > 1 {
			return fmt.Errorf("Found more than 1 activity with name %s in this region", name)
		}

		act := acts[0]

		d.SetId(*act.ActivityArn)
		d.Set("name", act.Name)
		d.Set("arn", act.ActivityArn)
		if err := d.Set("creation_date", act.CreationDate.Format(time.RFC3339)); err != nil {
			log.Printf("[DEBUG] Error setting creation_date: %s", err)
		}
	}

	if rnm, ok := d.GetOk("arn"); ok {
		arn := rnm.(string)
		params := &sfn.DescribeActivityInput{
			ActivityArn: aws.String(arn),
		}

		act, err := conn.DescribeActivity(params)
		if err != nil {
			return fmt.Errorf("Error describing activities: %s", err)
		}

		if act == nil {
			return fmt.Errorf("No activity found with arn %s in this region", arn)
		}

		d.SetId(*act.ActivityArn)
		d.Set("name", act.Name)
		d.Set("arn", act.ActivityArn)
		if err := d.Set("creation_date", act.CreationDate.Format(time.RFC3339)); err != nil {
			log.Printf("[DEBUG] Error setting creation_date: %s", err)
		}
	}

	return nil
}
