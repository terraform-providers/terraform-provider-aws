package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsSfnStateMachine() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSfnStateMachineRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsSfnStateMachineRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	params := &sfn.ListStateMachinesInput{}
	log.Printf("[DEBUG] Reading Step Function State Machine: %s", d.Id())

	target := d.Get("name")
	var arns []string

	err := conn.ListStateMachinesPages(params, func(page *sfn.ListStateMachinesOutput, lastPage bool) bool {
		for _, sm := range page.StateMachines {
			if *sm.Name == target {
				arns = append(arns, *sm.StateMachineArn)
			}
		}
		return true
	})

	if err != nil {
		return fmt.Errorf("Error listing state machines: %s", err)
	}

	if len(arns) == 0 {
		return fmt.Errorf("No state machine with name %q found in this region.", target)
	}
	if len(arns) > 1 {
		return fmt.Errorf("Multiple state machines with name %q found in this region.", target)
	}

	sm, err := conn.DescribeStateMachine(&sfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(arns[0]),
	})
	if err != nil {
		return fmt.Errorf("error describing SFN State Machine (%s): %w", arns[0], err)
	}

	d.Set("definition", sm.Definition)
	d.Set("name", sm.Name)
	d.Set("arn", sm.StateMachineArn)
	d.Set("role_arn", sm.RoleArn)
	d.Set("status", sm.Status)
	if err := d.Set("creation_date", sm.CreationDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting creation_date: %s", err)
	}

	d.SetId(arns[0])

	return nil
}
