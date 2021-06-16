package aws

import (
	"testing"
	"time"
)

// Serialize to limit API rate-limit exceeded errors.
func TestAccAWSAmplify_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"App": {
			"basic":                    testAccAWSAmplifyApp_basic,
			"disappears":               testAccAWSAmplifyApp_disappears,
			"Tags":                     testAccAWSAmplifyApp_Tags,
			"AutoBranchCreationConfig": testAccAWSAmplifyApp_AutoBranchCreationConfig,
			"BasicAuthCredentials":     testAccAWSAmplifyApp_BasicAuthCredentials,
			"BuildSpec":                testAccAWSAmplifyApp_BuildSpec,
			"CustomRules":              testAccAWSAmplifyApp_CustomRules,
			"Description":              testAccAWSAmplifyApp_Description,
			"EnvironmentVariables":     testAccAWSAmplifyApp_EnvironmentVariables,
			"IamServiceRole":           testAccAWSAmplifyApp_IamServiceRole,
			"Name":                     testAccAWSAmplifyApp_Name,
			"Repository":               testAccAWSAmplifyApp_Repository,
		},
		"BackendEnvironment": {
			"basic":                         testAccAWSAmplifyBackendEnvironment_basic,
			"disappears":                    testAccAWSAmplifyBackendEnvironment_disappears,
			"DeploymentArtifacts_StackName": testAccAWSAmplifyBackendEnvironment_DeploymentArtifacts_StackName,
		},
		"Branch": {
			"basic":                testAccAWSAmplifyBranch_basic,
			"disappears":           testAccAWSAmplifyBranch_disappears,
			"Tags":                 testAccAWSAmplifyBranch_Tags,
			"BasicAuthCredentials": testAccAWSAmplifyBranch_BasicAuthCredentials,
			"EnvironmentVariables": testAccAWSAmplifyBranch_EnvironmentVariables,
			"OptionalArguments":    testAccAWSAmplifyBranch_OptionalArguments,
		},
		"DomainAssociation": {
			"basic":      testAccAWSAmplifyDomainAssociation_basic,
			"disappears": testAccAWSAmplifyDomainAssociation_disappears,
			"update":     testAccAWSAmplifyDomainAssociation_update,
		},
		"Webhook": {
			"basic":      testAccAWSAmplifyWebhook_basic,
			"disappears": testAccAWSAmplifyWebhook_disappears,
			"update":     testAccAWSAmplifyWebhook_update,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
					// Explicitly sleep between tests.
					time.Sleep(5 * time.Second)
				})
			}
		})
	}
}
