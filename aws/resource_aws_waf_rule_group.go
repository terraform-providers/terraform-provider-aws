package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWafRuleGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRuleGroupCreate,
		Read:   resourceAwsWafRuleGroupRead,
		Update: resourceAwsWafRuleGroupUpdate,
		Delete: resourceAwsWafRuleGroupDelete,
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
							Default:  waf.WafRuleTypeRegular,
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

func resourceAwsWafRuleGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WafTags()

	wr := newWafRetryer(conn)
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

	activatedRules := d.Get("activated_rule").(*schema.Set).List()
	if len(activatedRules) > 0 {
		noActivatedRules := []interface{}{}

		err := updateWafRuleGroupResource(d.Id(), noActivatedRules, activatedRules, conn)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule Group: %s", err)
		}
	}

	return resourceAwsWafRuleGroupRead(d, meta)
}

func resourceAwsWafRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &waf.GetRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	}

	resp, err := conn.GetRuleGroup(params)
	if err != nil {
		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			log.Printf("[WARN] WAF Rule Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	rResp, err := conn.ListActivatedRulesInRuleGroup(&waf.ListActivatedRulesInRuleGroupInput{
		RuleGroupId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error listing activated rules in WAF Rule Group (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "waf",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("rulegroup/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	tags, err := keyvaluetags.WafListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for WAF Rule Group (%s): %s", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("activated_rule", flattenWafActivatedRules(rResp.ActivatedRules))
	d.Set("name", resp.RuleGroup.Name)
	d.Set("metric_name", resp.RuleGroup.MetricName)

	return nil
}

func resourceAwsWafRuleGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn

	if d.HasChange("activated_rule") {
		o, n := d.GetChange("activated_rule")
		oldRules, newRules := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateWafRuleGroupResource(d.Id(), oldRules, newRules, conn)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule Group: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.WafUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWafRuleGroupRead(d, meta)
}

func resourceAwsWafRuleGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn

	oldRules := d.Get("activated_rule").(*schema.Set).List()
	err := deleteWafRuleGroup(d.Id(), oldRules, conn)

	return err
}

func deleteWafRuleGroup(id string, oldRules []interface{}, conn *waf.WAF) error {
	if len(oldRules) > 0 {
		noRules := []interface{}{}
		err := updateWafRuleGroupResource(id, oldRules, noRules, conn)
		if err != nil {
			return fmt.Errorf("Error updating WAF Rule Group Predicates: %s", err)
		}
	}

	wr := newWafRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
		}
		log.Printf("[INFO] Deleting WAF Rule Group")
		return conn.DeleteRuleGroup(req)
	})
	if err != nil {
		return fmt.Errorf("Error deleting WAF Rule Group: %s", err)
	}
	return nil
}

func updateWafRuleGroupResource(id string, oldRules, newRules []interface{}, conn *waf.WAF) error {
	wr := newWafRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRuleGroupInput{
			ChangeToken: token,
			RuleGroupId: aws.String(id),
			Updates:     diffWafRuleGroupActivatedRules(oldRules, newRules),
		}

		return conn.UpdateRuleGroup(req)
	})
	if err != nil {
		return fmt.Errorf("Error Updating WAF Rule Group: %s", err)
	}

	return nil
}
