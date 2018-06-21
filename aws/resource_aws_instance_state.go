package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsInstanceState() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsInstanceStateCreate,
		Read:   resourceAwsInstanceStateRead,
		Update: resourceAwsInstanceStateUpdate,
		Delete: resourceAwsInstanceStateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAwsInstanceStateCreate(d *schema.ResourceData, meta interface{}) error {
	if err := resourceAwsInstanceStateUpdate(d, meta); err != nil {
		return err
	}

	d.SetId(d.Get("instance_id").(string))
	return nil
}

func resourceAwsInstanceStateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	id := d.Get("instance_id").(string)
	state, err := awsInStateReadInstance(conn, id)

	if err != nil {
		return err
	}

	d.Set("state", state)
	return nil
}

func resourceAwsInstanceStateUpdate(d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("state") {
		return nil
	}
	conn := meta.(*AWSClient).ec2conn
	state := d.Get("state").(string)
	id := d.Get("instance_id").(string)
	var err error

	switch state {
	case "stopped":
		timeout := d.Timeout(schema.TimeoutDelete)
		err = awsInstanceStateChange(
			id, "stopped", conn, timeout, func() error {
				_, err := conn.StopInstances(&ec2.StopInstancesInput{
					InstanceIds: aws.StringSlice([]string{id}),
				})
				return err
			})
	case "running":
	case "started":
		timeout := d.Timeout(schema.TimeoutCreate)
		err = awsInstanceStateChange(
			id, "running", conn, timeout, func() error {
				_, err := conn.StartInstances(&ec2.StartInstancesInput{
					InstanceIds: aws.StringSlice([]string{id}),
				})
				return err
			})
	default:
		return fmt.Errorf("State name %s is invalid.", state)
	}

	if err != nil {
		return err
	}

	return resourceAwsInstanceStateRead(d, meta)
}

func resourceAwsInstanceStateDelete(d *schema.ResourceData, meta interface{}) error {
	d.SetId("")
	return nil
}

func awsInstanceStateChange(id, state string, conn *ec2.EC2, timeout time.Duration, changeFunc func() error) error {
	if err := changeFunc(); err != nil {
		return err
	}

	log.Printf("[DEBUG] Waiting for instance (%s) to become %s", id, state)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "running", "shutting-down", "stopped", "stopping"},
		Target:     []string{state},
		Refresh:    InstanceStateRefreshFunc(conn, id, []string{}),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to stopped: %s", id, err)
	}
	return nil
}

func awsInStateReadInstance(conn *ec2.EC2, id string) (string, error) {
	descOutput, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	})

	if err != nil {
		return "", err
	}

	instance := descOutput.Reservations[0].Instances[0]
	log.Printf("[DEBUG] current state is %s", *instance.State.Name)

	return *instance.State.Name, nil
}
