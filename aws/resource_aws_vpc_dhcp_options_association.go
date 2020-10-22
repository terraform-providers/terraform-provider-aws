package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsVpcDhcpOptionsAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVpcDhcpOptionsAssociationCreate,
		Read:   resourceAwsVpcDhcpOptionsAssociationRead,
		Update: resourceAwsVpcDhcpOptionsAssociationUpdate,
		Delete: resourceAwsVpcDhcpOptionsAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsVpcDhcpOptionsAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"dhcp_options_id": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsVpcDhcpOptionsAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).ec2conn
	// Provide the vpc_id as the id to import
	vpcRaw, _, err := VPCStateRefreshFunc(conn, d.Id())()
	if err != nil {
		return nil, err
	}
	if vpcRaw == nil {
		return nil, nil
	}
	vpc := vpcRaw.(*ec2.Vpc)
	if err = d.Set("vpc_id", vpc.VpcId); err != nil {
		return nil, err
	}
	if err = d.Set("dhcp_options_id", vpc.DhcpOptionsId); err != nil {
		return nil, err
	}
	d.SetId(fmt.Sprintf("%s-%s", aws.StringValue(vpc.DhcpOptionsId), aws.StringValue(vpc.VpcId)))
	return []*schema.ResourceData{d}, nil
}

func resourceAwsVpcDhcpOptionsAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	vpcId := d.Get("vpc_id").(string)
	optsID := d.Get("dhcp_options_id").(string)

	log.Printf("[INFO] Creating DHCP Options association: %s => %s", vpcId, optsID)

	if _, err := conn.AssociateDhcpOptions(&ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String(optsID),
		VpcId:         aws.String(vpcId),
	}); err != nil {
		return err
	}

	// Set the ID and return
	d.SetId(fmt.Sprintf("%s-%s", optsID, vpcId))

	log.Printf("[INFO] VPC DHCP Association ID: %s", d.Id())

	return resourceAwsVpcDhcpOptionsAssociationRead(d, meta)
}

func resourceAwsVpcDhcpOptionsAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	// Get the VPC that this association belongs to
	vpcRaw, _, err := VPCStateRefreshFunc(conn, d.Get("vpc_id").(string))()

	if err != nil {
		return err
	}

	if vpcRaw == nil {
		return nil
	}

	vpc := vpcRaw.(*ec2.Vpc)
	if aws.StringValue(vpc.VpcId) != d.Get("vpc_id") ||
		aws.StringValue(vpc.DhcpOptionsId) != d.Get("dhcp_options_id") {
		log.Printf("[INFO] It seems the DHCP Options association is gone. Deleting reference from Graph...")
		d.SetId("")
	}

	d.Set("vpc_id", vpc.VpcId)
	d.Set("dhcp_options_id", vpc.DhcpOptionsId)

	return nil
}

// DHCP Options Asociations cannot be updated.
func resourceAwsVpcDhcpOptionsAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsVpcDhcpOptionsAssociationCreate(d, meta)
}

// AWS does not provide an API to disassociate a DHCP Options set from a VPC.
// So, we do this by setting the VPC to the default DHCP Options Set.
func resourceAwsVpcDhcpOptionsAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	log.Printf("[INFO] Disassociating DHCP Options Set %s from VPC %s...", d.Get("dhcp_options_id"), d.Get("vpc_id"))
	_, err := conn.AssociateDhcpOptions(&ec2.AssociateDhcpOptionsInput{
		DhcpOptionsId: aws.String("default"),
		VpcId:         aws.String(d.Get("vpc_id").(string)),
	})

	if isAWSErr(err, "InvalidVpcID.NotFound", "") {
		return nil
	}

	return err
}
