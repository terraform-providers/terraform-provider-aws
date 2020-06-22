package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsVpnGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpnGatewayCreate,
		Read:   resourceAwsVpnGatewayRead,
		Update: resourceAwsVpnGatewayUpdate,
		Delete: resourceAwsVpnGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"amazon_side_asn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ValidateFunc: validateAmazonSideAsn,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsVpnGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	createOpts := &ec2.CreateVpnGatewayInput{
		AvailabilityZone: aws.String(d.Get("availability_zone").(string)),
		Type:             aws.String("ipsec.1"),
	}
	if asn, ok := d.GetOk("amazon_side_asn"); ok {
		i, err := strconv.ParseInt(asn.(string), 10, 64)
		if err != nil {
			return err
		}
		createOpts.AmazonSideAsn = aws.Int64(i)
	}

	// Create the VPN gateway
	log.Printf("[DEBUG] Creating VPN gateway")
	resp, err := conn.CreateVpnGateway(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating VPN gateway: %s", err)
	}

	d.SetId(aws.StringValue(resp.VpnGateway.VpnGatewayId))

	if _, ok := d.GetOk("vpc_id"); ok {
		if err := resourceAwsVpnGatewayAttach(d, meta); err != nil {
			return fmt.Errorf("error attaching EC2 VPN Gateway (%s) to VPC: %s", d.Id(), err)
		}
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error adding EC2 VPN Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsVpnGatewayRead(d, meta)
}

func resourceAwsVpnGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVpnGatewayID.NotFound" {
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] Error finding VpnGateway: %s", err)
			return err
		}
	}

	vpnGateway := resp.VpnGateways[0]
	if vpnGateway == nil || *vpnGateway.State == "deleted" {
		// Seems we have lost our VPN gateway
		d.SetId("")
		return nil
	}

	vpnAttachment := vpnGatewayGetAttachment(vpnGateway)
	if vpnAttachment == nil {
		// Gateway exists but not attached to the VPC
		d.Set("vpc_id", "")
	} else {
		d.Set("vpc_id", vpnAttachment.VpcId)
	}

	if vpnGateway.AvailabilityZone != nil && *vpnGateway.AvailabilityZone != "" {
		d.Set("availability_zone", vpnGateway.AvailabilityZone)
	}
	d.Set("amazon_side_asn", strconv.FormatInt(aws.Int64Value(vpnGateway.AmazonSideAsn), 10))

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(vpnGateway.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsVpnGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("vpc_id") {
		// If we're already attached, detach it first
		if err := resourceAwsVpnGatewayDetach(d, meta); err != nil {
			return err
		}

		// Attach the VPN gateway to the new vpc
		if err := resourceAwsVpnGatewayAttach(d, meta); err != nil {
			return err
		}
	}

	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 VPN Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsVpnGatewayRead(d, meta)
}

func resourceAwsVpnGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Detach if it is attached
	if err := resourceAwsVpnGatewayDetach(d, meta); err != nil {
		return err
	}

	log.Printf("[INFO] Deleting VPN gateway: %s", d.Id())
	input := &ec2.DeleteVpnGatewayInput{
		VpnGatewayId: aws.String(d.Id()),
	}
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteVpnGateway(input)
		if err == nil {
			return nil
		}

		if isAWSErr(err, "InvalidVpnGatewayID.NotFound", "") {
			return nil
		}
		if isAWSErr(err, "IncorrectState", "") {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteVpnGateway(input)
		if isAWSErr(err, "InvalidVpnGatewayID.NotFound", "") {
			return nil
		}
	}

	if err != nil {
		return fmt.Errorf("Error deleting VPN gateway: %s", err)
	}
	return nil
}

func resourceAwsVpnGatewayAttach(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcId := d.Get("vpc_id").(string)

	if vpcId == "" {
		log.Printf(
			"[DEBUG] Not attaching VPN Gateway '%s' as no VPC ID is set",
			d.Id())
		return nil
	}

	log.Printf(
		"[INFO] Attaching VPN Gateway '%s' to VPC '%s'",
		d.Id(),
		vpcId)

	req := &ec2.AttachVpnGatewayInput{
		VpnGatewayId: aws.String(d.Id()),
		VpcId:        aws.String(vpcId),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.AttachVpnGateway(req)
		if err != nil {
			if isAWSErr(err, "InvalidVpnGatewayID.NotFound", "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.AttachVpnGateway(req)
	}

	if err != nil {
		return fmt.Errorf("Error attaching VPN gateway: %s", err)
	}

	// Wait for it to be fully attached before continuing
	log.Printf("[DEBUG] Waiting for VPN gateway (%s) to attach", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"detached", "attaching"},
		Target:  []string{"attached"},
		Refresh: vpnGatewayAttachmentStateRefresh(conn, vpcId, d.Id()),
		Timeout: 15 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for VPN gateway (%s) to attach: %s",
			d.Id(), err)
	}

	return nil
}

func resourceAwsVpnGatewayDetach(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	// Get the old VPC ID to detach from
	vpcIdRaw, _ := d.GetChange("vpc_id")
	vpcId := vpcIdRaw.(string)

	if vpcId == "" {
		log.Printf(
			"[DEBUG] Not detaching VPN Gateway '%s' as no VPC ID is set",
			d.Id())
		return nil
	}

	log.Printf(
		"[INFO] Detaching VPN Gateway '%s' from VPC '%s'",
		d.Id(),
		vpcId)

	wait := true
	_, err := conn.DetachVpnGateway(&ec2.DetachVpnGatewayInput{
		VpnGatewayId: aws.String(d.Id()),
		VpcId:        aws.String(vpcId),
	})
	if err != nil {
		ec2err, ok := err.(awserr.Error)
		if ok {
			if ec2err.Code() == "InvalidVpnGatewayID.NotFound" {
				err = nil
				wait = false
			} else if ec2err.Code() == "InvalidVpnGatewayAttachment.NotFound" {
				err = nil
				wait = false
			}
		}

		if err != nil {
			return err
		}
	}

	if !wait {
		return nil
	}

	// Wait for it to be fully detached before continuing
	log.Printf("[DEBUG] Waiting for VPN gateway (%s) to detach", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"attached", "detaching", "available"},
		Target:  []string{"detached"},
		Refresh: vpnGatewayAttachmentStateRefresh(conn, vpcId, d.Id()),
		Timeout: 10 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for vpn gateway (%s) to detach: %s",
			d.Id(), err)
	}

	return nil
}

// vpnGatewayGetAttachment returns any VGW attachment that's in "attached" state or nil.
func vpnGatewayGetAttachment(vgw *ec2.VpnGateway) *ec2.VpcAttachment {
	for _, vpcAttachment := range vgw.VpcAttachments {
		if aws.StringValue(vpcAttachment.State) == ec2.AttachmentStatusAttached {
			return vpcAttachment
		}
	}
	return nil
}
