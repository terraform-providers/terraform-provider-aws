package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
)

func resourceAwsEc2ClientVpnNetworkAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnNetworkAssociationCreate,
		Read:   resourceAwsEc2ClientVpnNetworkAssociationRead,
		Update: resourceAwsEc2ClientVpnNetworkAssociationUpdate,
		Delete: resourceAwsEc2ClientVpnNetworkAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEc2ClientVpnNetworkAssociationImport,
		},

		Schema: map[string]*schema.Schema{
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				MinItems: 1,
				MaxItems: 5,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnNetworkAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.AssociateClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		SubnetId:            aws.String(d.Get("subnet_id").(string)),
	}

	log.Printf("[DEBUG] Creating Client VPN network association: %#v", req)
	resp, err := conn.AssociateClientVpnTargetNetwork(req)
	if err != nil {
		return fmt.Errorf("Error creating Client VPN network association: %w", err)
	}

	d.SetId(aws.StringValue(resp.AssociationId))

	log.Printf("[DEBUG] Waiting for Client VPN endpoint to associate with target network: %s", d.Id())
	targetNetwork, err := waiter.ClientVpnNetworkAssociationAssociated(conn, d.Id(), d.Get("client_vpn_endpoint_id").(string))
	if err != nil {
		return fmt.Errorf("error waiting for Client VPN endpoint to associate with target network: %w", err)
	}

	if v, ok := d.GetOk("security_groups"); ok {
		sgReq := &ec2.ApplySecurityGroupsToClientVpnTargetNetworkInput{
			ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
			VpcId:               targetNetwork.VpcId,
			SecurityGroupIds:    expandStringSet(v.(*schema.Set)),
		}

		_, err := conn.ApplySecurityGroupsToClientVpnTargetNetwork(sgReq)
		if err != nil {
			return fmt.Errorf("Error applying security groups to Client VPN network association: %s", err)
		}
	}

	return resourceAwsEc2ClientVpnNetworkAssociationRead(d, meta)
}

func resourceAwsEc2ClientVpnNetworkAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChange("security_groups") {
		input := &ec2.ApplySecurityGroupsToClientVpnTargetNetworkInput{
			ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
			SecurityGroupIds:    expandStringSet(d.Get("security_groups").(*schema.Set)),
			VpcId:               aws.String(d.Get("vpc_id").(string)),
		}

		if _, err := conn.ApplySecurityGroupsToClientVpnTargetNetwork(input); err != nil {
			return fmt.Errorf("error applying security groups to Client VPN Target Network: %s", err)
		}
	}

	return resourceAwsEc2ClientVpnNetworkAssociationRead(d, meta)
}

func resourceAwsEc2ClientVpnNetworkAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	var err error

	result, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		AssociationIds:      []*string{aws.String(d.Id())},
	})

	if isAWSErr(err, tfec2.ErrCodeClientVpnAssociationIdNotFound, "") || isAWSErr(err, tfec2.ErrCodeClientVpnEndpointIdNotFound, "") {
		log.Printf("[WARN] EC2 Client VPN Network Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading Client VPN network association: %w", err)
	}

	if result == nil || len(result.ClientVpnTargetNetworks) == 0 || result.ClientVpnTargetNetworks[0] == nil {
		log.Printf("[WARN] EC2 Client VPN Network Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	network := result.ClientVpnTargetNetworks[0]
	if network.Status != nil && aws.StringValue(network.Status.Code) == ec2.AssociationStatusCodeDisassociated {
		log.Printf("[WARN] EC2 Client VPN Network Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("client_vpn_endpoint_id", network.ClientVpnEndpointId)
	d.Set("association_id", network.AssociationId)
	d.Set("status", network.Status.Code)
	d.Set("subnet_id", network.TargetNetworkId)
	d.Set("vpc_id", network.VpcId)

	if err := d.Set("security_groups", aws.StringValueSlice(network.SecurityGroups)); err != nil {
		return fmt.Errorf("error setting security_groups: %w", err)
	}

	return nil
}

func resourceAwsEc2ClientVpnNetworkAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	err := deleteClientVpnNetworkAssociation(conn, d.Id(), d.Get("client_vpn_endpoint_id").(string))
	if err != nil {
		return fmt.Errorf("error deleting Client VPN network association: %w", err)
	}

	return nil
}

func deleteClientVpnNetworkAssociation(conn *ec2.EC2, networkAssociationID, clientVpnEndpointID string) error {
	_, err := conn.DisassociateClientVpnTargetNetwork(&ec2.DisassociateClientVpnTargetNetworkInput{
		ClientVpnEndpointId: aws.String(clientVpnEndpointID),
		AssociationId:       aws.String(networkAssociationID),
	})

	if isAWSErr(err, tfec2.ErrCodeClientVpnAssociationIdNotFound, "") || isAWSErr(err, tfec2.ErrCodeClientVpnEndpointIdNotFound, "") {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = waiter.ClientVpnNetworkAssociationDisassociated(conn, networkAssociationID, clientVpnEndpointID)

	return err
}

func resourceAwsEc2ClientVpnNetworkAssociationImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	endpointID, associationID, err := tfec2.ClientVpnNetworkAssociationParseID(d.Id())
	if err != nil {
		return nil, err
	}

	d.SetId(associationID)
	d.Set("client_vpn_endpoint_id", endpointID)
	return []*schema.ResourceData{d}, nil
}
