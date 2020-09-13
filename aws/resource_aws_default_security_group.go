package aws

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

const DefaultSecurityGroupName = "default"

func resourceAwsDefaultSecurityGroup() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsDefaultSecurityGroupCreate,
		Read:   resourceAwsDefaultSecurityGroupRead,
		Update: resourceAwsDefaultSecurityGroupUpdate,
		Delete: resourceAwsDefaultSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsDefaultSecurityGroupMigrateState,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"ingress": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_blocks": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRNetworkAddress,
							},
						},
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateSecurityGroupRuleDescription,
						},
						"from_port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"ipv6_cidr_blocks": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRNetworkAddress,
							},
						},
						"prefix_list_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"protocol": {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: protocolStateFunc,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"self": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"to_port": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
				Set: resourceAwsDefaultSecurityGroupRuleHash,
			},
			"egress": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_blocks": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRNetworkAddress,
							},
						},
						"description": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateSecurityGroupRuleDescription,
						},
						"from_port": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"ipv6_cidr_blocks": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateCIDRNetworkAddress,
							},
						},
						"prefix_list_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"protocol": {
							Type:      schema.TypeString,
							Required:  true,
							StateFunc: protocolStateFunc,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"self": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"to_port": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
				Set: resourceAwsDefaultSecurityGroupRuleHash,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			// This is not implemented. Added to prevent breaking changes.
			"revoke_rules_on_delete": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func resourceAwsDefaultSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	securityGroupOpts := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("group-name"),
				Values: aws.StringSlice([]string{DefaultSecurityGroupName}),
			},
		},
	}

	var vpcID string
	if v, ok := d.GetOk("vpc_id"); ok {
		vpcID = v.(string)
		securityGroupOpts.Filters = append(securityGroupOpts.Filters, &ec2.Filter{
			Name:   aws.String("vpc-id"),
			Values: aws.StringSlice([]string{vpcID}),
		})
	}

	var err error
	log.Printf("[DEBUG] Commandeer Default Security Group: %s", securityGroupOpts)
	resp, err := conn.DescribeSecurityGroups(securityGroupOpts)
	if err != nil {
		return fmt.Errorf("Error creating Default Security Group: %w", err)
	}

	var g *ec2.SecurityGroup
	if vpcID != "" {
		// if vpcId contains a value, then we expect just a single Security Group
		// returned, as default is a protected name for each VPC, and for each
		// Region on EC2 Classic
		if len(resp.SecurityGroups) != 1 {
			return fmt.Errorf("Error finding default security group; found (%d) groups: %s", len(resp.SecurityGroups), resp)
		}
		g = resp.SecurityGroups[0]
	} else {
		// we need to filter through any returned security groups for the group
		// named "default", and does not belong to a VPC
		for _, sg := range resp.SecurityGroups {
			if sg.VpcId == nil && aws.StringValue(sg.GroupName) == DefaultSecurityGroupName {
				g = sg
				break
			}
		}
	}

	if g == nil {
		return fmt.Errorf("Error finding default security group: no matching group found")
	}

	d.SetId(aws.StringValue(g.GroupId))

	log.Printf("[INFO] Default Security Group ID: %s", d.Id())

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(conn, d.Id(), v); err != nil {
			return fmt.Errorf("error adding EC2 Default Security Group (%s) tags: %w", d.Id(), err)
		}
	}

	if err := revokeDefaultSecurityGroupRules(meta, g); err != nil {
		return err
	}

	return resourceAwsDefaultSecurityGroupUpdate(d, meta)
}

func resourceAwsDefaultSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	group, err := finder.SecurityGroupByID(conn, d.Id())
	if err != nil {
		return err
	}
	if group == nil {
		log.Printf("[WARN] Security group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	remoteIngressRules := resourceAwsSecurityGroupIPPermGather(d.Id(), group.IpPermissions, group.OwnerId)
	remoteEgressRules := resourceAwsSecurityGroupIPPermGather(d.Id(), group.IpPermissionsEgress, group.OwnerId)

	localIngressRules := d.Get("ingress").(*schema.Set).List()
	localEgressRules := d.Get("egress").(*schema.Set).List()

	// Loop through the local state of rules, doing a match against the remote
	// ruleSet we built above.
	ingressRules := matchRules("ingress", localIngressRules, remoteIngressRules)
	egressRules := matchRules("egress", localEgressRules, remoteEgressRules)

	sgArn := arn.ARN{
		AccountID: aws.StringValue(group.OwnerId),
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("security-group/%s", aws.StringValue(group.GroupId)),
		Service:   ec2.ServiceName,
	}

	d.Set("arn", sgArn.String())
	d.Set("description", group.Description)
	d.Set("name", group.GroupName)
	d.Set("owner_id", group.OwnerId)
	d.Set("vpc_id", group.VpcId)

	if err := d.Set("ingress", ingressRules); err != nil {
		return fmt.Errorf("error setting Ingress rule set for (%s): %w", d.Id(), err)
	}

	if err := d.Set("egress", egressRules); err != nil {
		return fmt.Errorf("error setting Egress rule set for (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(group.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags for (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsDefaultSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	group, err := finder.SecurityGroupByID(conn, d.Id())
	if err != nil {
		return err
	}
	if group == nil {
		log.Printf("[WARN] Security group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	err = resourceAwsSecurityGroupUpdateRules(d, "ingress", meta, group)
	if err != nil {
		return err
	}

	if d.Get("vpc_id") != nil {
		err = resourceAwsSecurityGroupUpdateRules(d, "egress", meta, group)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") && !d.IsNewResource() {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Security Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsDefaultSecurityGroupRead(d, meta)
}

func resourceAwsDefaultSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Cannot destroy Default Security Group. Terraform will remove this resource from the state file, however resources may remain.")
	return nil
}

func revokeDefaultSecurityGroupRules(meta interface{}, g *ec2.SecurityGroup) error {
	conn := meta.(*AWSClient).ec2conn

	groupID := aws.StringValue(g.GroupId)
	log.Printf("[WARN] Removing all ingress and egress rules found on Default Security Group (%s)", groupID)
	if len(g.IpPermissionsEgress) > 0 {
		req := &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       g.GroupId,
			IpPermissions: g.IpPermissionsEgress,
		}

		log.Printf("[DEBUG] Revoking default egress rules for Default Security Group for %s", groupID)
		if _, err := conn.RevokeSecurityGroupEgress(req); err != nil {
			return fmt.Errorf("error revoking default egress rules for Default Security Group (%s): %w", groupID, err)
		}
	}
	if len(g.IpPermissions) > 0 {
		// a limitation in EC2 Classic is that a call to RevokeSecurityGroupIngress
		// cannot contain both the GroupName and the GroupId
		for _, p := range g.IpPermissions {
			for _, uigp := range p.UserIdGroupPairs {
				if uigp.GroupId != nil && uigp.GroupName != nil {
					uigp.GroupName = nil
				}
			}
		}
		req := &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       g.GroupId,
			IpPermissions: g.IpPermissions,
		}

		log.Printf("[DEBUG] Revoking default ingress rules for Default Security Group for (%s): %s", groupID, req)
		if _, err := conn.RevokeSecurityGroupIngress(req); err != nil {
			return fmt.Errorf("Error revoking default ingress rules for Default Security Group (%s): %w", groupID, err)
		}
	}

	return nil
}

func resourceAwsDefaultSecurityGroupRuleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["from_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["to_port"].(int)))
	p := protocolForValue(m["protocol"].(string))
	buf.WriteString(fmt.Sprintf("%s-", p))
	buf.WriteString(fmt.Sprintf("%t-", m["self"].(bool)))

	// We need to make sure to sort the strings below so that we always
	// generate the same hash code no matter what is in the set.
	if v, ok := m["cidr_blocks"]; ok {
		vs := v.([]interface{})
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["ipv6_cidr_blocks"]; ok {
		vs := v.([]interface{})
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["prefix_list_ids"]; ok {
		vs := v.([]interface{})
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["security_groups"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if m["description"].(string) != "" {
		buf.WriteString(fmt.Sprintf("%s-", m["description"].(string)))
	}

	return hashcode.String(buf.String())
}

func resourceAwsDefaultSecurityGroupMigrateState(
	v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	switch v {
	case 0:
		log.Println("[INFO] Found AWS Default Security Group state v0; migrating to v1")
		return migrateAwsDefaultSecurityGroupStateV0toV1(is)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateAwsDefaultSecurityGroupStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	if is.Empty() || is.Attributes == nil {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	// set default for revoke_rules_on_delete
	is.Attributes["revoke_rules_on_delete"] = "false"

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil
}
