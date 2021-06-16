package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
)

func resourceAwsRoute53ResolverFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverFirewallRuleGroupCreate,
		Read:   resourceAwsRoute53ResolverFirewallRuleGroupRead,
		Update: resourceAwsRoute53ResolverFirewallRuleGroupUpdate,
		Delete: resourceAwsRoute53ResolverFirewallRuleGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRoute53ResolverName,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsRoute53ResolverFirewallRuleGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &route53resolver.CreateFirewallRuleGroupInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-firewall-rule-group-")),
		Name:             aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.Tags = tags.IgnoreAws().Route53resolverTags()
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver DNS Firewall rule group: %#v", input)
	output, err := conn.CreateFirewallRuleGroup(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall rule group: %w", err)
	}

	d.SetId(aws.StringValue(output.FirewallRuleGroup.Id))

	return resourceAwsRoute53ResolverFirewallRuleGroupRead(d, meta)
}

func resourceAwsRoute53ResolverFirewallRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	ruleGroup, err := finder.FirewallRuleGroupByID(conn, d.Id())

	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Route53 Resolver DNS Firewall rule group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Resolver DNS Firewall rule group (%s): %w", d.Id(), err)
	}

	if ruleGroup == nil {
		log.Printf("[WARN] Route 53 Resolver DNS Firewall rule group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(ruleGroup.Arn)
	d.Set("arn", arn)
	d.Set("id", ruleGroup.Id)
	d.Set("name", ruleGroup.Name)
	d.Set("owner_id", ruleGroup.OwnerId)
	d.Set("share_status", ruleGroup.ShareStatus)

	tags, err := keyvaluetags.Route53resolverListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver DNS Firewall rule group (%s): %w", arn, err)
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

func resourceAwsRoute53ResolverFirewallRuleGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.Route53resolverUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver DNS Firewall rule group (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceAwsRoute53ResolverFirewallRuleGroupRead(d, meta)
}

func resourceAwsRoute53ResolverFirewallRuleGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	_, err := conn.DeleteFirewallRuleGroup(&route53resolver.DeleteFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(d.Id()),
	})

	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNS Firewall rule group (%s): %w", d.Id(), err)
	}

	return nil
}
