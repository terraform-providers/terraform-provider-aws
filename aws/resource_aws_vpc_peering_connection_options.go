package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
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

	pcRaw, _, err := vpcPeeringConnectionRefreshState(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error reading VPC Peering Connection: %s", err)
	}

	if pcRaw == nil {
		log.Printf("[WARN] VPC Peering Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	pc := pcRaw.(*ec2.VpcPeeringConnection)

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

	pcRaw, _, err := vpcPeeringConnectionRefreshState(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error reading VPC Peering Connection (%s): %s", d.Id(), err)
	}

	if pcRaw == nil {
		log.Printf("[WARN] VPC Peering Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	pc := pcRaw.(*ec2.VpcPeeringConnection)

	crossRegionPeering := false
	if aws.StringValue(pc.RequesterVpcInfo.Region) != aws.StringValue(pc.AccepterVpcInfo.Region) {
		crossRegionPeering = true
	}
	if err := resourceAwsVpcPeeringConnectionModifyOptions(d, meta, crossRegionPeering); err != nil {
		return fmt.Errorf("error modifying VPC Peering Connection (%s) Options: %s", d.Id(), err)
	}

	return resourceAwsVpcPeeringConnectionOptionsRead(d, meta)
}

func resourceAwsVpcPeeringConnectionOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	// Don't do anything with the underlying VPC peering connection.
	return nil
}
