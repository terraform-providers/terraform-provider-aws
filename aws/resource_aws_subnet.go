package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSubnet() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsSubnetCreate,
		Read:   resourceAwsSubnetRead,
		Update: resourceAwsSubnetUpdate,
		Delete: resourceAwsSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsSubnetMigrateState,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"ipv6_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"availability_zone": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"availability_zone_id"},
			},

			"availability_zone_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"availability_zone"},
			},

			"map_public_ip_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"assign_ipv6_address_on_creation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ipv6_cidr_block_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tagsSchema(),

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	createOpts := &ec2.CreateSubnetInput{
		AvailabilityZone:   aws.String(d.Get("availability_zone").(string)),
		AvailabilityZoneId: aws.String(d.Get("availability_zone_id").(string)),
		CidrBlock:          aws.String(d.Get("cidr_block").(string)),
		VpcId:              aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		createOpts.Ipv6CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		createOpts.OutpostArn = aws.String(v.(string))
	}

	var err error
	resp, err := conn.CreateSubnet(createOpts)

	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	// Get the ID and store it
	subnet := resp.Subnet
	d.SetId(*subnet.SubnetId)
	log.Printf("[INFO] Subnet ID: %s", *subnet.SubnetId)

	// Wait for the Subnet to become available
	log.Printf("[DEBUG] Waiting for subnet (%s) to become available", *subnet.SubnetId)
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Refresh: SubnetStateRefreshFunc(conn, *subnet.SubnetId),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf(
			"Error waiting for subnet (%s) to become ready: %s",
			d.Id(), err)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	// You cannot modify multiple subnet attributes in the same request.
	// Reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySubnetAttribute.html

	if d.Get("assign_ipv6_address_on_creation").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error enabling EC2 Subnet (%s) assign IPv6 address on creation: %s", d.Id(), err)
		}
	}

	if d.Get("map_public_ip_on_launch").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error enabling EC2 Subnet (%s) map public IP on launch: %s", d.Id(), err)
		}
	}

	return resourceAwsSubnetRead(d, meta)
}

func resourceAwsSubnetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidSubnetID.NotFound" {
			// Update state to indicate the subnet no longer exists.
			d.SetId("")
			return nil
		}
		return err
	}
	if resp == nil {
		return nil
	}

	subnet := resp.Subnets[0]

	d.Set("vpc_id", subnet.VpcId)
	d.Set("availability_zone", subnet.AvailabilityZone)
	d.Set("availability_zone_id", subnet.AvailabilityZoneId)
	d.Set("cidr_block", subnet.CidrBlock)
	d.Set("map_public_ip_on_launch", subnet.MapPublicIpOnLaunch)
	d.Set("assign_ipv6_address_on_creation", subnet.AssignIpv6AddressOnCreation)
	d.Set("outpost_arn", subnet.OutpostArn)

	// Make sure those values are set, if an IPv6 block exists it'll be set in the loop
	d.Set("ipv6_cidr_block_association_id", "")
	d.Set("ipv6_cidr_block", "")

	for _, a := range subnet.Ipv6CidrBlockAssociationSet {
		if *a.Ipv6CidrBlockState.State == "associated" { //we can only ever have 1 IPv6 block associated at once
			d.Set("ipv6_cidr_block_association_id", a.AssociationId)
			d.Set("ipv6_cidr_block", a.Ipv6CidrBlock)
			break
		}
	}

	d.Set("arn", subnet.SubnetArn)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(subnet.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("owner_id", subnet.OwnerId)

	return nil
}

func resourceAwsSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Subnet (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("map_public_ip_on_launch") {
		modifyOpts := &ec2.ModifySubnetAttributeInput{
			SubnetId: aws.String(d.Id()),
			MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("map_public_ip_on_launch").(bool)),
			},
		}

		log.Printf("[DEBUG] Subnet modify attributes: %#v", modifyOpts)

		_, err := conn.ModifySubnetAttribute(modifyOpts)

		if err != nil {
			return err
		}
	}

	if d.HasChange("ipv6_cidr_block") {
		// We need to handle that we disassociate the IPv6 CIDR block before we try and associate the new one
		// This could be an issue as, we could error out when we try and add the new one
		// We may need to roll back the state and reattach the old one if this is the case

		_, new := d.GetChange("ipv6_cidr_block")

		if v, ok := d.GetOk("ipv6_cidr_block_association_id"); ok {

			//Firstly we have to disassociate the old IPv6 CIDR Block
			disassociateOps := &ec2.DisassociateSubnetCidrBlockInput{
				AssociationId: aws.String(v.(string)),
			}

			_, err := conn.DisassociateSubnetCidrBlock(disassociateOps)
			if err != nil {
				return err
			}

			// Wait for the CIDR to become disassociated
			log.Printf(
				"[DEBUG] Waiting for IPv6 CIDR (%s) to become disassociated",
				d.Id())
			stateConf := &resource.StateChangeConf{
				Pending: []string{"disassociating", "associated"},
				Target:  []string{"disassociated"},
				Refresh: SubnetIpv6CidrStateRefreshFunc(conn, d.Id(), d.Get("ipv6_cidr_block_association_id").(string)),
				Timeout: 3 * time.Minute,
			}
			if _, err := stateConf.WaitForState(); err != nil {
				return fmt.Errorf(
					"Error waiting for IPv6 CIDR (%s) to become disassociated: %s",
					d.Id(), err)
			}
		}

		//Now we need to try and associate the new CIDR block
		associatesOpts := &ec2.AssociateSubnetCidrBlockInput{
			SubnetId:      aws.String(d.Id()),
			Ipv6CidrBlock: aws.String(new.(string)),
		}

		resp, err := conn.AssociateSubnetCidrBlock(associatesOpts)
		if err != nil {
			//The big question here is, do we want to try and reassociate the old one??
			//If we have a failure here, then we may be in a situation that we have nothing associated
			return err
		}

		// Wait for the CIDR to become associated
		log.Printf(
			"[DEBUG] Waiting for IPv6 CIDR (%s) to become associated",
			d.Id())
		stateConf := &resource.StateChangeConf{
			Pending: []string{"associating", "disassociated"},
			Target:  []string{"associated"},
			Refresh: SubnetIpv6CidrStateRefreshFunc(conn, d.Id(), *resp.Ipv6CidrBlockAssociation.AssociationId),
			Timeout: 3 * time.Minute,
		}
		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for IPv6 CIDR (%s) to become associated: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("assign_ipv6_address_on_creation") {
		modifyOpts := &ec2.ModifySubnetAttributeInput{
			SubnetId: aws.String(d.Id()),
			AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("assign_ipv6_address_on_creation").(bool)),
			},
		}

		log.Printf("[DEBUG] Subnet modify attributes: %#v", modifyOpts)

		_, err := conn.ModifySubnetAttribute(modifyOpts)

		if err != nil {
			return err
		}
	}

	return resourceAwsSubnetRead(d, meta)
}

func resourceAwsSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Deleting subnet: %s", d.Id())

	if err := deleteLingeringLambdaENIs(conn, "subnet-id", d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error deleting Lambda ENIs using subnet (%s): %s", d.Id(), err)
	}

	req := &ec2.DeleteSubnetInput{
		SubnetId: aws.String(d.Id()),
	}

	wait := resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"destroyed"},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			_, err := conn.DeleteSubnet(req)
			if err != nil {
				if apiErr, ok := err.(awserr.Error); ok {
					if apiErr.Code() == "DependencyViolation" {
						// There is some pending operation, so just retry
						// in a bit.
						return 42, "pending", nil
					}

					if apiErr.Code() == "InvalidSubnetID.NotFound" {
						return 42, "destroyed", nil
					}
				}

				return 42, "failure", err
			}

			return 42, "destroyed", nil
		},
	}

	if _, err := wait.WaitForState(); err != nil {
		return fmt.Errorf("error deleting subnet (%s): %s", d.Id(), err)
	}

	return nil
}

// SubnetStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a Subnet.
func SubnetStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(id)},
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidSubnetID.NotFound" {
				resp = nil
			} else {
				log.Printf("Error on SubnetStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		subnet := resp.Subnets[0]
		return subnet, *subnet.State, nil
	}
}

func SubnetIpv6CidrStateRefreshFunc(conn *ec2.EC2, id string, associationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(id)},
		}
		resp, err := conn.DescribeSubnets(opts)
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidSubnetID.NotFound" {
				resp = nil
			} else {
				log.Printf("Error on SubnetIpv6CidrStateRefreshFunc: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		if resp.Subnets[0].Ipv6CidrBlockAssociationSet == nil {
			return nil, "", nil
		}

		for _, association := range resp.Subnets[0].Ipv6CidrBlockAssociationSet {
			if *association.AssociationId == associationId {
				return association, *association.Ipv6CidrBlockState.State, nil
			}
		}

		return nil, "", nil
	}
}
