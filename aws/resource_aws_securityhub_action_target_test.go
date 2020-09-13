package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsSecurityHubActionTarget_basic(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubActionTargetConfigIdentifier("testaction"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubActionTargetExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "securityhub", "action/custom/testaction"),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test custom action"),
					resource.TestCheckResourceAttr(resourceName, "identifier", "testaction"),
					resource.TestCheckResourceAttr(resourceName, "name", "Test action"),
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

func testAccAwsSecurityHubActionTarget_disappears(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubActionTargetConfigIdentifier("testaction"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubActionTargetExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSecurityHubActionTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsSecurityHubActionTarget_Description(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubActionTargetConfigDescription("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsSecurityHubActionTargetConfigDescription("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccAwsSecurityHubActionTarget_Name(t *testing.T) {
	resourceName := "aws_securityhub_action_target.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsSecurityHubActionTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSecurityHubActionTargetConfigName("name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsSecurityHubActionTargetConfigName("name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsSecurityHubActionTargetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "name2"),
				),
			},
		},
	})
}

func testAccCheckAwsSecurityHubActionTargetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Security Hub custom action ARN is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		action, err := resourceAwsSecurityHubActionTargetCheckExists(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if action == nil {
			return fmt.Errorf("Security Hub custom action %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAwsSecurityHubActionTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_action_target" {
			continue
		}

		action, err := resourceAwsSecurityHubActionTargetCheckExists(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if action != nil {
			return fmt.Errorf("Security Hub custom action %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsSecurityHubActionTargetConfigDescription(description string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = %[1]q
  identifier  = "testaction"
  name        = "Test action"
}
`, description)
}

func testAccAwsSecurityHubActionTargetConfigIdentifier(identifier string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = "This is a test custom action"
  identifier  = %[1]q
  name        = "Test action"
}
`, identifier)
}

func testAccAwsSecurityHubActionTargetConfigName(name string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "test" {}

resource "aws_securityhub_action_target" "test" {
  depends_on  = [aws_securityhub_account.test]
  description = "This is a test custom action"
  identifier  = "testaction"
  name        = %[1]q
}
`, name)
}
