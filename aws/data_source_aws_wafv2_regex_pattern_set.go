package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func dataSourceAwsWafv2RegexPatternSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsWafv2RegexPatternSetRead,

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
				Required: true,
			},
			"regular_expression": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"regex_string": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

func dataSourceAwsWafv2RegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	name := d.Get("name").(string)

	var foundRegexPatternSet *wafv2.RegexPatternSetSummary
	input := &wafv2.ListRegexPatternSetsInput{
		Scope: aws.String(d.Get("scope").(string)),
		Limit: aws.Int64(100),
	}

	for {
		resp, err := conn.ListRegexPatternSets(input)
		if err != nil {
			return fmt.Errorf("Error reading WAFv2 RegexPatternSets: %s", err)
		}

		if resp == nil || resp.RegexPatternSets == nil {
			return fmt.Errorf("Error reading WAFv2 RegexPatternSets")
		}

		for _, regexPatternSet := range resp.RegexPatternSets {
			if regexPatternSet != nil && aws.StringValue(regexPatternSet.Name) == name {
				foundRegexPatternSet = regexPatternSet
				break
			}
		}

		if resp.NextMarker == nil || foundRegexPatternSet != nil {
			break
		}
		input.NextMarker = resp.NextMarker
	}

	if foundRegexPatternSet == nil {
		return fmt.Errorf("WAFv2 RegexPatternSet not found for name: %s", name)
	}

	resp, err := conn.GetRegexPatternSet(&wafv2.GetRegexPatternSetInput{
		Id:    foundRegexPatternSet.Id,
		Name:  foundRegexPatternSet.Name,
		Scope: aws.String(d.Get("scope").(string)),
	})

	if err != nil {
		return fmt.Errorf("Error reading WAFv2 RegexPatternSet: %s", err)
	}

	if resp == nil || resp.RegexPatternSet == nil {
		return fmt.Errorf("Error reading WAFv2 RegexPatternSet")
	}

	d.SetId(aws.StringValue(resp.RegexPatternSet.Id))
	d.Set("arn", aws.StringValue(resp.RegexPatternSet.ARN))
	d.Set("description", aws.StringValue(resp.RegexPatternSet.Description))

	if err := d.Set("regular_expression", flattenWafv2RegexPatternSet(resp.RegexPatternSet.RegularExpressionList)); err != nil {
		return fmt.Errorf("Error setting regular_expression: %s", err)
	}

	return nil
}
