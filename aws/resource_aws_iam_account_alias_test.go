package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSIAMAccountAlias_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"DataSource": {
			"basic": testAccAWSIAMAccountAliasDataSource_basic,
		},
		"Resource": {
			"basic": testAccAWSIAMAccountAlias_basic,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAWSIAMAccountAlias_basic(t *testing.T) {
	resourceName := "aws_iam_account_alias.test"

	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, iam.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMAccountAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMAccountAliasConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSIAMAccountAliasExists(resourceName),
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

func testAccCheckAWSIAMAccountAliasDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_account_alias" {
			continue
		}

		params := &iam.ListAccountAliasesInput{}

		resp, err := conn.ListAccountAliases(params)

		if err != nil {
			return fmt.Errorf("error reading IAM Account Alias (%s): %w", rs.Primary.ID, err)
		}

		if resp == nil {
			return fmt.Errorf("error reading IAM Account Alias (%s): empty response", rs.Primary.ID)
		}

		if len(resp.AccountAliases) > 0 {
			return fmt.Errorf("Bad: Account alias still exists: %q", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAWSIAMAccountAliasExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		params := &iam.ListAccountAliasesInput{}

		resp, err := conn.ListAccountAliases(params)

		if err != nil {
			return fmt.Errorf("error reading IAM Account Alias (%s): %w", rs.Primary.ID, err)
		}

		if resp == nil {
			return fmt.Errorf("error reading IAM Account Alias (%s): empty response", rs.Primary.ID)
		}

		if len(resp.AccountAliases) == 0 {
			return fmt.Errorf("Bad: Account alias %q does not exist", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSIAMAccountAliasConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_account_alias" "test" {
  account_alias = %[1]q
}
`, rName)
}
