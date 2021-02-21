package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsGuardDutyInviteAccepter_basic(t *testing.T) {
	var providers []*schema.Provider
	masterDetectorResourceName := "aws_guardduty_detector.master"
	memberDetectorResourceName := "aws_guardduty_detector.member"
	resourceName := "aws_guardduty_invite_accepter.test"
	_, email := testAccAWSGuardDutyMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsGuardDutyInviteAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsGuardDutyInviteAccepterConfig_basic(email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyInviteAccepterExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "detector_id", memberDetectorResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "master_account_id", masterDetectorResourceName, "account_id"),
				),
			},
			{
				Config:            testAccAwsGuardDutyInviteAccepterConfig_basic(email),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsGuardDutyInviteAccepterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_invite_accepter" {
			continue
		}

		input := &guardduty.GetMasterAccountInput{
			DetectorId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetMasterAccount(input)

		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			return nil
		}

		if err != nil {
			return err
		}

		if output == nil || output.Master == nil || aws.StringValue(output.Master.AccountId) != rs.Primary.Attributes["master_account_id"] {
			continue
		}

		return fmt.Errorf("GuardDuty Detector (%s) still has GuardDuty Master Account ID (%s)", rs.Primary.ID, aws.StringValue(output.Master.AccountId))
	}

	return nil
}

func testAccCheckAwsGuardDutyInviteAccepterExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) has empty ID", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

		input := &guardduty.GetMasterAccountInput{
			DetectorId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetMasterAccount(input)

		if err != nil {
			return err
		}

		if output == nil || output.Master == nil || aws.StringValue(output.Master.AccountId) == "" {
			return fmt.Errorf("no master account found for: %s", resourceName)
		}

		return nil
	}
}

func testAccAwsGuardDutyInviteAccepterConfig_basic(email string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_guardduty_detector" "master" {
  provider = "awsalternate"
}

resource "aws_guardduty_detector" "member" {}

resource "aws_guardduty_member" "member" {
  provider = "awsalternate"

  account_id                 = aws_guardduty_detector.member.account_id
  detector_id                = aws_guardduty_detector.master.id
  disable_email_notification = true
  email                      = %q
  invite                     = true
}

resource "aws_guardduty_invite_accepter" "test" {
  depends_on = [aws_guardduty_member.member]

  detector_id       = aws_guardduty_detector.member.id
  master_account_id = aws_guardduty_detector.master.account_id
}
`, email)
}
