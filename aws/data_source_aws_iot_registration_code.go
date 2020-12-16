package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsIotRegistrationCode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIotRegistrationCodeRead,
		Schema: map[string]*schema.Schema{
			"code": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIotRegistrationCodeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	input := &iot.GetRegistrationCodeInput{}

	output, err := conn.GetRegistrationCode(input)
	if err != nil {
		return fmt.Errorf("error reading registration code: %v", err)
	}
	code := aws.StringValue(output.RegistrationCode)
	d.SetId(code)
	if err := d.Set("code", code); err != nil {
		return fmt.Errorf("error setting code: %s", err)
	}
	return nil
}
