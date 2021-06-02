package aws

import (
	"testing"
)

func TestAccAWSSecurityHub_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic": testAccAWSSecurityHubAccount_basic,
		},
		"Member": {
			"basic":  testAccAWSSecurityHubMember_basic,
			"invite": testAccAWSSecurityHubMember_invite,
		},
		"ActionTarget": {
			"basic":       testAccAwsSecurityHubActionTarget_basic,
			"disappears":  testAccAwsSecurityHubActionTarget_disappears,
			"Description": testAccAwsSecurityHubActionTarget_Description,
			"Name":        testAccAwsSecurityHubActionTarget_Name,
		},
		"Insight": {
			"basic":            testAccAwsSecurityHubInsight_basic,
			"disappears":       testAccAwsSecurityHubInsight_disappears,
			"DateFilters":      testAccAwsSecurityHubInsight_DateFilters,
			"GroupByAttribute": testAccAwsSecurityHubInsight_GroupByAttribute,
			"IpFilters":        testAccAwsSecurityHubInsight_IpFilters,
			"KeywordFilters":   testAccAwsSecurityHubInsight_KeywordFilters,
			"MapFilters":       testAccAwsSecurityHubInsight_MapFilters,
			"MultipleFilters":  testAccAwsSecurityHubInsight_MultipleFilters,
			"Name":             testAccAwsSecurityHubInsight_Name,
			"NumberFilters":    testAccAwsSecurityHubInsight_NumberFilters,
			"WorkflowStatus":   testAccAwsSecurityHubInsight_WorkflowStatus,
		},
		"InviteAccepter": {
			"basic": testAccAWSSecurityHubInviteAccepter_basic,
		},
		"OrganizationAdminAccount": {
			"basic":       testAccAwsSecurityHubOrganizationAdminAccount_basic,
			"disappears":  testAccAwsSecurityHubOrganizationAdminAccount_disappears,
			"MultiRegion": testAccAwsSecurityHubOrganizationAdminAccount_MultiRegion,
		},
		"ProductSubscription": {
			"basic": testAccAWSSecurityHubProductSubscription_basic,
		},
		"StandardsSubscription": {
			"basic": testAccAWSSecurityHubStandardsSubscription_basic,
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
