package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsSecurityHubOrganizationConfiguration_basic(t *testing.T) {
	resourceName := "aws_securityhub_organization_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsSecurityHubOrganizationConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubOrganizationConfigurationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccAwsSecurityHubOrganizationConfigurationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAwsSecurityHubOrganizationConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		_, err := conn.DescribeOrganizationConfiguration(&securityhub.DescribeOrganizationConfigurationInput{})

		if err != nil {
			// Can only call DescribeOrganizationConfiguration on the Organization Admin Account
			if isAWSErr(err, "InvalidAccessException", fmt.Sprintf("Account %s is not an administrator for this organization", testAccAwsProviderAccountID(testAccProvider))) {
				return nil
			}
			return err
		}

		return err
	}
}

func testAccAwsSecurityHubOrganizationConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_organization_configuration" {
			continue
		}

		_, err := conn.DescribeOrganizationConfiguration(&securityhub.DescribeOrganizationConfigurationInput{})

		if err != nil {
			// Can only call DescribeOrganizationConfiguration on the Organization Admin Account
			if isAWSErr(err, "InvalidAccessException", fmt.Sprintf("Account %s is not an administrator for this organization", testAccAwsProviderAccountID(testAccProvider))) {
				return nil
			}
			return err
		}

		return fmt.Errorf("SecurityHub AutoEnable still enabled")
	}

	return nil
}

func testAccAwsSecurityHubOrganizationConfigurationConfig() string {
	return `
resource "aws_securityhub_organization_configuration" "test" {}
`
}
