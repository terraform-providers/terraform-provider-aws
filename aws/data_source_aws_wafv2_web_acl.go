package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsWafv2WebACL() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafv2WebACLRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
		},
	}
}

func dataSourceAwsWafv2WebACLRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	name, nameOk := d.GetOk("name")
	nameRegex, nameRegexOk := d.GetOk("name_regex")

	if !nameOk && !nameRegexOk {
		return fmt.Errorf("One of name or name_regex must be assigned")
	}

	var foundWebACL *wafv2.WebACLSummary
	input := &wafv2.ListWebACLsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListWebACLs(input)
		if err != nil {
			return fmt.Errorf("Error reading WAFv2 WebACLs: %w", err)
		}

		if resp == nil || resp.WebACLs == nil {
			return fmt.Errorf("Error reading WAFv2 WebACLs")
		}

		if nameRegexOk {
			r := regexp.MustCompile(nameRegex.(string))
			for _, webACL := range resp.WebACLs {
				if r.MatchString(aws.StringValue(webACL.Name)) {
					foundWebACL = webACL
					break
				}
			}
		} else {
			for _, webACL := range resp.WebACLs {
				if aws.StringValue(webACL.Name) == name.(string) {
					foundWebACL = webACL
					break
				}
			}
		}

		if resp.NextMarker == nil || foundWebACL != nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundWebACL == nil {
		return fmt.Errorf("WAFv2 WebACL not found for name: %s", name)
	}

	d.SetId(aws.StringValue(foundWebACL.Id))
	d.Set("name", aws.StringValue(foundWebACL.Name))
	d.Set("arn", aws.StringValue(foundWebACL.ARN))
	d.Set("description", aws.StringValue(foundWebACL.Description))

	return nil
}
