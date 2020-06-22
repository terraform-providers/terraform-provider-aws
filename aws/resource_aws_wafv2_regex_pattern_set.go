package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWafv2RegexPatternSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2RegexPatternSetCreate,
		Read:   resourceAwsWafv2RegexPatternSetRead,
		Update: resourceAwsWafv2RegexPatternSetUpdate,
		Delete: resourceAwsWafv2RegexPatternSetDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"lock_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"regular_expression": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"regex_string": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 200),
						},
					},
				},
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsWafv2RegexPatternSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	params := &wafv2.CreateRegexPatternSetInput{
		Name:                  aws.String(d.Get("name").(string)),
		Scope:                 aws.String(d.Get("scope").(string)),
		RegularExpressionList: []*wafv2.Regex{},
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
		params.RegularExpressionList = expandWafv2RegexPatternSet(v.(*schema.Set).List())
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().Wafv2Tags()
	}

	resp, err := conn.CreateRegexPatternSet(params)

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 RegexPatternSet: %s", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 RegexPatternSet")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceAwsWafv2RegexPatternSetRead(d, meta)
}

func resourceAwsWafv2RegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	params := &wafv2.GetRegexPatternSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetRegexPatternSet(params)
	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 RegexPatternSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.RegexPatternSet == nil {
		return fmt.Errorf("Error getting WAFv2 RegexPatternSet")
	}

	d.Set("name", aws.StringValue(resp.RegexPatternSet.Name))
	d.Set("description", aws.StringValue(resp.RegexPatternSet.Description))
	d.Set("arn", aws.StringValue(resp.RegexPatternSet.ARN))
	d.Set("lock_token", aws.StringValue(resp.LockToken))

	if err := d.Set("regular_expression", flattenWafv2RegexPatternSet(resp.RegexPatternSet.RegularExpressionList)); err != nil {
		return fmt.Errorf("Error setting regular_expression: %s", err)
	}

	tags, err := keyvaluetags.Wafv2ListTags(conn, aws.StringValue(resp.RegexPatternSet.ARN))
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 RegexPatternSet (%s): %s", aws.StringValue(resp.RegexPatternSet.ARN), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWafv2RegexPatternSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Updating WAFv2 RegexPatternSet %s", d.Id())

	params := &wafv2.UpdateRegexPatternSetInput{
		Id:                    aws.String(d.Id()),
		Name:                  aws.String(d.Get("name").(string)),
		Scope:                 aws.String(d.Get("scope").(string)),
		LockToken:             aws.String(d.Get("lock_token").(string)),
		RegularExpressionList: []*wafv2.Regex{},
	}

	if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
		params.RegularExpressionList = expandWafv2RegexPatternSet(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateRegexPatternSet(params)

	if err != nil {
		return fmt.Errorf("Error updating WAFv2 RegexPatternSet: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating tags: %s", err)
		}
	}

	return resourceAwsWafv2RegexPatternSetRead(d, meta)
}

func resourceAwsWafv2RegexPatternSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Deleting WAFv2 RegexPatternSet %s", d.Id())
	params := &wafv2.DeleteRegexPatternSetInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteRegexPatternSet(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFAssociatedItemException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRegexPatternSet(params)
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 RegexPatternSet: %s", err)
	}

	return nil
}

func expandWafv2RegexPatternSet(l []interface{}) []*wafv2.Regex {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	regexPatterns := make([]*wafv2.Regex, 0)
	for _, regexPattern := range l {
		if regexPattern == nil {
			continue
		}
		regexPatterns = append(regexPatterns, expandWafv2Regex(regexPattern.(map[string]interface{})))
	}

	return regexPatterns
}

func expandWafv2Regex(m map[string]interface{}) *wafv2.Regex {
	if m == nil {
		return nil
	}

	return &wafv2.Regex{
		RegexString: aws.String(m["regex_string"].(string)),
	}
}

func flattenWafv2RegexPatternSet(r []*wafv2.Regex) interface{} {
	if r == nil {
		return []interface{}{}
	}

	regexPatterns := make([]interface{}, 0)

	for _, regexPattern := range r {
		if regexPattern == nil {
			continue
		}
		d := map[string]interface{}{
			"regex_string": aws.StringValue(regexPattern.RegexString),
		}
		regexPatterns = append(regexPatterns, d)
	}

	return regexPatterns
}
