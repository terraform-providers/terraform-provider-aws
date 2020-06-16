package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsIotGreengrassServiceRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIotGreengrassServiceRoleRead,
		Schema: map[string]*schema.Schema{
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIotGreengrassServiceRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn
	input := &greengrass.GetServiceRoleForAccountInput{}

	output, err := conn.GetServiceRoleForAccount(input)

	if err != nil {
		return fmt.Errorf("error while getting greengrass service role for account: %s", err)
	}

	roleArn := aws.StringValue(output.RoleArn)

	if err := d.Set("role_arn", roleArn); err != nil {
		return fmt.Errorf("error setting role_arn: %s", err)
	}

	d.SetId(roleArn)

	return nil
}
