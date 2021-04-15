package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAWSSecurityHubAccount_basic(t *testing.T) {
	resourceName := "aws_securityhub_account.example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubAccountConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubAccountExists(resourceName),
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

func testAccAWSSecurityHubAccount_disappears(t *testing.T) {
	resourceName := "aws_securityhub_account.example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubAccountConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubAccountExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSecurityHubAccount(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSSecurityHubAccount_EnableDefaultStandards(t *testing.T) {
	resourceName := "aws_securityhub_account.example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, securityhub.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubAccountConfig_EnableDefaultStandards(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubAccountExists(resourceName),
				),
			},
			{
				Config: testAccAWSSecurityHubAccountConfig_EnableDefaultStandards(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubAccountExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSecurityHubAccountConfig_EnableDefaultStandards(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubAccountExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckAWSSecurityHubAccountExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		_, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{})

		if err != nil {
			// Can only read enabled standards if Security Hub is enabled
			if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
				return fmt.Errorf("Security Hub account not found")
			}
			return err
		}

		return nil
	}
}

func testAccCheckAWSSecurityHubAccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_account" {
			continue
		}

		_, err := conn.GetEnabledStandards(&securityhub.GetEnabledStandardsInput{})

		if err != nil {
			// Can only read enabled standards if Security Hub is enabled
			if tfawserr.ErrMessageContains(err, securityhub.ErrCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Security Hub account still exists")
	}

	return nil
}

func testAccAWSSecurityHubAccountConfig() string {
	return `
resource "aws_securityhub_account" "example" {}
`
}

func testAccAWSSecurityHubAccountConfig_EnableDefaultStandards(enable bool) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {
  enable_default_standards = %t
}
`, enable)
}
