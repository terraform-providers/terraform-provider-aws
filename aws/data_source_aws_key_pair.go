package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsKeyPair() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsKeyPair,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"fingerprint": {
				Type:      schema.TypeString,
				Required: false,
			}
		},
	}
}

func dataSourceAwsKeyPair(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	params := &ec2.DescribeKeyPairs{}

	keyName := d.Get("name").(string)
	params.KeyName = []*string(aws.String(keyname))
	log.Printf("[DEBUG] Reading key name: %s", keyName)
	resp, err := conn.DescribeKeyPairs(params)
	if err != nil {
		return err
	}
	log.Printf("Resp: %s", resp)
	v := KeyPairInfo[0]
	d.set("fingerprint", v.keyFingerprint)

	return nil
}
