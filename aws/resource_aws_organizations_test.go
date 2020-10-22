package aws

import (
	"testing"
)

func TestAccAWSOrganizations_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Organization": {
			"basic":                      testAccAwsOrganizationsOrganization_basic,
			"AwsServiceAccessPrincipals": testAccAwsOrganizationsOrganization_AwsServiceAccessPrincipals,
			"EnabledPolicyTypes":         testAccAwsOrganizationsOrganization_EnabledPolicyTypes,
			"FeatureSet":                 testAccAwsOrganizationsOrganization_FeatureSet,
			"DataSource":                 testAccDataSourceAwsOrganizationsOrganization_basic,
		},
		"Account": {
			"basic":    testAccAwsOrganizationsAccount_basic,
			"ParentId": testAccAwsOrganizationsAccount_ParentId,
			"Tags":     testAccAwsOrganizationsAccount_Tags,
		},
		"OrganizationalUnit": {
			"basic": testAccAwsOrganizationsOrganizationalUnit_basic,
			"Name":  testAccAwsOrganizationsOrganizationalUnit_Name,
		},
		"OrganizationalUnits": {
			"DataSource": testAccDataSourceAwsOrganizationsOrganizationalUnits_basic,
		},
		"Policy": {
			"basic":                  testAccAwsOrganizationsPolicy_basic,
			"concurrent":             testAccAwsOrganizationsPolicy_concurrent,
			"Description":            testAccAwsOrganizationsPolicy_description,
			"Tags":                   testAccAwsOrganizationsPolicy_tags,
			"disappears":             testAccAwsOrganizationsPolicy_disappears,
			"Type_AI_OPT_OUT":        testAccAwsOrganizationsPolicy_type_AI_OPT_OUT,
			"Type_Backup":            testAccAwsOrganizationsPolicy_type_Backup,
			"Type_SCP":               testAccAwsOrganizationsPolicy_type_SCP,
			"Type_Tag":               testAccAwsOrganizationsPolicy_type_Tag,
			"ImportAwsManagedPolicy": testAccAwsOrganizationsPolicy_ImportAwsManagedPolicy,
		},
		"PolicyAttachment": {
			"Account":            testAccAwsOrganizationsPolicyAttachment_Account,
			"OrganizationalUnit": testAccAwsOrganizationsPolicyAttachment_OrganizationalUnit,
			"Root":               testAccAwsOrganizationsPolicyAttachment_Root,
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
