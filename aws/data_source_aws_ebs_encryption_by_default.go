package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsEbsEncryptionByDefault() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEbsEncryptionByDefaultRead,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}
func dataSourceAwsEbsEncryptionByDefaultRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	res, err := conn.GetEbsEncryptionByDefault(&ec2.GetEbsEncryptionByDefaultInput{})
	if err != nil {
		return fmt.Errorf("Error reading default EBS encryption toggle: %q", err)
	}

	d.SetId(time.Now().UTC().String())
	d.Set("enabled", res.EbsEncryptionByDefault)

	return nil
}
