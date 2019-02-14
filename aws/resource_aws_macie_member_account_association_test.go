package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMacieMemberAccountAssociation_basic(t *testing.T) {
	key := "MACIE_MEMBER_ACCOUNT_ID"
	memberAcctId := os.Getenv(key)
	if memberAcctId == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMacieMemberAccountAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMacieMemberAccountAssociationConfig_basic(memberAcctId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMacieMemberAccountAssociationExists("aws_macie_member_account_association.test"),
				),
			},
		},
	})
}

func TestAccAWSMacieMemberAccountAssociation_self(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMacieMemberAccountAssociationConfig_self,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMacieMemberAccountAssociationExists("aws_macie_member_account_association.test"),
				),
			},
		},
	})
}

func testAccCheckAWSMacieMemberAccountAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macieconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie_member_account_association" {
			continue
		}

		req := &macie.ListMemberAccountsInput{}

		dissociated := true
		err := conn.ListMemberAccountsPages(req, func(page *macie.ListMemberAccountsOutput, lastPage bool) bool {
			for _, v := range page.MemberAccounts {
				if aws.StringValue(v.AccountId) == rs.Primary.Attributes["member_account_id"] {
					dissociated = false
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !dissociated {
			return fmt.Errorf("Member account %s is not dissociated from Macie", rs.Primary.Attributes["member_account_id"])
		}
	}
	return nil
}

func testAccCheckAWSMacieMemberAccountAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).macieconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		req := &macie.ListMemberAccountsInput{}

		exists := false
		err := conn.ListMemberAccountsPages(req, func(page *macie.ListMemberAccountsOutput, lastPage bool) bool {
			for _, v := range page.MemberAccounts {
				if aws.StringValue(v.AccountId) == rs.Primary.Attributes["member_account_id"] {
					exists = true
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Member account %s is not associated with Macie", rs.Primary.Attributes["member_account_id"])
		}

		return nil
	}
}

func testAccAWSMacieMemberAccountAssociationConfig_basic(memberAcctId string) string {
	return fmt.Sprintf(`
resource "aws_macie_member_account_association" "test" {
  member_account_id = "%s"
}
`, memberAcctId)
}

const testAccAWSMacieMemberAccountAssociationConfig_self = `
data "aws_caller_identity" "current" {}

resource "aws_macie_member_account_association" "test" {
  member_account_id = "${data.aws_caller_identity.current.account_id}"
}
`
