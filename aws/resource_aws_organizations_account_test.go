package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccAwsOrganizationsAccount_basic(t *testing.T) {
	var account organizations.Account

	orgsEmailDomain, ok := os.LookupEnv("TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN")

	if !ok {
		t.Skip("'TEST_AWS_ORGANIZATION_ACCOUNT_EMAIL_DOMAIN' not set, skipping test.")
	}

	rInt := acctest.RandInt()
	name := fmt.Sprintf("tf_acctest_%d", rInt)
	email := fmt.Sprintf("tf-acctest+%d@%s", rInt, orgsEmailDomain)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccOrganizationsAccountPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOrganizationsAccountDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOrganizationsAccountConfig(name, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsOrganizationsAccountExists("aws_organizations_account.test", &account),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "joined_method"),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "joined_timestamp"),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "name", name),
					resource.TestCheckResourceAttr("aws_organizations_account.test", "email", email),
					resource.TestCheckResourceAttrSet("aws_organizations_account.test", "status"),
				),
			},
			{
				ResourceName:      "aws_organizations_account.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsOrganizationsAccountDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).organizationsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_organizations_account" {
			continue
		}

		params := &organizations.DescribeAccountInput{
			AccountId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeAccount(params)

		if err != nil {
			if isAWSErr(err, organizations.ErrCodeAccountNotFoundException, "") {
				return nil
			}
			return err
		}

		if resp == nil && resp.Account != nil {
			return fmt.Errorf("Bad: Account still exists: %q", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsOrganizationsAccountExists(n string, a *organizations.Account) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).organizationsconn
		params := &organizations.DescribeAccountInput{
			AccountId: &rs.Primary.ID,
		}

		resp, err := conn.DescribeAccount(params)

		if err != nil {
			return err
		}

		if resp == nil || resp.Account == nil {
			return fmt.Errorf("Account %q does not exist", rs.Primary.ID)
		}

		a = resp.Account

		return nil
	}
}

func testAccAwsOrganizationsAccountConfig(name, email string) string {
	return fmt.Sprintf(`
resource "aws_organizations_account" "test" {
  name = "%s"
  email = "%s"
}
`, name, email)
}
