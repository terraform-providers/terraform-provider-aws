package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfsecurityhub "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub"
)

func testAccAWSSecurityHubMember_basic(t *testing.T) {
	var member securityhub.Member
	resourceName := "aws_securityhub_member.example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubMemberConfig_basic("111111111111", "example@example.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists(resourceName, &member),
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

func testAccAWSSecurityHubMember_invite(t *testing.T) {
	var member securityhub.Member
	resourceName := "aws_securityhub_member.example"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSecurityHubMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSecurityHubMemberConfig_invite("111111111111", "example@example.com", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSecurityHubMemberExists(resourceName, &member),
					resource.TestCheckResourceAttr(resourceName, "member_status", "Invited"),
					resource.TestCheckResourceAttr(resourceName, "invite", "true"),
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

func testAccCheckAWSSecurityHubMemberExists(n string, member *securityhub.Member) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).securityhubconn

		resp, err := conn.GetMembers(&securityhub.GetMembersInput{
			AccountIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(resp.Members) == 0 {
			return fmt.Errorf("Security Hub member %s not found", rs.Primary.ID)
		}

		member = resp.Members[0]

		return nil
	}
}

func testAccCheckAWSSecurityHubMemberDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).securityhubconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_member" {
			continue
		}

		resp, err := conn.GetMembers(&securityhub.GetMembersInput{
			AccountIds: []*string{aws.String(rs.Primary.ID)},
		})

		if tfawserr.ErrCodeEquals(err, tfsecurityhub.ErrCodeBadRequestException) {
			continue
		}

		if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Security Hub Member (%s): %w", rs.Primary.ID, err)
		}

		if len(resp.Members) != 0 {
			return fmt.Errorf("Security Hub member still exists")
		}

		return nil
	}

	return nil
}

func testAccAWSSecurityHubMemberConfig_basic(accountId, email string) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  depends_on = [aws_securityhub_account.example]
  account_id = "%s"
  email      = "%s"
}
`, accountId, email)
}

func testAccAWSSecurityHubMemberConfig_invite(accountId, email string, invite bool) string {
	return fmt.Sprintf(`
resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_member" "example" {
  depends_on = [aws_securityhub_account.example]
  account_id = "%s"
  email      = "%s"
  invite     = %t
}
`, accountId, email, invite)
}
