package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsEc2AvailabilityZoneGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2AvailabilityZoneGroupCreate,
		Read:   resourceAwsEc2AvailabilityZoneGroupRead,
		Update: resourceAwsEc2AvailabilityZoneGroupUpdate,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				d.Set("group_name", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"opt_in_status": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.AvailabilityZoneOptInStatusOptedIn,
					ec2.AvailabilityZoneOptInStatusNotOptedIn,
				}, false),
			},
		},
	}
}

func resourceAwsEc2AvailabilityZoneGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	configurationOptInStatus := d.Get("opt_in_status").(string)

	d.SetId(d.Get("group_name").(string))

	if err := resourceAwsEc2AvailabilityZoneGroupRead(d, meta); err != nil {
		return err
	}

	apiOptInStatus := d.Get("opt_in_status").(string)

	if apiOptInStatus != configurationOptInStatus {
		input := &ec2.ModifyAvailabilityZoneGroupInput{
			GroupName:   aws.String(d.Id()),
			OptInStatus: aws.String(configurationOptInStatus),
		}

		if _, err := conn.ModifyAvailabilityZoneGroup(input); err != nil {
			return fmt.Errorf("error modifying EC2 Availability Zone Group (%s): %w", d.Id(), err)
		}

		if err := waitForEc2AvailabilityZoneGroupOptInStatus(conn, d.Id(), configurationOptInStatus); err != nil {
			return fmt.Errorf("error waiting for EC2 Availability Zone Group (%s) opt-in status update: %w", d.Id(), err)
		}
	}

	return resourceAwsEc2AvailabilityZoneGroupRead(d, meta)
}

func resourceAwsEc2AvailabilityZoneGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	availabilityZone, err := ec2DescribeAvailabilityZoneGroup(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error describing EC2 Availability Zone Group (%s): %w", d.Id(), err)
	}

	if aws.StringValue(availabilityZone.OptInStatus) == ec2.AvailabilityZoneOptInStatusOptInNotRequired {
		return fmt.Errorf("unnecessary handling of EC2 Availability Zone Group (%s), status: %s", d.Id(), ec2.AvailabilityZoneOptInStatusOptInNotRequired)
	}

	d.Set("group_name", availabilityZone.GroupName)
	d.Set("opt_in_status", availabilityZone.OptInStatus)

	return nil
}

func resourceAwsEc2AvailabilityZoneGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	optInStatus := d.Get("opt_in_status").(string)

	input := &ec2.ModifyAvailabilityZoneGroupInput{
		GroupName:   aws.String(d.Id()),
		OptInStatus: aws.String(optInStatus),
	}

	if _, err := conn.ModifyAvailabilityZoneGroup(input); err != nil {
		return fmt.Errorf("error modifying EC2 Availability Zone Group (%s): %w", d.Id(), err)
	}

	if err := waitForEc2AvailabilityZoneGroupOptInStatus(conn, d.Id(), optInStatus); err != nil {
		return fmt.Errorf("error waiting for EC2 Availability Zone Group (%s) opt-in status update: %w", d.Id(), err)
	}

	return resourceAwsEc2AvailabilityZoneGroupRead(d, meta)
}

func ec2DescribeAvailabilityZoneGroup(conn *ec2.EC2, groupName string) (*ec2.AvailabilityZone, error) {
	input := &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: aws.StringSlice([]string{groupName}),
			},
		},
	}

	output, err := conn.DescribeAvailabilityZones(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.AvailabilityZones) == 0 {
		return nil, nil
	}

	for _, availabilityZone := range output.AvailabilityZones {
		if availabilityZone == nil {
			continue
		}

		if aws.StringValue(availabilityZone.GroupName) == groupName {
			return availabilityZone, nil
		}
	}

	return nil, nil
}

func ec2AvailabilityZoneGroupOptInStatusRefreshFunc(conn *ec2.EC2, groupName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		availabilityZone, err := ec2DescribeAvailabilityZoneGroup(conn, groupName)

		if err != nil {
			return nil, "", fmt.Errorf("error describing EC2 Availability Zone Group (%s): %w", groupName, err)
		}

		if availabilityZone == nil {
			return nil, "", fmt.Errorf("error describing EC2 Availability Zone Group (%s): not found", groupName)
		}

		return availabilityZone, aws.StringValue(availabilityZone.OptInStatus), nil
	}
}

func waitForEc2AvailabilityZoneGroupOptInStatus(conn *ec2.EC2, groupName string, optInStatus string) error {
	pending := ec2.AvailabilityZoneOptInStatusNotOptedIn

	if optInStatus == ec2.AvailabilityZoneOptInStatusNotOptedIn {
		pending = ec2.AvailabilityZoneOptInStatusOptedIn
	}

	stateConf := &resource.StateChangeConf{
		Pending:                   []string{pending},
		Target:                    []string{optInStatus},
		Refresh:                   ec2AvailabilityZoneGroupOptInStatusRefreshFunc(conn, groupName),
		Timeout:                   10 * time.Minute,
		Delay:                     10 * time.Second,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 3,
	}

	log.Printf("[DEBUG] Waiting for EC2 Availability Zone Group (%s) opt-in status update", groupName)
	_, err := stateConf.WaitForState()

	return err
}
