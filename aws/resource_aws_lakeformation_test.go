package aws

import (
	"testing"
)

func TestAccAWSLakeFormation_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"DataLakeSettings": {
			"basic":            testAccAWSLakeFormationDataLakeSettings_basic,
			"disappears":       testAccAWSLakeFormationDataLakeSettings_disappears,
			"withoutCatalogId": testAccAWSLakeFormationDataLakeSettings_withoutCatalogId,
			"dataSource":       testAccAWSLakeFormationDataLakeSettingsDataSource_basic,
		},
		"BasicPermissions": {
			"basic":        testAccAWSLakeFormationPermissions_basic,
			"dataLocation": testAccAWSLakeFormationPermissions_dataLocation,
			"database":     testAccAWSLakeFormationPermissions_database,
			"policyTag":    testAccAWSLakeFormationPermissions_policy_tag,
		},
		"TablePermissions": {
			"columnWildcardPermissions":           testAccAWSLakeFormationPermissions_columnWildcardPermissions,
			"implicitTableWithColumnsPermissions": testAccAWSLakeFormationPermissions_implicitTableWithColumnsPermissions,
			"implicitTablePermissions":            testAccAWSLakeFormationPermissions_implicitTablePermissions,
			"selectPermissions":                   testAccAWSLakeFormationPermissions_selectPermissions,
			"tableName":                           testAccAWSLakeFormationPermissions_tableName,
			"tableWildcard":                       testAccAWSLakeFormationPermissions_tableWildcard,
			"tableWildcardPermissions":            testAccAWSLakeFormationPermissions_tableWildcardPermissions,
			"tableWithColumns":                    testAccAWSLakeFormationPermissions_tableWithColumns,
		},
		"DataSourcePermissions": {
			"basicDataSource":            testAccAWSLakeFormationPermissionsDataSource_basic,
			"dataLocationDataSource":     testAccAWSLakeFormationPermissionsDataSource_dataLocation,
			"databaseDataSource":         testAccAWSLakeFormationPermissionsDataSource_database,
			"policyTagDataSource":        testAccAWSLakeFormationPermissionsDataSource_policy_tag,
			"tableDataSource":            testAccAWSLakeFormationPermissionsDataSource_table,
			"tableWithColumnsDataSource": testAccAWSLakeFormationPermissionsDataSource_tableWithColumns,
		},
		"PolicyTags": {
			"basic":      testAccAWSLakeFormationPolicyTag_basic,
			"disappears": testAccAWSLakeFormationPolicyTag_disappears,
			"values":     testAccAWSLakeFormationPolicyTag_values,
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
