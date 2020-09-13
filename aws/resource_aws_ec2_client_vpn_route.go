package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsEc2ClientVpnRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnRouteCreate,
		Read:   resourceAwsEc2ClientVpnRouteRead,
		Delete: resourceAwsEc2ClientVpnRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEc2ClientVpnRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateIpv4CIDRNetworkAddress,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"target_vpc_subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	targetSubnetID := d.Get("target_vpc_subnet_id").(string)
	destinationCidr := d.Get("destination_cidr_block").(string)

	req := &ec2.CreateClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(endpointID),
		DestinationCidrBlock: aws.String(destinationCidr),
		TargetVpcSubnetId:    aws.String(targetSubnetID),
	}

	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	id := tfec2.ClientVpnRouteCreateID(endpointID, targetSubnetID, destinationCidr)

	_, err := conn.CreateClientVpnRoute(req)

	if err != nil {
		return fmt.Errorf("error creating client VPN route %q: %w", id, err)
	}

	d.SetId(id)

	return resourceAwsEc2ClientVpnRouteRead(d, meta)
}

func resourceAwsEc2ClientVpnRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	resp, err := finder.ClientVpnRoute(conn,
		d.Get("client_vpn_endpoint_id").(string),
		d.Get("target_vpc_subnet_id").(string),
		d.Get("destination_cidr_block").(string),
	)

	if isAWSErr(err, tfec2.ErrCodeClientVpnRouteNotFound, "") {
		log.Printf("[WARN] EC2 Client VPN Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading client VPN route (%s): %w", d.Id(), err)
	}

	if resp == nil || len(resp.Routes) == 0 || resp.Routes[0] == nil {
		log.Printf("[WARN] EC2 Client VPN Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if len(resp.Routes) > 1 {
		return fmt.Errorf("internal error: found %d results for Client VPN route (%s) status, need 1", len(resp.Routes), d.Id())
	}

	route := resp.Routes[0]
	d.Set("client_vpn_endpoint_id", route.ClientVpnEndpointId)
	d.Set("destination_cidr_block", route.DestinationCidr)
	d.Set("description", route.Description)
	d.Set("target_vpc_subnet_id", route.TargetSubnet)
	d.Set("origin", route.Origin)
	d.Set("type", route.Type)

	return nil
}

func resourceAwsEc2ClientVpnRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	err := deleteClientVpnRoute(conn, &ec2.DeleteClientVpnRouteInput{
		ClientVpnEndpointId:  aws.String(d.Get("client_vpn_endpoint_id").(string)),
		DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
		TargetVpcSubnetId:    aws.String(d.Get("target_vpc_subnet_id").(string)),
	})
	if err != nil {
		return fmt.Errorf("error deleting client VPN route %q: %w", d.Id(), err)
	}

	return nil
}

func deleteClientVpnRoute(conn *ec2.EC2, input *ec2.DeleteClientVpnRouteInput) error {
	id := tfec2.ClientVpnRouteCreateID(
		aws.StringValue(input.ClientVpnEndpointId),
		aws.StringValue(input.TargetVpcSubnetId),
		aws.StringValue(input.DestinationCidrBlock),
	)

	_, err := conn.DeleteClientVpnRoute(input)
	if isAWSErr(err, tfec2.ErrCodeClientVpnRouteNotFound, "") {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = waiter.ClientVpnRouteDeleted(conn, id)

	return err
}

func resourceAwsEc2ClientVpnRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	endpointID, targetSubnetID, destinationCidr, err := tfec2.ClientVpnRouteParseID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("client_vpn_endpoint_id", endpointID)
	d.Set("target_vpc_subnet_id", targetSubnetID)
	d.Set("destination_cidr_block", destinationCidr)

	return []*schema.ResourceData{d}, nil
}
