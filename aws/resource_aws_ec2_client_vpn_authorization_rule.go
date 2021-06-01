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

func resourceAwsEc2ClientVpnAuthorizationRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2ClientVpnAuthorizationRuleCreate,
		Read:   resourceAwsEc2ClientVpnAuthorizationRuleRead,
		Delete: resourceAwsEc2ClientVpnAuthorizationRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEc2ClientVpnAuthorizationRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"client_vpn_endpoint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_network_cidr": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"access_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"access_group_id", "authorize_all_groups"},
			},
			"authorize_all_groups": {
				Type:         schema.TypeBool,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"access_group_id", "authorize_all_groups"},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsEc2ClientVpnAuthorizationRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	endpointID := d.Get("client_vpn_endpoint_id").(string)
	targetNetworkCidr := d.Get("target_network_cidr").(string)

	input := &ec2.AuthorizeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(endpointID),
		TargetNetworkCidr:   aws.String(targetNetworkCidr),
	}

	var accessGroupID string
	if v, ok := d.GetOk("access_group_id"); ok {
		accessGroupID = v.(string)
		input.AccessGroupId = aws.String(accessGroupID)
	}

	if v, ok := d.GetOk("authorize_all_groups"); ok {
		input.AuthorizeAllGroups = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	id := tfec2.ClientVpnAuthorizationRuleCreateID(endpointID, targetNetworkCidr, accessGroupID)

	log.Printf("[DEBUG] Creating Client VPN authorization rule: %#v", input)
	_, err := conn.AuthorizeClientVpnIngress(input)
	if err != nil {
		return fmt.Errorf("error creating Client VPN authorization rule %q: %w", id, err)
	}

	_, err = waiter.ClientVpnAuthorizationRuleAuthorized(conn, id)
	if err != nil {
		return fmt.Errorf("error waiting for Client VPN authorization rule %q to be active: %w", id, err)
	}

	d.SetId(id)

	return resourceAwsEc2ClientVpnAuthorizationRuleRead(d, meta)
}

func resourceAwsEc2ClientVpnAuthorizationRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	result, err := finder.ClientVpnAuthorizationRule(conn,
		d.Get("client_vpn_endpoint_id").(string),
		d.Get("target_network_cidr").(string),
		d.Get("access_group_id").(string),
	)

	if isAWSErr(err, tfec2.ErrCodeClientVpnAuthorizationRuleNotFound, "") {
		log.Printf("[WARN] EC2 Client VPN authorization rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading Client VPN authorization rule: %w", err)
	}

	if result == nil || len(result.AuthorizationRules) == 0 || result.AuthorizationRules[0] == nil {
		log.Printf("[WARN] EC2 Client VPN authorization rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	rule := result.AuthorizationRules[0]
	d.Set("client_vpn_endpoint_id", rule.ClientVpnEndpointId)
	d.Set("target_network_cidr", rule.DestinationCidr)
	d.Set("access_group_id", rule.GroupId)
	d.Set("authorize_all_groups", rule.AccessAll)
	d.Set("description", rule.Description)

	return nil
}

func resourceAwsEc2ClientVpnAuthorizationRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	input := &ec2.RevokeClientVpnIngressInput{
		ClientVpnEndpointId: aws.String(d.Get("client_vpn_endpoint_id").(string)),
		TargetNetworkCidr:   aws.String(d.Get("target_network_cidr").(string)),
		RevokeAllGroups:     aws.Bool(d.Get("authorize_all_groups").(bool)),
	}
	if v, ok := d.GetOk("access_group_id"); ok {
		input.AccessGroupId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Revoking Client VPN authorization rule %q", d.Id())
	err := deleteClientVpnAuthorizationRule(conn, input)
	if err != nil {
		return fmt.Errorf("error revoking Client VPN authorization rule %q: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsEc2ClientVpnAuthorizationRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	endpointID, targetNetworkCidr, accessGroupID, err := tfec2.ClientVpnAuthorizationRuleParseID(d.Id())
	if err != nil {
		return nil, err
	}

	d.Set("client_vpn_endpoint_id", endpointID)
	d.Set("target_network_cidr", targetNetworkCidr)
	d.Set("access_group_id", accessGroupID)
	return []*schema.ResourceData{d}, nil
}

func deleteClientVpnAuthorizationRule(conn *ec2.EC2, input *ec2.RevokeClientVpnIngressInput) error {
	id := tfec2.ClientVpnAuthorizationRuleCreateID(
		aws.StringValue(input.ClientVpnEndpointId),
		aws.StringValue(input.TargetNetworkCidr),
		aws.StringValue(input.AccessGroupId))

	_, err := conn.RevokeClientVpnIngress(input)
	if isAWSErr(err, tfec2.ErrCodeClientVpnAuthorizationRuleNotFound, "") {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = waiter.ClientVpnAuthorizationRuleRevoked(conn, id)

	return err
}
