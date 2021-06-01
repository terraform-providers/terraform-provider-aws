package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsVpcPeeringConnectionOptions() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcPeeringConnectionOptionsCreate,
		Read:   resourceAwsVpcPeeringConnectionOptionsRead,
		Update: resourceAwsVpcPeeringConnectionOptionsUpdate,
		Delete: resourceAwsVpcPeeringConnectionOptionsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"accepter":  vpcPeeringConnectionOptionsSchema(),
			"requester": vpcPeeringConnectionOptionsSchema(),
		},
	}
}

func resourceAwsVpcPeeringConnectionOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("vpc_peering_connection_id").(string))

	return resourceAwsVpcPeeringConnectionOptionsUpdate(d, meta)
}

func resourceAwsVpcPeeringConnectionOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	pc, err := vpcPeeringConnection(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading VPC Peering Connection (%s): %w", d.Id(), err)
	}

	if pc == nil {
		log.Printf("[WARN] VPC Peering Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("vpc_peering_connection_id", pc.VpcPeeringConnectionId)

	if err := d.Set("accepter", flattenVpcPeeringConnectionOptions(pc.AccepterVpcInfo.PeeringOptions)); err != nil {
		return fmt.Errorf("error setting VPC Peering Connection Options accepter information: %s", err)
	}
	if err := d.Set("requester", flattenVpcPeeringConnectionOptions(pc.RequesterVpcInfo.PeeringOptions)); err != nil {
		return fmt.Errorf("error setting VPC Peering Connection Options requester information: %s", err)
	}

	return nil
}

func resourceAwsVpcPeeringConnectionOptionsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	pc, err := vpcPeeringConnection(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error reading VPC Peering Connection (%s): %w", d.Id(), err)
	}

	if pc == nil {
		return fmt.Errorf("VPC Peering Connection (%s) not found", d.Id())
	}

	if d.HasChanges("accepter", "requester") {
		crossRegionPeering := aws.StringValue(pc.RequesterVpcInfo.Region) != aws.StringValue(pc.AccepterVpcInfo.Region)

		input := &ec2.ModifyVpcPeeringConnectionOptionsInput{
			VpcPeeringConnectionId: aws.String(d.Id()),
		}
		if d.HasChange("accepter") {
			input.AccepterPeeringConnectionOptions = expandVpcPeeringConnectionOptions(d.Get("accepter").([]interface{}), crossRegionPeering)
		}
		if d.HasChange("requester") {
			input.RequesterPeeringConnectionOptions = expandVpcPeeringConnectionOptions(d.Get("requester").([]interface{}), crossRegionPeering)
		}

		log.Printf("[DEBUG] Modifying VPC Peering Connection options: %s", input)
		_, err = conn.ModifyVpcPeeringConnectionOptions(input)

		if err != nil {
			return fmt.Errorf("error modifying VPC Peering Connection (%s) Options: %w", d.Id(), err)
		}

		// Retry reading back the modified options to deal with eventual consistency.
		// Often this is to do with a delay transitioning from pending-acceptance to active.
		err = resource.Retry(3*time.Minute, func() *resource.RetryError {
			pc, err = vpcPeeringConnection(conn, d.Id())

			if err != nil {
				return resource.NonRetryableError(err)
			}

			if pc == nil {
				return nil
			}

			if d.HasChange("accepter") && pc.AccepterVpcInfo != nil {
				if aws.BoolValue(pc.AccepterVpcInfo.PeeringOptions.AllowDnsResolutionFromRemoteVpc) != aws.BoolValue(input.AccepterPeeringConnectionOptions.AllowDnsResolutionFromRemoteVpc) ||
					aws.BoolValue(pc.AccepterVpcInfo.PeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVpc) != aws.BoolValue(input.AccepterPeeringConnectionOptions.AllowEgressFromLocalClassicLinkToRemoteVpc) ||
					aws.BoolValue(pc.AccepterVpcInfo.PeeringOptions.AllowEgressFromLocalVpcToRemoteClassicLink) != aws.BoolValue(input.AccepterPeeringConnectionOptions.AllowEgressFromLocalVpcToRemoteClassicLink) {
					return resource.RetryableError(fmt.Errorf("VPC Peering Connection (%s) accepter Options not stable", d.Id()))
				}
			}
			if d.HasChange("requester") && pc.RequesterVpcInfo != nil {
				if aws.BoolValue(pc.RequesterVpcInfo.PeeringOptions.AllowDnsResolutionFromRemoteVpc) != aws.BoolValue(input.RequesterPeeringConnectionOptions.AllowDnsResolutionFromRemoteVpc) ||
					aws.BoolValue(pc.RequesterVpcInfo.PeeringOptions.AllowEgressFromLocalClassicLinkToRemoteVpc) != aws.BoolValue(input.RequesterPeeringConnectionOptions.AllowEgressFromLocalClassicLinkToRemoteVpc) ||
					aws.BoolValue(pc.RequesterVpcInfo.PeeringOptions.AllowEgressFromLocalVpcToRemoteClassicLink) != aws.BoolValue(input.RequesterPeeringConnectionOptions.AllowEgressFromLocalVpcToRemoteClassicLink) {
					return resource.RetryableError(fmt.Errorf("VPC Peering Connection (%s) requester Options not stable", d.Id()))
				}
			}

			return nil
		})
	}

	return resourceAwsVpcPeeringConnectionOptionsRead(d, meta)
}

func resourceAwsVpcPeeringConnectionOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	// Don't do anything with the underlying VPC peering connection.
	return nil
}
