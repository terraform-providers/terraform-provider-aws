package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	WafRuleDeleteTimeout = 5 * time.Minute
)

func resourceAwsWafRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRuleCreate,
		Read:   resourceAwsWafRuleRead,
		Update: resourceAwsWafRuleUpdate,
		Delete: resourceAwsWafRuleDelete,
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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z]+$`), "must contain only alphanumeric characters"),
			},
			"predicates": {
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
							ValidateFunc: validation.StringLenBetween(0, 128),
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(waf.PredicateType_Values(), false),
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

func resourceAwsWafRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().WafTags()

	wr := newWafRetryer(conn)
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
		return fmt.Errorf("error creating WAF Rule (%s): %w", d.Get("name").(string), err)
	}

	resp := out.(*waf.CreateRuleOutput)
	d.SetId(aws.StringValue(resp.Rule.RuleId))

	newPredicates := d.Get("predicates").(*schema.Set).List()
	if len(newPredicates) > 0 {
		noPredicates := []interface{}{}
		err := updateWafRuleResource(d.Id(), noPredicates, newPredicates, conn)
		if err != nil {
			return fmt.Errorf("error updating WAF Rule (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsWafRuleRead(d, meta)
}

func resourceAwsWafRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &waf.GetRuleInput{
		RuleId: aws.String(d.Id()),
	}

	resp, err := conn.GetRule(params)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
		log.Printf("[WARN] WAF Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading WAF Rule (%s): %w", d.Id(), err)
	}

	if resp == nil || resp.Rule == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading WAF Rule (%s): not found", d.Id())
		}

		log.Printf("[WARN] WAF Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var predicates []map[string]interface{}

	for _, predicateSet := range resp.Rule.Predicates {
		predicate := map[string]interface{}{
			"negated": *predicateSet.Negated,
			"type":    *predicateSet.Type,
			"data_id": *predicateSet.DataId,
		}
		predicates = append(predicates, predicate)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "waf",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("rule/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	tags, err := keyvaluetags.WafListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for WAF Rule (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("predicates", predicates)
	d.Set("name", resp.Rule.Name)
	d.Set("metric_name", resp.Rule.MetricName)

	return nil
}

func resourceAwsWafRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn

	if d.HasChange("predicates") {
		o, n := d.GetChange("predicates")
		oldP, newP := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateWafRuleResource(d.Id(), oldP, newP, conn)
		if err != nil {
			return fmt.Errorf("error updating WAF Rule (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.WafUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating WAF Rule (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsWafRuleRead(d, meta)
}

func resourceAwsWafRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn

	oldPredicates := d.Get("predicates").(*schema.Set).List()
	if len(oldPredicates) > 0 {
		noPredicates := []interface{}{}
		err := updateWafRuleResource(d.Id(), oldPredicates, noPredicates, conn)
		if err != nil {
			return fmt.Errorf("error updating WAF Rule (%s) predicates: %w", d.Id(), err)
		}
	}

	wr := newWafRetryer(conn)
	err := resource.Retry(WafRuleDeleteTimeout, func() *resource.RetryError {
		var err error
		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.DeleteRuleInput{
				ChangeToken: token,
				RuleId:      aws.String(d.Id()),
			}

			return conn.DeleteRule(req)
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, waf.ErrCodeReferencedItemException) {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.DeleteRuleInput{
				ChangeToken: token,
				RuleId:      aws.String(d.Id()),
			}

			return conn.DeleteRule(req)
		})
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			return nil
		}
		return fmt.Errorf("error deleting WAF Rule (%s): %w", d.Id(), err)
	}

	return nil
}

func updateWafRuleResource(id string, oldP, newP []interface{}, conn *waf.WAF) error {
	wr := newWafRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRuleInput{
			ChangeToken: token,
			RuleId:      aws.String(id),
			Updates:     diffWafRulePredicates(oldP, newP),
		}

		return conn.UpdateRule(req)
	})

	return err
}
