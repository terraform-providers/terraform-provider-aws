package aws

import (
	"testing"
)

func TestAccAWSMacie2_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Account": {
			"basic":                        testAccAwsMacie2Account_basic,
			"finding_publishing_frequency": testAccAwsMacie2Account_FindingPublishingFrequency,
			"status":                       testAccAwsMacie2Account_WithStatus,
			"finding_and_status":           testAccAwsMacie2Account_WithFindingAndStatus,
			"disappears":                   testAccAwsMacie2Account_disappears,
		},
		"ClassificationJob": {
			"basic":          testAccAwsMacie2ClassificationJob_basic,
			"name_generated": testAccAwsMacie2ClassificationJob_Name_Generated,
			"name_prefix":    testAccAwsMacie2ClassificationJob_NamePrefix,
			"disappears":     testAccAwsMacie2ClassificationJob_disappears,
			"status":         testAccAwsMacie2ClassificationJob_Status,
			"complete":       testAccAwsMacie2ClassificationJob_complete,
			"tags":           testAccAwsMacie2ClassificationJob_WithTags,
		},
		"CustomDataIdentifier": {
			"basic":              testAccAwsMacie2CustomDataIdentifier_basic,
			"name_generated":     testAccAwsMacie2CustomDataIdentifier_Name_Generated,
			"name_prefix":        testAccAwsMacie2CustomDataIdentifier_disappears,
			"disappears":         testAccAwsMacie2CustomDataIdentifier_NamePrefix,
			"classification_job": testAccAwsMacie2CustomDataIdentifier_WithClassificationJob,
			"tags":               testAccAwsMacie2CustomDataIdentifier_WithTags,
		},
		"FindingsFilter": {
			"basic":          testAccAwsMacie2FindingsFilter_basic,
			"name_generated": testAccAwsMacie2FindingsFilter_Name_Generated,
			"name_prefix":    testAccAwsMacie2FindingsFilter_NamePrefix,
			"disappears":     testAccAwsMacie2FindingsFilter_disappears,
			"complete":       testAccAwsMacie2FindingsFilter_complete,
			"date":           testAccAwsMacie2FindingsFilter_WithDate,
			"number":         testAccAwsMacie2FindingsFilter_WithNumber,
			"tags":           testAccAwsMacie2FindingsFilter_withTags,
		},
		"OrganizationAdminAccount": {
			"basic":      testAccAwsMacie2OrganizationAdminAccount_basic,
			"disappears": testAccAwsMacie2OrganizationAdminAccount_disappears,
		},
		"Member": {
			"basic":          testAccAwsMacie2Member_basic,
			"disappears":     testAccAwsMacie2Member_disappears,
			"tags":           testAccAwsMacie2Member_withTags,
			"invite":         testAccAwsMacie2Member_invite,
			"invite_removed": testAccAwsMacie2Member_inviteRemoved,
			"status":         testAccAwsMacie2Member_status,
		},
		"InvitationAccepter": {
			"basic": testAccAwsMacie2InvitationAccepter_basic,
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
