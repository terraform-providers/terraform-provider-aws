package aws

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsNetworkAclRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkAclRuleCreate,
		Read:   resourceAwsNetworkAclRuleRead,
		Delete: resourceAwsNetworkAclRuleDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected NETWORK_ACL_ID:RULE_NUMBER:PROTOCOL:EGRESS", d.Id())
				}
				networkAclID := idParts[0]
				ruleNumber, err := strconv.Atoi(idParts[1])
				if err != nil {
					return nil, err
				}
				protocol := idParts[2]
				egress, err := strconv.ParseBool(idParts[3])
				if err != nil {
					return nil, err
				}

				d.Set("network_acl_id", networkAclID)
				d.Set("rule_number", ruleNumber)
				d.Set("egress", egress)
				d.SetId(networkAclIdRuleNumberEgressHash(networkAclID, ruleNumber, egress, protocol))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"network_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rule_number": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"egress": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					pi := protocolIntegers()
					if val, ok := pi[old]; ok {
						old = strconv.Itoa(val)
					}
					if val, ok := pi[new]; ok {
						new = strconv.Itoa(val)
					}

					return old == new
				},
			},
			"rule_action": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cidr_block": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"ipv6_cidr_block"},
			},
			"ipv6_cidr_block": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"cidr_block"},
			},
			"from_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"to_port": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"icmp_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateICMPArgumentValue,
			},
			"icmp_code": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateICMPArgumentValue,
			},
		},
	}
}

func resourceAwsNetworkAclRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	egress := d.Get("egress").(bool)
	networkAclID := d.Get("network_acl_id").(string)
	ruleNumber := d.Get("rule_number").(int)

	protocol := d.Get("protocol").(string)
	p, protocolErr := strconv.Atoi(protocol)
	if protocolErr != nil {
		var ok bool
		p, ok = protocolIntegers()[protocol]
		if !ok {
			return fmt.Errorf("Invalid Protocol %s for rule %d", protocol, d.Get("rule_number").(int))
		}
	}
	log.Printf("[INFO] Transformed Protocol %s into %d", protocol, p)

	params := &ec2.CreateNetworkAclEntryInput{
		NetworkAclId: aws.String(networkAclID),
		Egress:       aws.Bool(egress),
		RuleNumber:   aws.Int64(int64(ruleNumber)),
		Protocol:     aws.String(strconv.Itoa(p)),
		RuleAction:   aws.String(d.Get("rule_action").(string)),
		PortRange: &ec2.PortRange{
			From: aws.Int64(int64(d.Get("from_port").(int))),
			To:   aws.Int64(int64(d.Get("to_port").(int))),
		},
	}

	cidr, hasCidr := d.GetOk("cidr_block")
	ipv6Cidr, hasIpv6Cidr := d.GetOk("ipv6_cidr_block")

	if !hasCidr && !hasIpv6Cidr {
		return fmt.Errorf("Either `cidr_block` or `ipv6_cidr_block` must be defined")
	}

	if hasCidr {
		params.CidrBlock = aws.String(cidr.(string))
	}

	if hasIpv6Cidr {
		params.Ipv6CidrBlock = aws.String(ipv6Cidr.(string))
	}

	// Specify additional required fields for ICMP. For the list
	// of ICMP codes and types, see: https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml
	if p == 1 || p == 58 {
		params.IcmpTypeCode = &ec2.IcmpTypeCode{}
		if v, ok := d.GetOk("icmp_type"); ok {
			icmpType, err := strconv.Atoi(v.(string))
			if err != nil {
				return fmt.Errorf("Unable to parse ICMP type %s for rule %d", v, ruleNumber)
			}
			params.IcmpTypeCode.Type = aws.Int64(int64(icmpType))
			log.Printf("[DEBUG] Got ICMP type %d for rule %d", icmpType, ruleNumber)
		}
		if v, ok := d.GetOk("icmp_code"); ok {
			icmpCode, err := strconv.Atoi(v.(string))
			if err != nil {
				return fmt.Errorf("Unable to parse ICMP code %s for rule %d", v, ruleNumber)
			}
			params.IcmpTypeCode.Code = aws.Int64(int64(icmpCode))
			log.Printf("[DEBUG] Got ICMP code %d for rule %d", icmpCode, ruleNumber)
		}
	}

	log.Printf("[INFO] Creating Network Acl Rule: %d (%t)", ruleNumber, egress)
	_, err := conn.CreateNetworkAclEntry(params)

	if err != nil {
		return fmt.Errorf("error creating Network ACL (%s) Egress (%t) Rule (%d): %w", networkAclID, egress, ruleNumber, err)
	}

	d.SetId(networkAclIdRuleNumberEgressHash(networkAclID, ruleNumber, egress, d.Get("protocol").(string)))

	return resourceAwsNetworkAclRuleRead(d, meta)
}

func resourceAwsNetworkAclRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	egress := d.Get("egress").(bool)
	networkAclID := d.Get("network_acl_id").(string)
	ruleNumber := d.Get("rule_number").(int)

	var resp *ec2.NetworkAclEntry

	err := resource.Retry(waiter.NetworkAclEntryPropagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = finder.NetworkAclEntry(conn, networkAclID, egress, ruleNumber)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidNetworkAclID.NotFound") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && resp == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("EC2 Network ACL (%s) Egress (%t) Rule (%d) not found", networkAclID, egress, ruleNumber),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = finder.NetworkAclEntry(conn, networkAclID, egress, ruleNumber)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidNetworkAclID.NotFound") {
		log.Printf("[WARN] EC2 Network ACL (%s) not found, removing from state", networkAclID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL (%s) Egress (%t) Rule (%d): %w", networkAclID, egress, ruleNumber, err)
	}

	if resp == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 Network ACL (%s) Egress (%t) Rule (%d): not found after creation", networkAclID, egress, ruleNumber)
		}

		log.Printf("[WARN] EC2 Network ACL (%s) Egress (%t) Rule (%d) not found, removing from state", networkAclID, egress, ruleNumber)
		d.SetId("")
		return nil
	}

	d.Set("rule_number", resp.RuleNumber)
	d.Set("cidr_block", resp.CidrBlock)
	d.Set("ipv6_cidr_block", resp.Ipv6CidrBlock)
	d.Set("egress", resp.Egress)
	if resp.IcmpTypeCode != nil {
		d.Set("icmp_code", strconv.FormatInt(aws.Int64Value(resp.IcmpTypeCode.Code), 10))
		d.Set("icmp_type", strconv.FormatInt(aws.Int64Value(resp.IcmpTypeCode.Type), 10))
	}
	if resp.PortRange != nil {
		d.Set("from_port", resp.PortRange.From)
		d.Set("to_port", resp.PortRange.To)
	}

	d.Set("rule_action", resp.RuleAction)

	p, protocolErr := strconv.Atoi(*resp.Protocol)
	log.Printf("[INFO] Converting the protocol %v", p)
	if protocolErr == nil {
		var ok bool
		protocol, ok := protocolStrings(protocolIntegers())[p]
		if !ok {
			return fmt.Errorf("Invalid Protocol %s for rule %d", *resp.Protocol, d.Get("rule_number").(int))
		}
		log.Printf("[INFO] Transformed Protocol %s back into %s", *resp.Protocol, protocol)
		d.Set("protocol", protocol)
	}

	return nil
}

func resourceAwsNetworkAclRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DeleteNetworkAclEntryInput{
		NetworkAclId: aws.String(d.Get("network_acl_id").(string)),
		RuleNumber:   aws.Int64(int64(d.Get("rule_number").(int))),
		Egress:       aws.Bool(d.Get("egress").(bool)),
	}

	log.Printf("[INFO] Deleting Network Acl Rule: %s", d.Id())
	_, err := conn.DeleteNetworkAclEntry(params)
	if err != nil {
		return fmt.Errorf("Error Deleting Network Acl Rule: %s", err.Error())
	}

	return nil
}

func networkAclIdRuleNumberEgressHash(networkAclId string, ruleNumber int, egress bool, protocol string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", networkAclId))
	buf.WriteString(fmt.Sprintf("%d-", ruleNumber))
	buf.WriteString(fmt.Sprintf("%t-", egress))
	buf.WriteString(fmt.Sprintf("%s-", protocol))
	return fmt.Sprintf("nacl-%d", hashcode.String(buf.String()))
}

func validateICMPArgumentValue(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := strconv.Atoi(value)
	if len(value) == 0 || err != nil {
		errors = append(errors, fmt.Errorf("%q must be an integer value: %q", k, value))
	}
	return
}
