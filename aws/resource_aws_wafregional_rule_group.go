package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWafRegionalRuleGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalRuleGroupCreate,
		Read:   resourceAwsWafRegionalRuleGroupRead,
		Update: resourceAwsWafRegionalRuleGroupUpdate,
		Delete: resourceAwsWafRegionalRuleGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metric_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateWafMetricName,
			},
			"activated_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"rule_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  wafregional.WafRuleTypeRegular,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsWafRegionalRuleGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WafregionalTags()

	wr := newWafRegionalRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateRuleGroupInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(d.Get("name").(string)),
		}

		if len(tags) > 0 {
			params.Tags = tags
		}

		return conn.CreateRuleGroup(params)
	})
	if err != nil {
		return err
	}
	resp := out.(*waf.CreateRuleGroupOutput)
	d.SetId(*resp.RuleGroup.RuleGroupId)

	activatedRule := d.Get("activated_rule").(*schema.Set).List()
	if len(activatedRule) > 0 {
		noActivatedRules := []interface{}{}

		err := updateWafRuleGroupResourceWR(d.Id(), noActivatedRules, activatedRule, conn, region)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Regional Rule Group: %s", err)
		}
	}

	return resourceAwsWafRegionalRuleGroupRead(d, meta)
}

func resourceAwsWafRegionalRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &waf.GetRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	}

	resp, err := conn.GetRuleGroup(params)
	if err != nil {
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAF Regional Rule Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	rResp, err := conn.ListActivatedRulesInRuleGroup(&waf.ListActivatedRulesInRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error listing activated rules in WAF Regional Rule Group (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("rulegroup/%s", d.Id()),
		Service:   "waf-regional",
	}.String()
	d.Set("arn", arn)

	tags, err := keyvaluetags.WafregionalListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for WAF Regional Rule Group (%s): %s", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("activated_rule", flattenWafActivatedRules(rResp.ActivatedRules))
	d.Set("name", resp.RuleGroup.Name)
	d.Set("metric_name", resp.RuleGroup.MetricName)

	return nil
}

func resourceAwsWafRegionalRuleGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	if d.HasChange("activated_rule") {
		o, n := d.GetChange("activated_rule")
		oldRules, newRules := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateWafRuleGroupResourceWR(d.Id(), oldRules, newRules, conn, region)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Regional Rule Group: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.WafregionalUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWafRegionalRuleGroupRead(d, meta)
}

func resourceAwsWafRegionalRuleGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	oldRules := d.Get("activated_rule").(*schema.Set).List()
	err := deleteWafRegionalRuleGroup(d.Id(), oldRules, conn, region)

	return err
}

func deleteWafRegionalRuleGroup(id string, oldRules []interface{}, conn *wafregional.WAFRegional, region string) error {
	if len(oldRules) > 0 {
		noRules := []interface{}{}
		err := updateWafRuleGroupResourceWR(id, oldRules, noRules, conn, region)
		if err != nil {
			return fmt.Errorf("Error updating WAF Regional Rule Group Predicates: %s", err)
		}
	}

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
		}
		log.Printf("[INFO] Deleting WAF Regional Rule Group")
		return conn.DeleteRuleGroup(req)
	})
	if err != nil {
		return fmt.Errorf("Error deleting WAF Regional Rule Group: %s", err)
	}
	return nil
}

func updateWafRuleGroupResourceWR(id string, oldRules, newRules []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
			Updates:     diffWafRuleGroupActivatedRules(oldRules, newRules),
		}

		return conn.UpdateRuleGroup(req)
	})
	if err != nil {
		return fmt.Errorf("Error Updating WAF Regional Rule Group: %s", err)
	}

	return nil
}
