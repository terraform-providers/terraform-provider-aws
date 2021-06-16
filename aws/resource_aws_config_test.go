package aws

import (
	"testing"
)

func TestAccAWSConfig_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Config": {
			"basic":            testAccConfigConfigRule_basic,
			"ownerAws":         testAccConfigConfigRule_ownerAws,
			"customlambda":     testAccConfigConfigRule_customlambda,
			"importAws":        testAccConfigConfigRule_importAws,
			"importLambda":     testAccConfigConfigRule_importLambda,
			"scopeTagKey":      testAccConfigConfigRule_Scope_TagKey,
			"scopeTagKeyEmpty": testAccConfigConfigRule_Scope_TagKey_Empty,
			"scopeTagValue":    testAccConfigConfigRule_Scope_TagValue,
			"tags":             testAccConfigConfigRule_tags,
		},
		"ConfigurationRecorderStatus": {
			"basic":        testAccConfigConfigurationRecorderStatus_basic,
			"startEnabled": testAccConfigConfigurationRecorderStatus_startEnabled,
			"importBasic":  testAccConfigConfigurationRecorderStatus_importBasic,
		},
		"ConfigurationRecorder": {
			"basic":       testAccConfigConfigurationRecorder_basic,
			"allParams":   testAccConfigConfigurationRecorder_allParams,
			"importBasic": testAccConfigConfigurationRecorder_importBasic,
		},
		"ConformancePack": {
			"basic":                     testAccConfigConformancePack_basic,
			"disappears":                testAccConfigConformancePack_disappears,
			"forceNew":                  testAccConfigConformancePack_forceNew,
			"inputParameters":           testAccConfigConformancePack_inputParameters,
			"S3Delivery":                testAccConfigConformancePack_S3Delivery,
			"S3Template":                testAccConfigConformancePack_S3Template,
			"S3TemplateAndTemplateBody": testAccConfigConformancePack_S3TemplateAndTemplateBody,
			"updateInputParameters":     testAccConfigConformancePack_updateInputParameters,
			"updateS3Delivery":          testAccConfigConformancePack_updateS3Delivery,
			"updateS3Template":          testAccConfigConformancePack_updateS3Template,
			"updateTemplateBody":        testAccConfigConformancePack_updateTemplateBody,
		},
		"DeliveryChannel": {
			"basic":       testAccConfigDeliveryChannel_basic,
			"allParams":   testAccConfigDeliveryChannel_allParams,
			"importBasic": testAccConfigDeliveryChannel_importBasic,
		},
		"OrganizationCustomRule": {
			"basic":                     testAccConfigOrganizationCustomRule_basic,
			"disappears":                testAccConfigOrganizationCustomRule_disappears,
			"errorHandling":             testAccConfigOrganizationCustomRule_errorHandling,
			"Description":               testAccConfigOrganizationCustomRule_Description,
			"ExcludedAccounts":          testAccConfigOrganizationCustomRule_ExcludedAccounts,
			"InputParameters":           testAccConfigOrganizationCustomRule_InputParameters,
			"LambdaFunctionArn":         testAccConfigOrganizationCustomRule_LambdaFunctionArn,
			"MaximumExecutionFrequency": testAccConfigOrganizationCustomRule_MaximumExecutionFrequency,
			"ResourceIdScope":           testAccConfigOrganizationCustomRule_ResourceIdScope,
			"ResourceTypesScope":        testAccConfigOrganizationCustomRule_ResourceTypesScope,
			"TagKeyScope":               testAccConfigOrganizationCustomRule_TagKeyScope,
			"TagValueScope":             testAccConfigOrganizationCustomRule_TagValueScope,
			"TriggerTypes":              testAccConfigOrganizationCustomRule_TriggerTypes,
		},
		"OrganizationManagedRule": {
			"basic":                     testAccConfigOrganizationManagedRule_basic,
			"disappears":                testAccConfigOrganizationManagedRule_disappears,
			"errorHandling":             testAccConfigOrganizationManagedRule_errorHandling,
			"Description":               testAccConfigOrganizationManagedRule_Description,
			"ExcludedAccounts":          testAccConfigOrganizationManagedRule_ExcludedAccounts,
			"InputParameters":           testAccConfigOrganizationManagedRule_InputParameters,
			"MaximumExecutionFrequency": testAccConfigOrganizationManagedRule_MaximumExecutionFrequency,
			"ResourceIdScope":           testAccConfigOrganizationManagedRule_ResourceIdScope,
			"ResourceTypesScope":        testAccConfigOrganizationManagedRule_ResourceTypesScope,
			"RuleIdentifier":            testAccConfigOrganizationManagedRule_RuleIdentifier,
			"TagKeyScope":               testAccConfigOrganizationManagedRule_TagKeyScope,
			"TagValueScope":             testAccConfigOrganizationManagedRule_TagValueScope,
		},
		"RemediationConfiguration": {
			"basic":      testAccConfigRemediationConfiguration_basic,
			"disappears": testAccConfigRemediationConfiguration_disappears,
			"recreates":  testAccConfigRemediationConfiguration_recreates,
			"updates":    testAccConfigRemediationConfiguration_updates,
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
