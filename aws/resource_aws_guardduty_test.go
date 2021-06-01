package aws

import (
	"os"
	"testing"
)

func TestAccAWSGuardDuty_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Detector": {
			"basic":            testAccAwsGuardDutyDetector_basic,
			"tags":             testAccAwsGuardDutyDetector_tags,
			"datasource_basic": testAccAWSGuarddutyDetectorDataSource_basic,
			"datasource_id":    testAccAWSGuarddutyDetectorDataSource_Id,
		},
		"Filter": {
			"basic":      testAccAwsGuardDutyFilter_basic,
			"update":     testAccAwsGuardDutyFilter_update,
			"tags":       testAccAwsGuardDutyFilter_tags,
			"disappears": testAccAwsGuardDutyFilter_disappears,
		},
		"InviteAccepter": {
			"basic": testAccAwsGuardDutyInviteAccepter_basic,
		},
		"IPSet": {
			"basic": testAccAwsGuardDutyIpset_basic,
			"tags":  testAccAwsGuardDutyIpset_tags,
		},
		"OrganizationAdminAccount": {
			"basic": testAccAwsGuardDutyOrganizationAdminAccount_basic,
		},
		"OrganizationConfiguration": {
			"basic": testAccAwsGuardDutyOrganizationConfiguration_basic,
		},
		"ThreatIntelSet": {
			"basic": testAccAwsGuardDutyThreatintelset_basic,
			"tags":  testAccAwsGuardDutyThreatintelset_tags,
		},
		"Member": {
			"basic":              testAccAwsGuardDutyMember_basic,
			"inviteOnUpdate":     testAccAwsGuardDutyMember_invite_onUpdate,
			"inviteDisassociate": testAccAwsGuardDutyMember_invite_disassociate,
			"invitationMessage":  testAccAwsGuardDutyMember_invitationMessage,
		},
		"PublishingDestination": {
			"basic":      testAccAwsGuardDutyPublishingDestination_basic,
			"disappears": testAccAwsGuardDutyPublishingDestination_disappears,
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
