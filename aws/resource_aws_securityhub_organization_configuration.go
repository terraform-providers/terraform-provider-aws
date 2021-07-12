package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSecurityHubOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubOrganizationConfigurationCreate,
		Read:   resourceAwsSecurityHubOrganizationConfigurationRead,
		Delete: resourceAwsSecurityHubOrganizationConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{},
	}
}

func resourceAwsSecurityHubOrganizationConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	input := &securityhub.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(true),
	}

	_, err := conn.UpdateOrganizationConfiguration(input)

	if err != nil {
		return fmt.Errorf("error updating Security Hub Organization Configuration (%s): %w", d.Id(), err)
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceAwsSecurityHubOrganizationConfigurationRead(d, meta)
}

func resourceAwsSecurityHubOrganizationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	_, err := conn.DescribeOrganizationConfiguration(&securityhub.DescribeOrganizationConfigurationInput{})

	if err != nil {
		return fmt.Errorf("error checking Security Hub Organization Configuration: %s", err)
	}

	return nil
}

func resourceAwsSecurityHubOrganizationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	input := &securityhub.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(false),
	}

	_, err := conn.UpdateOrganizationConfiguration(input)

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disabling Security Hub Organization Configuration auto enable (%s): %w", d.Id(), err)
	}

	return nil
}
