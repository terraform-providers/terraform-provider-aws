package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/envvar"
)

func testAccAwsMacie2InvitationAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_macie2_invitation_accepter.member"
	email := envvar.TestSkipIfEmpty(t, EnvVarMacie2PrincipalEmail, EnvVarMacie2PrincipalEmailMessageError)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsMacie2InvitationAccepterDestroy,
		ErrorCheck:        testAccErrorCheck(t, macie2.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMacieInvitationAccepterConfigBasic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMacie2InvitationAccepterExists(resourceName),
				),
			},
			{
				Config:            testAccAwsMacieInvitationAccepterConfigBasic(email),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMacie2InvitationAccepterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource (%s) has empty ID", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).macie2conn
		input := &macie2.GetAdministratorAccountInput{}
		output, err := conn.GetAdministratorAccount(input)

		if err != nil {
			return err
		}

		if output == nil || output.Administrator == nil || aws.StringValue(output.Administrator.AccountId) == "" {
			return fmt.Errorf("no administrator account found for: %s", resourceName)
		}

		return nil
	}
}

func testAccCheckAwsMacie2InvitationAccepterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macie2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie2_invitation_accepter" {
			continue
		}

		input := &macie2.GetAdministratorAccountInput{}
		output, err := conn.GetAdministratorAccount(input)
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			continue
		}

		if output == nil || output.Administrator == nil || aws.StringValue(output.Administrator.AccountId) != rs.Primary.Attributes["administrator_account_id"] {
			continue
		}

		return fmt.Errorf("macie InvitationAccepter %q still exists", rs.Primary.ID)

	}

	return nil

}

func testAccAwsMacieInvitationAccepterConfigBasic(email string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
data "aws_caller_identity" "admin" {
  provider = "awsalternate"
}

data "aws_caller_identity" "member" {}

resource "aws_macie2_account" "admin" {
  provider = "awsalternate"
}

resource "aws_macie2_account" "member" {}

resource "aws_macie2_member" "member" {
  provider           = "awsalternate"
  account_id         = data.aws_caller_identity.member.account_id
  email              = %[1]q
  invite             = true
  invitation_message = "This is a message of the invite"
  depends_on         = [aws_macie2_account.admin]
}

resource "aws_macie2_invitation_accepter" "member" {
  administrator_account_id = data.aws_caller_identity.admin.account_id
  depends_on               = [aws_macie2_member.member]
}
`, email)
}
