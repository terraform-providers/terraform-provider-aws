package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/wafregional/finder"
)

func resourceAwsWafRegionalRegexMatchSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalRegexMatchSetCreate,
		Read:   resourceAwsWafRegionalRegexMatchSetRead,
		Update: resourceAwsWafRegionalRegexMatchSetUpdate,
		Delete: resourceAwsWafRegionalRegexMatchSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"regex_match_tuple": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourceAwsWafRegexMatchSetTupleHash,
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
										StateFunc: func(v interface{}) string {
											return strings.ToLower(v.(string))
										},
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"regex_pattern_set_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"text_transformation": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsWafRegionalRegexMatchSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	log.Printf("[INFO] Creating WAF Regional Regex Match Set: %s", d.Get("name").(string))

	wr := newWafRegionalRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateRegexMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}
		return conn.CreateRegexMatchSet(params)
	})
	if err != nil {
		return fmt.Errorf("Failed creating WAF Regional Regex Match Set: %s", err)
	}
	resp := out.(*waf.CreateRegexMatchSetOutput)

	d.SetId(aws.StringValue(resp.RegexMatchSet.RegexMatchSetId))

	return resourceAwsWafRegionalRegexMatchSetUpdate(d, meta)
}

func resourceAwsWafRegionalRegexMatchSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn

	log.Printf("[INFO] Reading WAF Regional Regex Match Set: %s", d.Get("name").(string))
	set, err := finder.RegexMatchSetByID(conn, d.Id())
	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		log.Printf("[WARN] WAF Regional Regex Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting WAF Regional Regex Match Set (%s): %s", d.Id(), err)
	}

	d.Set("name", set.Name)
	d.Set("regex_match_tuple", flattenWafRegexMatchTuples(set.RegexMatchTuples))

	return nil
}

func resourceAwsWafRegionalRegexMatchSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	log.Printf("[INFO] Updating WAF Regional Regex Match Set: %s", d.Get("name").(string))

	if d.HasChange("regex_match_tuple") {
		o, n := d.GetChange("regex_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		err := updateRegexMatchSetResourceWR(d.Id(), oldT, newT, conn, region)
		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			log.Printf("[WARN] WAF Regional Rate Based Rule (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regional Regex Match Set(%s): %s", d.Id(), err)
		}
	}

	return resourceAwsWafRegionalRegexMatchSetRead(d, meta)
}

func resourceAwsWafRegionalRegexMatchSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	err := deleteRegexMatchSetResource(conn, region, "global", d.Id(), getRegexMatchTuplesFromResourceData(d))
	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		log.Printf("[WARN] WAF Regional Regex Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

func getRegexMatchTuplesFromResourceData(d *schema.ResourceData) []*waf.RegexMatchTuple {
	result := []*waf.RegexMatchTuple{}

	for _, t := range d.Get("regex_match_tuple").(*schema.Set).List() {
		result = append(result, expandWafRegexMatchTuple(t.(map[string]interface{})))
	}

	return result
}

func getRegexMatchTuplesFromAPIResource(r *waf.RegexMatchSet) []*waf.RegexMatchTuple {
	return r.RegexMatchTuples
}

func clearRegexMatchTuples(conn *wafregional.WAFRegional, region string, id string, tuples []*waf.RegexMatchTuple) error {
	if len(tuples) > 0 {
		input := &waf.UpdateRegexMatchSetInput{
			RegexMatchSetId: aws.String(id),
		}
		for _, tuple := range tuples {
			input.Updates = append(input.Updates, &waf.RegexMatchSetUpdate{
				Action:          aws.String(waf.ChangeActionDelete),
				RegexMatchTuple: tuple,
			})
		}

		log.Printf("[INFO] Clearing WAF Regional Regex Match Set: %s", id)
		wr := newWafRegionalRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			input.ChangeToken = token
			return conn.UpdateRegexMatchSet(input)
		})
		if err != nil {
			return fmt.Errorf("error clearing WAF Regional Regex Match Set (%s): %w", id, err)
		}
	}
	return nil
}

func deleteRegexMatchSet(conn *wafregional.WAFRegional, region, id string) error {
	log.Printf("[INFO] Deleting WAF Regional Regex Match Set: %s", id)
	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(id),
		}
		return conn.DeleteRegexMatchSet(req)
	})
	if err != nil {
		return fmt.Errorf("error deleting WAF Regional Regex Match Set (%s): %w", id, err)
	}
	return nil
}

func deleteRegexMatchSetResource(conn *wafregional.WAFRegional, region, tokenRegion, id string, tuples []*waf.RegexMatchTuple) error {
	err := clearRegexMatchTuples(conn, region, id, tuples)
	if err != nil {
		return err
	}

	return deleteRegexMatchSet(conn, tokenRegion, id)
}

func updateRegexMatchSetResourceWR(id string, oldT, newT []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(id),
			Updates:         diffWafRegexMatchSetTuples(oldT, newT),
		}

		return conn.UpdateRegexMatchSet(req)
	})

	return err
}
