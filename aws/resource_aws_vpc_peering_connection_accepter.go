package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsVpcPeeringConnectionAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsVPCPeeringAccepterCreate,
		Read:   resourceAwsVPCPeeringRead,
		Update: resourceAwsVPCPeeringUpdate,
		Delete: resourceAwsVPCPeeringAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, m interface{}) (result []*schema.ResourceData, err error) {
				d.Set("vpc_peering_connection_id", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"auto_accept": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"accept_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"accepter":  vpcPeeringConnectionOptionsSchema(),
			"requester": vpcPeeringConnectionOptionsSchema(),
			"tags":      tagsSchema(),
		},
	}
}

func resourceAwsVPCPeeringAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	id := d.Get("vpc_peering_connection_id").(string)

	_, statusCode, err := vpcPeeringConnectionRefreshState(conn, id)()

	if err != nil && statusCode != ec2.VpcPeeringConnectionStateReasonCodeFailed {
		return fmt.Errorf("error reading VPC Peering Connection (%s): %s", id, err)
	}

	status := map[string]bool{
		ec2.VpcPeeringConnectionStateReasonCodeDeleted:  true,
		ec2.VpcPeeringConnectionStateReasonCodeDeleting: true,
		ec2.VpcPeeringConnectionStateReasonCodeExpired:  true,
		ec2.VpcPeeringConnectionStateReasonCodeFailed:   true,
		ec2.VpcPeeringConnectionStateReasonCodeRejected: true,
		"": true, // AWS consistency issue, see vpcPeeringConnectionRefreshState
	}
	if _, ok := status[statusCode]; ok {
		return fmt.Errorf("VPC Peering Connection (%s) in unexpected status for acceptance: %s", id, statusCode)
	}

	d.SetId(id)

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	return resourceAwsVPCPeeringUpdate(d, meta)
}

func resourceAwsVPCPeeringAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete VPC peering connection. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}
