package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccAwsGuardDutyMember_basic(t *testing.T) {
	resourceName := "aws_guardduty_member.test"
	accountID := "111111111111"
	email := "required@example.com"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyMemberConfig_basic(accountID, email),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Created"),
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

func testAccAwsGuardDutyMember_invite_disassociate(t *testing.T) {
	resourceName := "aws_guardduty_member.test"
	accountID, email := testAccAWSGuardDutyMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyMemberConfig_invite(accountID, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", "true"),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Invited"),
				),
			},
			// Disassociate member
			{
				Config: testAccGuardDutyMemberConfig_invite(accountID, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", "false"),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Removed"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disable_email_notification",
				},
			},
		},
	})
}

func testAccAwsGuardDutyMember_invite_onUpdate(t *testing.T) {
	resourceName := "aws_guardduty_member.test"
	accountID, email := testAccAWSGuardDutyMemberFromEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyMemberConfig_invite(accountID, email, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", "false"),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Created"),
				),
			},
			// Invite member
			{
				Config: testAccGuardDutyMemberConfig_invite(accountID, email, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "invite", "true"),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Invited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disable_email_notification",
				},
			},
		},
	})
}

func testAccAwsGuardDutyMember_invitationMessage(t *testing.T) {
	resourceName := "aws_guardduty_member.test"
	accountID, email := testAccAWSGuardDutyMemberFromEnv(t)
	invitationMessage := "inviting"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsGuardDutyMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGuardDutyMemberConfig_invitationMessage(accountID, email, invitationMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsGuardDutyMemberExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", accountID),
					resource.TestCheckResourceAttrSet(resourceName, "detector_id"),
					resource.TestCheckResourceAttr(resourceName, "disable_email_notification", "true"),
					resource.TestCheckResourceAttr(resourceName, "email", email),
					resource.TestCheckResourceAttr(resourceName, "invite", "true"),
					resource.TestCheckResourceAttr(resourceName, "invitation_message", invitationMessage),
					resource.TestCheckResourceAttr(resourceName, "relationship_status", "Invited"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"disable_email_notification",
					"invitation_message",
				},
			},
		},
	})
}

func testAccCheckAwsGuardDutyMemberDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).guarddutyconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_guardduty_member" {
			continue
		}

		accountID, detectorID, err := decodeGuardDutyMemberID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetMembersInput{
			AccountIds: []*string{aws.String(accountID)},
			DetectorId: aws.String(detectorID),
		}

		gmo, err := conn.GetMembers(input)
		if err != nil {
			if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
				return nil
			}
			return err
		}

		if len(gmo.Members) < 1 {
			continue
		}

		return fmt.Errorf("Expected GuardDuty Detector to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsGuardDutyMemberExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		accountID, detectorID, err := decodeGuardDutyMemberID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &guardduty.GetMembersInput{
			AccountIds: []*string{aws.String(accountID)},
			DetectorId: aws.String(detectorID),
		}

		conn := testAccProvider.Meta().(*AWSClient).guarddutyconn
		gmo, err := conn.GetMembers(input)
		if err != nil {
			return err
		}

		if len(gmo.Members) < 1 {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccGuardDutyMemberConfig_basic(accountID, email string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id  = "%[2]s"
  detector_id = aws_guardduty_detector.test.id
  email       = "%[3]s"
}
`, testAccGuardDutyDetectorConfig_basic1, accountID, email)
}

func testAccGuardDutyMemberConfig_invite(accountID, email string, invite bool) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id                 = "%[2]s"
  detector_id                = aws_guardduty_detector.test.id
  disable_email_notification = true
  email                      = "%[3]s"
  invite                     = %[4]t
}
`, testAccGuardDutyDetectorConfig_basic1, accountID, email, invite)
}

func testAccGuardDutyMemberConfig_invitationMessage(accountID, email, invitationMessage string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_guardduty_member" "test" {
  account_id                 = "%[2]s"
  detector_id                = aws_guardduty_detector.test.id
  disable_email_notification = true
  email                      = "%[3]s"
  invitation_message         = "%[4]s"
  invite                     = true
}
`, testAccGuardDutyDetectorConfig_basic1, accountID, email, invitationMessage)
}
