package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsRoute53ResolverRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsRoute53ResolverRulesRead,

		Schema: map[string]*schema.Schema{
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.Any(
					validateAwsAccountId,
					// The owner of the default Internet Resolver rule.
					validation.StringInSlice([]string{"Route 53 Resolver"}, false),
				),
			},

			"resolver_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"resolver_rule_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"rule_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					route53resolver.RuleTypeOptionForward,
					route53resolver.RuleTypeOptionSystem,
					route53resolver.RuleTypeOptionRecursive,
				}, false),
			},

			"share_status": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					route53resolver.ShareStatusNotShared,
					route53resolver.ShareStatusSharedWithMe,
					route53resolver.ShareStatusSharedByMe,
				}, false),
			},
		},
	}
}

func dataSourceAwsRoute53ResolverRulesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	req := &route53resolver.ListResolverRulesInput{}
	resolverRuleIds := []*string{}

	log.Printf("[DEBUG] Listing Route53 Resolver rules: %s", req)
	err := conn.ListResolverRulesPages(req, func(page *route53resolver.ListResolverRulesOutput, isLast bool) bool {
		for _, rule := range page.ResolverRules {
			if v, ok := d.GetOk("owner_id"); ok && aws.StringValue(rule.OwnerId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("resolver_endpoint_id"); ok && aws.StringValue(rule.ResolverEndpointId) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("rule_type"); ok && aws.StringValue(rule.RuleType) != v.(string) {
				continue
			}
			if v, ok := d.GetOk("share_status"); ok && aws.StringValue(rule.ShareStatus) != v.(string) {
				continue
			}

			resolverRuleIds = append(resolverRuleIds, rule.Id)
		}
		return !isLast
	})
	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver rules: %s", err)
	}

	d.SetId(meta.(*AWSClient).region)

	err = d.Set("resolver_rule_ids", flattenStringSet(resolverRuleIds))
	if err != nil {
		return fmt.Errorf("error setting resolver_rule_ids: %s", err)
	}

	return nil
}
