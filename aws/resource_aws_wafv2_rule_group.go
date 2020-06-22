package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsWafv2RuleGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2RuleGroupCreate,
		Read:   resourceAwsWafv2RuleGroupRead,
		Update: resourceAwsWafv2RuleGroupUpdate,
		Delete: resourceAwsWafv2RuleGroupDelete,
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
			"capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-_]+$`), "must contain only alphanumeric hyphen and underscore characters"),
				),
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
			"rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allow": wafv2EmptySchema(),
									"block": wafv2EmptySchema(),
									"count": wafv2EmptySchema(),
								},
							},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
						"priority": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"statement":         wafv2RootStatementSchema(3),
						"visibility_config": wafv2VisibilityConfigSchema(),
					},
				},
			},
			"tags":              tagsSchema(),
			"visibility_config": wafv2VisibilityConfigSchema(),
		},
	}
}

func resourceAwsWafv2RuleGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.CreateRuleGroupOutput

	params := &wafv2.CreateRuleGroupInput{
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
		Capacity:         aws.Int64(int64(d.Get("capacity").(int))),
		Rules:            expandWafv2Rules(d.Get("rule").(*schema.Set).List()),
		VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().Wafv2Tags()
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateRuleGroup(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateRuleGroup(params)
	}

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 RuleGroup: %s", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 RuleGroup")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceAwsWafv2RuleGroupRead(d, meta)
}

func resourceAwsWafv2RuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &wafv2.GetRuleGroupInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetRuleGroup(params)
	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 RuleGroup (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.RuleGroup == nil {
		return fmt.Errorf("Error getting WAFv2 RuleGroup")
	}

	d.Set("name", aws.StringValue(resp.RuleGroup.Name))
	d.Set("capacity", aws.Int64Value(resp.RuleGroup.Capacity))
	d.Set("description", aws.StringValue(resp.RuleGroup.Description))
	d.Set("arn", aws.StringValue(resp.RuleGroup.ARN))
	d.Set("lock_token", aws.StringValue(resp.LockToken))

	if err := d.Set("rule", flattenWafv2Rules(resp.RuleGroup.Rules)); err != nil {
		return fmt.Errorf("Error setting rule: %s", err)
	}

	if err := d.Set("visibility_config", flattenWafv2VisibilityConfig(resp.RuleGroup.VisibilityConfig)); err != nil {
		return fmt.Errorf("Error setting visibility_config: %s", err)
	}

	arn := aws.StringValue(resp.RuleGroup.ARN)
	tags, err := keyvaluetags.Wafv2ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 RuleGroup (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWafv2RuleGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Updating WAFv2 RuleGroup %s", d.Id())

	u := &wafv2.UpdateRuleGroupInput{
		Id:               aws.String(d.Id()),
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
		LockToken:        aws.String(d.Get("lock_token").(string)),
		Rules:            expandWafv2Rules(d.Get("rule").(*schema.Set).List()),
		VisibilityConfig: expandWafv2VisibilityConfig(d.Get("visibility_config").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		u.Description = aws.String(v.(string))
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.UpdateRuleGroup(u)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.UpdateRuleGroup(u)
	}

	if err != nil {
		return fmt.Errorf("Error updating WAFv2 RuleGroup: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating tags: %s", err)
		}
	}

	return resourceAwsWafv2RuleGroupRead(d, meta)
}

func resourceAwsWafv2RuleGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Deleting WAFv2 RuleGroup %s", d.Id())

	r := &wafv2.DeleteRuleGroupInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteRuleGroup(r)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFAssociatedItemException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFUnavailableEntityException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRuleGroup(r)
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 RuleGroup: %s", err)
	}

	return nil
}
