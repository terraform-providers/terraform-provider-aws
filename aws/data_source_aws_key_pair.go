package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsKeyPair() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKeyPairRead,
		Schema: map[string]*schema.Schema{
			"key_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": dataSourceFiltersSchema(),
		},
	}
}

func dataSourceAwsKeyPairRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DescribeKeyPairsInput{}
	filters, filtersOk := d.GetOk("filter")
	if filtersOk {
		params.Filters = buildAwsDataSourceFilters(filters.(*schema.Set))
	}

	keyName := d.Get("key_name").(string)

	params.KeyNames = []*string{
		aws.String(keyName),
	}

	log.Printf("[DEBUG] Reading key pair: %s", keyName)
	resp, err := conn.DescribeKeyPairs(params)
	if err != nil {
		return fmt.Errorf("error describing EC2 Key Pairs: %w", err)
	}

	if resp == nil || len(resp.KeyPairs) == 0 {
		return errors.New("no matching Key Pair found")
	}

	filteredKeyPair := resp.KeyPairs

	if len(filteredKeyPair) > 1 {
		return fmt.Errorf("Your query returned more than one result. Please try a more " +
			"specific search criteria")
	}

	keyPair := filteredKeyPair[0]
	log.Printf("[DEBUG] aws_key_pair - Single key pair found: %s", *keyPair.KeyName)

	d.Set("fingerprint", keyPair.KeyFingerprint)
	d.SetId(aws.StringValue(keyPair.KeyPairId))

	return nil
}
