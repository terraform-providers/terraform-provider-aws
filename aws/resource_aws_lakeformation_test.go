package aws

import (
	"testing"
)

func TestAccAWSLakeFormation_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"DataLakeSettings": {
			"basic":            testAccAWSLakeFormationDataLakeSettings_basic,
			"dataSource":       testAccAWSLakeFormationDataLakeSettingsDataSource_basic,
			"disappears":       testAccAWSLakeFormationDataLakeSettings_disappears,
			"withoutCatalogId": testAccAWSLakeFormationDataLakeSettings_withoutCatalogId,
		},
		"PermissionsBasic": {
			"basic":              testAccAWSLakeFormationPermissions_basic,
			"database":           testAccAWSLakeFormationPermissions_database,
			"databaseIAMAllowed": testAccAWSLakeFormationPermissions_databaseIAMAllowed,
			"databaseMultiple":   testAccAWSLakeFormationPermissions_databaseMultiple,
			"dataLocation":       testAccAWSLakeFormationPermissions_dataLocation,
			"disappears":         testAccAWSLakeFormationPermissions_disappears,
		},
		"PermissionsDataSource": {
			"basic":            testAccAWSLakeFormationPermissionsDataSource_basic,
			"database":         testAccAWSLakeFormationPermissionsDataSource_database,
			"dataLocation":     testAccAWSLakeFormationPermissionsDataSource_dataLocation,
			"table":            testAccAWSLakeFormationPermissionsDataSource_table,
			"tableWithColumns": testAccAWSLakeFormationPermissionsDataSource_tableWithColumns,
		},
		"PermissionsTable": {
			"basic":              testAccAWSLakeFormationPermissions_tableBasic,
			"iamAllowed":         testAccAWSLakeFormationPermissions_tableIAMAllowed,
			"implicit":           testAccAWSLakeFormationPermissions_tableImplicit,
			"multipleRoles":      testAccAWSLakeFormationPermissions_tableMultipleRoles,
			"selectOnly":         testAccAWSLakeFormationPermissions_tableSelectOnly,
			"selectPlus":         testAccAWSLakeFormationPermissions_tableSelectPlus,
			"wildcardNoSelect":   testAccAWSLakeFormationPermissions_tableWildcardNoSelect,
			"wildcardSelectOnly": testAccAWSLakeFormationPermissions_tableWildcardSelectOnly,
			"wildcardSelectPlus": testAccAWSLakeFormationPermissions_tableWildcardSelectPlus,
		},
		"PermissionsTableWithColumns": {
			"basic":                   testAccAWSLakeFormationPermissions_twcBasic,
			"implicit":                testAccAWSLakeFormationPermissions_twcImplicit,
			"wildcardExcludedColumns": testAccAWSLakeFormationPermissions_twcWildcardExcludedColumns,
			"wildcardSelectOnly":      testAccAWSLakeFormationPermissions_twcWildcardSelectOnly,
			"wildcardSelectPlus":      testAccAWSLakeFormationPermissions_twcWildcardSelectPlus,
		},
		"PolicyTag": {
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
