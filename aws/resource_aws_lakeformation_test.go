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
		"Permissions": {
			"basic":                      testAccAWSLakeFormationPermissions_basic,
			"dataLocation":               testAccAWSLakeFormationPermissions_dataLocation,
			"database":                   testAccAWSLakeFormationPermissions_database,
			"tableName":                  testAccAWSLakeFormationPermissions_table_name,
			"tableWildcard":              testAccAWSLakeFormationPermissions_table_wildcard,
			"tableWithColumns":           testAccAWSLakeFormationPermissions_tableWithColumns,
			"tableWithColumnsAndTable":   testAccAWSLakeFormationPermissions_tableWithColumnsAndTable,
			"basicDataSource":            testAccAWSLakeFormationPermissionsDataSource_basic,
			"dataLocationDataSource":     testAccAWSLakeFormationPermissionsDataSource_dataLocation,
			"databaseDataSource":         testAccAWSLakeFormationPermissionsDataSource_database,
			"tableDataSource":            testAccAWSLakeFormationPermissionsDataSource_table,
			"tableWithColumnsDataSource": testAccAWSLakeFormationPermissionsDataSource_tableWithColumns,
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
