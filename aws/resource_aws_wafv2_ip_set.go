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

func resourceAwsWafv2IPSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2IPSetCreate,
		Read:   resourceAwsWafv2IPSetRead,
		Update: resourceAwsWafv2IPSetUpdate,
		Delete: resourceAwsWafv2IPSetDelete,
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
			"addresses": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10000,
				Elem:     &schema.Schema{Type: schema.TypeString},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					o, n := d.GetChange("addresses")
					oldAddresses := o.(*schema.Set).List()
					newAddresses := n.(*schema.Set).List()
					if len(oldAddresses) == len(newAddresses) {
						for _, ov := range oldAddresses {
							hasAddress := false
							for _, nv := range newAddresses {
								// isIpv6CidrsEquals works for both IPv4 and IPv6
								if isIpv6CidrsEquals(ov.(string), nv.(string)) {
									hasAddress = true
									break
								}
							}
							if !hasAddress {
								return false
							}
						}
						return true
					}
					return false
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"ip_address_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.IPAddressVersionIpv4,
					wafv2.IPAddressVersionIpv6,
				}, false),
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

func resourceAwsWafv2IPSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	params := &wafv2.CreateIPSetInput{
		Addresses:        aws.StringSlice([]string{}),
		IPAddressVersion: aws.String(d.Get("ip_address_version").(string)),
		Name:             aws.String(d.Get("name").(string)),
		Scope:            aws.String(d.Get("scope").(string)),
	}

	if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
		params.Addresses = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().Wafv2Tags()
	}

	resp, err := conn.CreateIPSet(params)

	if err != nil {
		return fmt.Errorf("Error creating WAFv2 IPSet: %s", err)
	}

	if resp == nil || resp.Summary == nil {
		return fmt.Errorf("Error creating WAFv2 IPSet")
	}

	d.SetId(aws.StringValue(resp.Summary.Id))

	return resourceAwsWafv2IPSetRead(d, meta)
}

func resourceAwsWafv2IPSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	params := &wafv2.GetIPSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetIPSet(params)
	if err != nil {
		if isAWSErr(err, wafv2.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAFv2 IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if resp == nil || resp.IPSet == nil {
		return fmt.Errorf("Error reading WAFv2 IPSet")
	}

	d.Set("name", aws.StringValue(resp.IPSet.Name))
	d.Set("description", aws.StringValue(resp.IPSet.Description))
	d.Set("ip_address_version", aws.StringValue(resp.IPSet.IPAddressVersion))
	d.Set("arn", aws.StringValue(resp.IPSet.ARN))
	d.Set("lock_token", aws.StringValue(resp.LockToken))

	if err := d.Set("addresses", flattenStringSet(resp.IPSet.Addresses)); err != nil {
		return fmt.Errorf("Error setting addresses: %s", err)
	}

	arn := aws.StringValue(resp.IPSet.ARN)
	tags, err := keyvaluetags.Wafv2ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("Error listing tags for WAFv2 IpSet (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("Error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWafv2IPSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Updating WAFv2 IPSet %s", d.Id())

	params := &wafv2.UpdateIPSetInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		Addresses: aws.StringSlice([]string{}),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
		params.Addresses = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok && len(v.(string)) > 0 {
		params.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateIPSet(params)

	if err != nil {
		return fmt.Errorf("Error updating WAFv2 IPSet: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("Error updating tags: %s", err)
		}
	}

	return resourceAwsWafv2IPSetRead(d, meta)
}

func resourceAwsWafv2IPSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Deleting WAFv2 IPSet %s", d.Id())

	params := &wafv2.DeleteIPSetInput{
		Id:        aws.String(d.Id()),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
		LockToken: aws.String(d.Get("lock_token").(string)),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		_, err = conn.DeleteIPSet(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFAssociatedItemException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteIPSet(params)
	}

	if err != nil {
		return fmt.Errorf("Error deleting WAFv2 IPSet: %s", err)
	}

	return nil
}
