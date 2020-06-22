package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsWafXssMatchSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafXssMatchSetCreate,
		Read:   resourceAwsWafXssMatchSetRead,
		Update: resourceAwsWafXssMatchSetUpdate,
		Delete: resourceAwsWafXssMatchSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xss_match_tuples": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_to_match": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											waf.MatchFieldTypeUri,
											waf.MatchFieldTypeSingleQueryArg,
											waf.MatchFieldTypeQueryString,
											waf.MatchFieldTypeMethod,
											waf.MatchFieldTypeHeader,
											waf.MatchFieldTypeBody,
											waf.MatchFieldTypeAllQueryArgs,
										}, false),
									},
								},
							},
						},
						"text_transformation": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								waf.TextTransformationUrlDecode,
								waf.TextTransformationNone,
								waf.TextTransformationHtmlEntityDecode,
								waf.TextTransformationCompressWhiteSpace,
								waf.TextTransformationCmdLine,
								waf.TextTransformationLowercase,
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceAwsWafXssMatchSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn

	log.Printf("[INFO] Creating XssMatchSet: %s", d.Get("name").(string))

	wr := newWafRetryer(conn)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateXssMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}

		return conn.CreateXssMatchSet(params)
	})
	if err != nil {
		return fmt.Errorf("Error creating WAF XSS Match Set: %s", err)
	}
	resp := out.(*waf.CreateXssMatchSetOutput)

	d.SetId(*resp.XssMatchSet.XssMatchSetId)

	if v, ok := d.GetOk("xss_match_tuples"); ok && v.(*schema.Set).Len() > 0 {
		err := updateXssMatchSetResource(d.Id(), nil, v.(*schema.Set).List(), conn)
		if err != nil {
			return fmt.Errorf("Error setting WAF XSS Match Set tuples: %s", err)
		}
	}
	return resourceAwsWafXssMatchSetRead(d, meta)
}

func resourceAwsWafXssMatchSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	log.Printf("[INFO] Reading WAF XSS Match Set: %s", d.Get("name").(string))
	params := &waf.GetXssMatchSetInput{
		XssMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetXssMatchSet(params)
	if err != nil {
		if isAWSErr(err, waf.ErrCodeNonexistentItemException, "") {
			log.Printf("[WARN] WAF XSS Match Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", resp.XssMatchSet.Name)
	d.Set("xss_match_tuples", flattenWafXssMatchTuples(resp.XssMatchSet.XssMatchTuples))

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "waf",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("xssmatchset/%s", d.Id()),
	}
	d.Set("arn", arn.String())

	return nil
}

func resourceAwsWafXssMatchSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn

	if d.HasChange("xss_match_tuples") {
		o, n := d.GetChange("xss_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateXssMatchSetResource(d.Id(), oldT, newT, conn)
		if err != nil {
			return fmt.Errorf("Error updating WAF XSS Match Set: %s", err)
		}
	}

	return resourceAwsWafXssMatchSetRead(d, meta)
}

func resourceAwsWafXssMatchSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn

	oldTuples := d.Get("xss_match_tuples").(*schema.Set).List()
	if len(oldTuples) > 0 {
		err := updateXssMatchSetResource(d.Id(), oldTuples, nil, conn)
		if err != nil {
			return fmt.Errorf("Error removing WAF XSS Match Set tuples: %s", err)
		}
	}

	wr := newWafRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteXssMatchSet(req)
	})
	if err != nil {
		return fmt.Errorf("Error deleting WAF XSS Match Set: %s", err)
	}

	return nil
}

func updateXssMatchSetResource(id string, oldT, newT []interface{}, conn *waf.WAF) error {
	wr := newWafRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(id),
			Updates:       diffWafXssMatchSetTuples(oldT, newT),
		}

		log.Printf("[INFO] Updating WAF XSS Match Set tuples: %s", req)
		return conn.UpdateXssMatchSet(req)
	})
	if err != nil {
		return fmt.Errorf("Error updating WAF XSS Match Set: %s", err)
	}

	return nil
}

func flattenWafXssMatchTuples(ts []*waf.XssMatchTuple) []interface{} {
	out := make([]interface{}, len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m["field_to_match"] = flattenFieldToMatch(t.FieldToMatch)
		m["text_transformation"] = aws.StringValue(t.TextTransformation)
		out[i] = m
	}
	return out
}

func diffWafXssMatchSetTuples(oldT, newT []interface{}) []*waf.XssMatchSetUpdate {
	updates := make([]*waf.XssMatchSetUpdate, 0)

	for _, od := range oldT {
		tuple := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, &waf.XssMatchSetUpdate{
			Action: aws.String(waf.ChangeActionDelete),
			XssMatchTuple: &waf.XssMatchTuple{
				FieldToMatch:       expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: aws.String(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nd := range newT {
		tuple := nd.(map[string]interface{})

		updates = append(updates, &waf.XssMatchSetUpdate{
			Action: aws.String(waf.ChangeActionInsert),
			XssMatchTuple: &waf.XssMatchTuple{
				FieldToMatch:       expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: aws.String(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
