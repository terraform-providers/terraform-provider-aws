package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/waiter"
)

func resourceAwsRoute53ResolverFirewallRuleGroupAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverFirewallRuleGroupAssociationCreate,
		Read:   resourceAwsRoute53ResolverFirewallRuleGroupAssociationRead,
		Update: resourceAwsRoute53ResolverFirewallRuleGroupAssociationUpdate,
		Delete: resourceAwsRoute53ResolverFirewallRuleGroupAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateRoute53ResolverName,
			},

			"firewall_rule_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"mutation_protection": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.MutationProtectionStatus_Values(), false),
			},

			"priority": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsRoute53ResolverFirewallRuleGroupAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &route53resolver.AssociateFirewallRuleGroupInput{
		CreatorRequestId:    aws.String(resource.PrefixedUniqueId("tf-r53-rslvr-frgassoc-")),
		Name:                aws.String(d.Get("name").(string)),
		FirewallRuleGroupId: aws.String(d.Get("firewall_rule_group_id").(string)),
		Priority:            aws.Int64(int64(d.Get("priority").(int))),
		VpcId:               aws.String(d.Get("vpc_id").(string)),
		Tags:                tags.IgnoreAws().Route53resolverTags(),
	}

	if v, ok := d.GetOk("mutation_protection"); ok {
		input.MutationProtection = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver DNS Firewall rule group association: %#v", input)
	output, err := conn.AssociateFirewallRuleGroup(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall rule group association: %w", err)
	}

	d.SetId(aws.StringValue(output.FirewallRuleGroupAssociation.Id))

	_, err = waiter.FirewallRuleGroupAssociationCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver DNS Firewall rule group association (%s) to become available: %w", d.Id(), err)
	}

	return resourceAwsRoute53ResolverFirewallRuleGroupAssociationRead(d, meta)
}

func resourceAwsRoute53ResolverFirewallRuleGroupAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	ruleGroupAssociation, err := finder.FirewallRuleGroupAssociationByID(conn, d.Id())

	if !d.IsNewResource() && isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Route53 Resolver DNS Firewall rule group association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Resolver DNS Firewall rule group association (%s): %w", d.Id(), err)
	}

	if ruleGroupAssociation == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error getting Route 53 Resolver DNS Firewall rule group association (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] Route 53 Resolver DNS Firewall rule group association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(ruleGroupAssociation.Arn)
	d.Set("arn", arn)
	d.Set("name", ruleGroupAssociation.Name)
	d.Set("firewall_rule_group_id", ruleGroupAssociation.FirewallRuleGroupId)
	d.Set("mutation_protection", ruleGroupAssociation.MutationProtection)
	d.Set("priority", ruleGroupAssociation.Priority)
	d.Set("vpc_id", ruleGroupAssociation.VpcId)

	tags, err := keyvaluetags.Route53resolverListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver DNS Firewall rule group association (%s): %w", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsRoute53ResolverFirewallRuleGroupAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	if d.HasChanges("name", "mutation_protection", "priority") {
		input := &route53resolver.UpdateFirewallRuleGroupAssociationInput{
			FirewallRuleGroupAssociationId: aws.String(d.Id()),
			Name:                           aws.String(d.Get("name").(string)),
			Priority:                       aws.Int64(int64(d.Get("priority").(int))),
		}

		if v, ok := d.GetOk("mutation_protection"); ok {
			input.MutationProtection = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating Route 53 Resolver DNS Firewall rule group association: %#v", input)
		_, err := conn.UpdateFirewallRuleGroupAssociation(input)
		if err != nil {
			return fmt.Errorf("error creating Route 53 Resolver DNS Firewall rule group association: %w", err)
		}

		_, err = waiter.FirewallRuleGroupAssociationUpdated(conn, d.Id())

		if err != nil {
			return fmt.Errorf("error waiting for Route53 Resolver DNS Firewall rule group association (%s) to be updated: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.Route53resolverUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver DNS Firewall rule group association (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsRoute53ResolverFirewallRuleGroupAssociationRead(d, meta)
}

func resourceAwsRoute53ResolverFirewallRuleGroupAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	_, err := conn.DisassociateFirewallRuleGroup(&route53resolver.DisassociateFirewallRuleGroupInput{
		FirewallRuleGroupAssociationId: aws.String(d.Id()),
	})

	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNS Firewall rule group association (%s): %w", d.Id(), err)
	}

	_, err = waiter.FirewallRuleGroupAssociationDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver DNS Firewall rule group association (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
