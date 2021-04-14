package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsWafIpSet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAWSWafIpSetRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAWSWafIpSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafconn
	name := d.Get("name").(string)

	ipsets := make([]*waf.IPSetSummary, 0)
	// ListIPSetsInput does not have a name parameter for filtering or a paginator
	input := &waf.ListIPSetsInput{}
	for {
		output, err := conn.ListIPSets(input)
		if err != nil {
			return fmt.Errorf("Error reading WAF IP sets: %w", err)
		}
		for _, ipset := range output.IPSets {
			if aws.StringValue(ipset.Name) == name {
				ipsets = append(ipsets, ipset)
			}
		}

		if output.NextMarker == nil {
			break
		}
		input.NextMarker = output.NextMarker
	}

	if len(ipsets) == 0 {
		return fmt.Errorf("WAF IP Set not found for name: %s", name)
	}
	if len(ipsets) > 1 {
		return fmt.Errorf("Multiple WAF IP Sets found for name: %s", name)
	}

	ipset := ipsets[0]
	d.SetId(aws.StringValue(ipset.IPSetId))

	return nil
}
