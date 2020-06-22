package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsEc2TrafficMirrorFilterRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEc2TrafficMirrorFilterRuleCreate,
		Read:   resourceAwsEc2TrafficMirrorFilterRuleRead,
		Update: resourceAwsEc2TrafficMirrorFilterRuleUpdate,
		Delete: resourceAwsEc2TrafficMirrorFilterRuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEc2TrafficMirrorFilterRuleImport,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"traffic_mirror_filter_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"destination_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"to_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"protocol": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"rule_action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.TrafficMirrorRuleActionAccept,
					ec2.TrafficMirrorRuleActionReject,
				}, false),
			},
			"rule_number": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"source_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateCIDRNetworkAddress,
			},
			"source_port_range": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"to_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"traffic_direction": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.TrafficDirectionIngress,
					ec2.TrafficDirectionEgress,
				}, false),
			},
		},
	}
}

func resourceAwsEc2TrafficMirrorFilterRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	filterId := d.Get("traffic_mirror_filter_id")

	input := &ec2.CreateTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterId: aws.String(filterId.(string)),
		DestinationCidrBlock:  aws.String(d.Get("destination_cidr_block").(string)),
		SourceCidrBlock:       aws.String(d.Get("source_cidr_block").(string)),
		RuleAction:            aws.String(d.Get("rule_action").(string)),
		RuleNumber:            aws.Int64(int64(d.Get("rule_number").(int))),
		TrafficDirection:      aws.String(d.Get("traffic_direction").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("protocol"); ok {
		input.Protocol = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("destination_port_range"); ok {
		input.DestinationPortRange = buildTrafficMirrorPortRangeRequest(v.([]interface{}))
	}

	if v, ok := d.GetOk("source_port_range"); ok {
		input.SourcePortRange = buildTrafficMirrorPortRangeRequest(v.([]interface{}))
	}

	out, err := conn.CreateTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Traffic Mirror Filter Rule (%s): %w", filterId, err)
	}

	d.SetId(*out.TrafficMirrorFilterRule.TrafficMirrorFilterRuleId)
	return resourceAwsEc2TrafficMirrorFilterRuleRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ruleId := d.Id()

	var rule *ec2.TrafficMirrorFilterRule
	filterId, filterIdSet := d.GetOk("traffic_mirror_filter_id")
	input := &ec2.DescribeTrafficMirrorFiltersInput{}
	if filterIdSet {
		input.TrafficMirrorFilterIds = aws.StringSlice([]string{filterId.(string)})
	}

	err := conn.DescribeTrafficMirrorFiltersPages(input, func(page *ec2.DescribeTrafficMirrorFiltersOutput, lastPage bool) bool {
		rule = findEc2TrafficMirrorFilterRule(ruleId, page.TrafficMirrorFilters)
		return nil == rule
	})

	if err != nil {
		return fmt.Errorf("Error while describing filters: %v", err)
	}

	if nil == rule {
		log.Printf("[WARN] EC2 Traffic Mirror Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(*rule.TrafficMirrorFilterRuleId)
	d.Set("traffic_mirror_filter_id", rule.TrafficMirrorFilterId)
	d.Set("destination_cidr_block", rule.DestinationCidrBlock)
	d.Set("source_cidr_block", rule.SourceCidrBlock)
	d.Set("rule_action", rule.RuleAction)
	d.Set("rule_number", rule.RuleNumber)
	d.Set("traffic_direction", rule.TrafficDirection)
	d.Set("description", rule.Description)
	d.Set("protocol", rule.Protocol)

	if err := d.Set("destination_port_range", buildTrafficMirrorFilterRulePortRangeSchema(rule.DestinationPortRange)); err != nil {
		return fmt.Errorf("error setting destination_port_range: %s", err)
	}

	if err := d.Set("source_port_range", buildTrafficMirrorFilterRulePortRangeSchema(rule.SourcePortRange)); err != nil {
		return fmt.Errorf("error setting source_port_range: %s", err)
	}

	return nil
}

func findEc2TrafficMirrorFilterRule(ruleId string, filters []*ec2.TrafficMirrorFilter) (rule *ec2.TrafficMirrorFilterRule) {
	log.Printf("[DEBUG] searching %s in %d filters", ruleId, len(filters))
	for _, v := range filters {
		log.Printf("[DEBUG]: searching filter %s, ingress rule count = %d, egress rule count = %d", *v.TrafficMirrorFilterId, len(v.IngressFilterRules), len(v.EgressFilterRules))
		for _, r := range v.IngressFilterRules {
			if *r.TrafficMirrorFilterRuleId == ruleId {
				rule = r
				break
			}
		}
		for _, r := range v.EgressFilterRules {
			if *r.TrafficMirrorFilterRuleId == ruleId {
				rule = r
				break
			}
		}
	}

	if nil != rule {
		log.Printf("[DEBUG]: Found %s in %s", ruleId, *rule.TrafficDirection)
	}

	return rule
}

func resourceAwsEc2TrafficMirrorFilterRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ruleId := d.Id()

	input := &ec2.ModifyTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterRuleId: &ruleId,
	}

	var removeFields []*string
	if d.HasChange("protocol") {
		n := d.Get("protocol")
		if n == "0" {
			removeFields = append(removeFields, aws.String(ec2.TrafficMirrorFilterRuleFieldProtocol))
		} else {
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
		}
	}

	if d.HasChange("description") {
		n := d.Get("description")
		if n != "" {
			input.Description = aws.String(n.(string))
		} else {
			removeFields = append(removeFields, aws.String(ec2.TrafficMirrorFilterRuleFieldDescription))
		}
	}

	if d.HasChange("destination_cidr_block") {
		input.DestinationCidrBlock = aws.String(d.Get("destination_cidr_block").(string))
	}

	if d.HasChange("source_cidr_block") {
		input.SourceCidrBlock = aws.String(d.Get("source_cidr_block").(string))
	}

	if d.HasChange("destination_port_range") {
		v := d.Get("destination_port_range")
		n := v.([]interface{})
		if 0 == len(n) {
			removeFields = append(removeFields, aws.String(ec2.TrafficMirrorFilterRuleFieldDestinationPortRange))
		} else {
			//Modify request that adds port range seems to fail if protocol is not set in the request
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
			input.DestinationPortRange = buildTrafficMirrorPortRangeRequest(n)
		}
	}

	if d.HasChange("source_port_range") {
		v := d.Get("source_port_range")
		n := v.([]interface{})
		if 0 == len(n) {
			removeFields = append(removeFields, aws.String(ec2.TrafficMirrorFilterRuleFieldSourcePortRange))
		} else {
			//Modify request that adds port range seems to fail if protocol is not set in the request
			input.Protocol = aws.Int64(int64(d.Get("protocol").(int)))
			input.SourcePortRange = buildTrafficMirrorPortRangeRequest(n)
		}
	}

	if d.HasChange("rule_action") {
		input.RuleAction = aws.String(d.Get("rule_action").(string))
	}

	if d.HasChange("rule_number") {
		input.RuleNumber = aws.Int64(int64(d.Get("rule_action").(int)))
	}

	if d.HasChange("traffic_direction") {
		input.TrafficDirection = aws.String(d.Get("traffic_direction").(string))
	}

	if len(removeFields) > 0 {
		input.SetRemoveFields(removeFields)
	}

	_, err := conn.ModifyTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("error modifying EC2 Traffic Mirror Filter Rule (%s): %w", ruleId, err)
	}

	return resourceAwsEc2TrafficMirrorFilterRuleRead(d, meta)
}

func resourceAwsEc2TrafficMirrorFilterRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	ruleId := d.Id()
	input := &ec2.DeleteTrafficMirrorFilterRuleInput{
		TrafficMirrorFilterRuleId: &ruleId,
	}

	_, err := conn.DeleteTrafficMirrorFilterRule(input)
	if err != nil {
		return fmt.Errorf("error deleting EC2 Traffic Mirror Filter Rule (%s): %w", ruleId, err)
	}

	return nil
}

func buildTrafficMirrorPortRangeRequest(p []interface{}) (out *ec2.TrafficMirrorPortRangeRequest) {
	portSchema := p[0].(map[string]interface{})

	portRange := ec2.TrafficMirrorPortRangeRequest{}
	if v, ok := portSchema["from_port"]; ok {
		portRange.FromPort = aws.Int64(int64(v.(int)))
		out = &portRange
	}

	if v, ok := portSchema["to_port"]; ok {
		portRange.ToPort = aws.Int64(int64(v.(int)))
		out = &portRange
	}

	return out
}

func buildTrafficMirrorFilterRulePortRangeSchema(portRange *ec2.TrafficMirrorPortRange) interface{} {
	if nil == portRange {
		return nil
	}

	var out [1]interface{}
	elem := make(map[string]interface{})
	elem["from_port"] = portRange.FromPort
	elem["to_port"] = portRange.ToPort
	out[0] = elem

	return out
}

func resourceAwsEc2TrafficMirrorFilterRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("unexpected format (%q), expected <filter-id>:<rule-id>", d.Id())
	}

	d.Set("traffic_mirror_filter_id", parts[0])
	d.SetId(parts[1])

	return []*schema.ResourceData{d}, nil
}
