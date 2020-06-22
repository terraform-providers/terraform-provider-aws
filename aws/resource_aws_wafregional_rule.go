package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWafRegionalRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalRuleCreate,
		Read:   resourceAwsWafRegionalRuleRead,
		Update: resourceAwsWafRegionalRuleUpdate,
		Delete: resourceAwsWafRegionalRuleDelete,
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"predicate": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"negated": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"data_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateWafPredicatesType(),
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

func resourceAwsWafRegionalRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WafregionalTags()

	wr := newWafRegionalRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateRuleInput{
			ChangeToken: token,
			MetricName:  aws.String(d.Get("metric_name").(string)),
			Name:        aws.String(d.Get("name").(string)),
		}

		if len(tags) > 0 {
			params.Tags = tags
		}

		return conn.CreateRule(params)
	})
	if err != nil {
		return err
	}
	resp := out.(*waf.CreateRuleOutput)
	d.SetId(*resp.Rule.RuleId)

	newPredicates := d.Get("predicate").(*schema.Set).List()
	if len(newPredicates) > 0 {
		noPredicates := []interface{}{}
		err := updateWafRegionalRuleResource(d.Id(), noPredicates, newPredicates, meta)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Regional Rule: %s", err)
		}
	}
	return resourceAwsWafRegionalRuleRead(d, meta)
}

func resourceAwsWafRegionalRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &waf.GetRuleInput{
		RuleId: aws.String(d.Id()),
	}

	resp, err := conn.GetRule(params)
	if err != nil {
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAF Rule (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	arn := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("rule/%s", d.Id()),
		Service:   "waf-regional",
	}.String()
	d.Set("arn", arn)

	tags, err := keyvaluetags.WafregionalListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for WAF Regional Rule (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("predicate", flattenWafPredicates(resp.Rule.Predicates))
	d.Set("name", resp.Rule.Name)
	d.Set("metric_name", resp.Rule.MetricName)

	return nil
}

func resourceAwsWafRegionalRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn

	if d.HasChange("predicate") {
		o, n := d.GetChange("predicate")
		oldP, newP := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateWafRegionalRuleResource(d.Id(), oldP, newP, meta)
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.WafregionalUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWafRegionalRuleRead(d, meta)
}

func resourceAwsWafRegionalRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	oldPredicates := d.Get("predicate").(*schema.Set).List()
	if len(oldPredicates) > 0 {
		noPredicates := []interface{}{}
		err := updateWafRegionalRuleResource(d.Id(), oldPredicates, noPredicates, meta)
		if err != nil {
			return fmt.Errorf("Error Removing WAF Rule Predicates: %s", err)
		}
	}

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Rule")
		return conn.DeleteRule(req)
	})
	if err != nil {
		return fmt.Errorf("Error deleting WAF Rule: %s", err)
	}

	return nil
}

func updateWafRegionalRuleResource(id string, oldP, newP []interface{}, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(id),
			Updates:     diffWafRulePredicates(oldP, newP),
		}

		return conn.UpdateRule(req)
	})

	if err != nil {
		return fmt.Errorf("Error Updating WAF Rule: %s", err)
	}

	return nil
}

func flattenWafPredicates(ts []*waf.Predicate) []interface{} {
	out := make([]interface{}, len(ts))
	for i, p := range ts {
		m := make(map[string]interface{})
		m["negated"] = *p.Negated
		m["type"] = *p.Type
		m["data_id"] = *p.DataId
		out[i] = m
	}
	return out
}
