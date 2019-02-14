package aws

import (
	"os"
	"testing"
)

func TestAccAWSGuardDuty(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Detector": {
			"basic":  testAccAwsGuardDutyDetector_basic,
			"import": testAccAwsGuardDutyDetector_import,
		},
		"IPSet": {
			"basic":  testAccAwsGuardDutyIpset_basic,
			"import": testAccAwsGuardDutyIpset_import,
		},
		"ThreatIntelSet": {
			"basic":  testAccAwsGuardDutyThreatintelset_basic,
			"import": testAccAwsGuardDutyThreatintelset_import,
		},
		"Member": {
			"basic":              testAccAwsGuardDutyMember_basic,
			"inviteOnUpdate":     testAccAwsGuardDutyMember_invite_onUpdate,
			"inviteDisassociate": testAccAwsGuardDutyMember_invite_disassociate,
			"invitationMessage":  testAccAwsGuardDutyMember_invitationMessage,
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

func testAccAWSGuardDutyMemberFromEnv(t *testing.T) (string, string) {
	accountID := os.Getenv("AWS_GUARDDUTY_MEMBER_ACCOUNT_ID")
	if accountID == "" {
		t.Skip(
			"Environment variable AWS_GUARDDUTY_MEMBER_ACCOUNT_ID is not set. " +
				"To properly test inviting GuardDuty member accounts, " +
				"a valid AWS account ID must be provided.")
	}
	email := os.Getenv("AWS_GUARDDUTY_MEMBER_EMAIL")
	if email == "" {
		t.Skip(
			"Environment variable AWS_GUARDDUTY_MEMBER_EMAIL is not set. " +
				"To properly test inviting GuardDuty member accounts, " +
				"a valid email associated with the AWS_GUARDDUTY_MEMBER_ACCOUNT_ID must be provided.")
	}
	return accountID, email
}
