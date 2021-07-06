package lakeformation

import (
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
)

const (
	TableNameAllTables        = "ALL_TABLES"
	TableTypeTable            = "Table"
	TableTypeTableWithColumns = "TableWithColumns"
)

func FilterPermissions(input *lakeformation.ListPermissionsInput, tableType string, columnNames []*string, excludedColumnNames []*string, columnWildcard bool, allPermissions []*lakeformation.PrincipalResourcePermissions) []*lakeformation.PrincipalResourcePermissions {
	// For most Lake Formation resources, filtering within the provider is unnecessary. The input
	// contains everything for AWS to give you back exactly what you want. However, many challenges
	// arise with tables and tables with columns. Perhaps the two biggest problems (so far) are as
	// follows:
	// 1. SELECT - when you grant SELECT, it may be part of a list of permissions. However, when
	//    listing permissions, SELECT comes back in a separate permission.
	// 2. Tables with columns. The ListPermissionsInput does not allow you to include a tables with
	//    columns resource. This means you might get back more permissions than actually pertain to
	//    the current situation. The table may have separate permissions that also come back.
	//
	// Thus, for most cases this is just a pass through filter but attempts to clean out
	// permissions in the special cases to avoid extra permissions being included.

	if input.Resource.Catalog != nil {
		return FilterLakeFormationCatalogPermissions(allPermissions)
	}

	if input.Resource.DataLocation != nil {
		return FilterLakeFormationDataLocationPermissions(allPermissions)
	}

	if input.Resource.Database != nil {
		return FilterLakeFormationDatabasePermissions(allPermissions)
	}

	if tableType == TableTypeTableWithColumns {
		return FilterLakeFormationTableWithColumnsPermissions(input.Resource.Table, columnNames, excludedColumnNames, columnWildcard, allPermissions)
	}

	if input.Resource.Table != nil || tableType == TableTypeTable {
		return FilterLakeFormationTablePermissions(input.Resource.Table, allPermissions)
	}

	return nil
}

func FilterLakeFormationTablePermissions(table *lakeformation.TableResource, allPermissions []*lakeformation.PrincipalResourcePermissions) []*lakeformation.PrincipalResourcePermissions {
	// CREATE PERMS (in)     = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT, SELECT on Table, Name = (Table Name)
	//      LIST PERMS (out) = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT         on Table, Name = (Table Name)
	//      LIST PERMS (out) = SELECT                                             on TableWithColumns, Name = (Table Name), ColumnWildcard

	// CREATE PERMS (in)       = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT, SELECT on Table, TableWildcard
	//        LIST PERMS (out) = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT         on Table, TableWildcard, Name = ALL_TABLES
	//        LIST PERMS (out) = SELECT                                             on TableWithColumns, Name = ALL_TABLES, ColumnWildcard

	var cleanPermissions []*lakeformation.PrincipalResourcePermissions

	for _, perm := range allPermissions {
		if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnWildcard != nil {
			if aws.StringValue(perm.Resource.TableWithColumns.Name) == aws.StringValue(table.Name) || (table.TableWildcard != nil && aws.StringValue(perm.Resource.TableWithColumns.Name) == TableNameAllTables) {
				if len(perm.Permissions) > 0 && aws.StringValue(perm.Permissions[0]) == lakeformation.PermissionSelect {
					cleanPermissions = append(cleanPermissions, perm)
					continue
				}

				if len(perm.PermissionsWithGrantOption) > 0 && aws.StringValue(perm.PermissionsWithGrantOption[0]) == lakeformation.PermissionSelect {
					cleanPermissions = append(cleanPermissions, perm)
					continue
				}
			}
		}

		if perm.Resource.Table != nil && aws.StringValue(perm.Resource.Table.DatabaseName) == aws.StringValue(table.DatabaseName) {
			if aws.StringValue(perm.Resource.Table.Name) == aws.StringValue(table.Name) {
				cleanPermissions = append(cleanPermissions, perm)
				continue
			}

			if perm.Resource.Table.TableWildcard != nil && table.TableWildcard != nil {
				cleanPermissions = append(cleanPermissions, perm)
				continue
			}
		}
		continue
	}

	return cleanPermissions
}

func FilterLakeFormationTableWithColumnsPermissions(twc *lakeformation.TableResource, columnNames []*string, excludedColumnNames []*string, columnWildcard bool, allPermissions []*lakeformation.PrincipalResourcePermissions) []*lakeformation.PrincipalResourcePermissions {
	// CREATE PERMS (in)       = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT, SELECT on TableWithColumns, Name = (Table Name), ColumnWildcard
	//        LIST PERMS (out) = ALL, ALTER, DELETE, DESCRIBE, DROP, INSERT         on Table, Name = (Table Name)
	//        LIST PERMS (out) = SELECT                                             on TableWithColumns, Name = (Table Name), ColumnWildcard

	var cleanPermissions []*lakeformation.PrincipalResourcePermissions

	for _, perm := range allPermissions {
		if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnNames != nil {
			if StringSlicesEqualIgnoreOrder(perm.Resource.TableWithColumns.ColumnNames, columnNames) {
				cleanPermissions = append(cleanPermissions, perm)
				continue
			}
		}

		if perm.Resource.TableWithColumns != nil && perm.Resource.TableWithColumns.ColumnWildcard != nil && (columnWildcard || len(excludedColumnNames) > 0) {
			if perm.Resource.TableWithColumns.ColumnWildcard.ExcludedColumnNames == nil && len(excludedColumnNames) == 0 {
				cleanPermissions = append(cleanPermissions, perm)
				continue
			}

			if len(excludedColumnNames) > 0 && StringSlicesEqualIgnoreOrder(perm.Resource.TableWithColumns.ColumnWildcard.ExcludedColumnNames, excludedColumnNames) {
				cleanPermissions = append(cleanPermissions, perm)
				continue
			}
		}

		if perm.Resource.Table != nil && aws.StringValue(perm.Resource.Table.Name) == aws.StringValue(twc.Name) {
			cleanPermissions = append(cleanPermissions, perm)
			continue
		}
	}

	return cleanPermissions
}

func FilterLakeFormationCatalogPermissions(allPermissions []*lakeformation.PrincipalResourcePermissions) []*lakeformation.PrincipalResourcePermissions {
	var cleanPermissions []*lakeformation.PrincipalResourcePermissions

	for _, perm := range allPermissions {
		if perm.Resource.Catalog != nil {
			cleanPermissions = append(cleanPermissions, perm)
		}
	}

	return cleanPermissions
}

func FilterLakeFormationDataLocationPermissions(allPermissions []*lakeformation.PrincipalResourcePermissions) []*lakeformation.PrincipalResourcePermissions {
	var cleanPermissions []*lakeformation.PrincipalResourcePermissions

	for _, perm := range allPermissions {
		if perm.Resource.DataLocation != nil {
			cleanPermissions = append(cleanPermissions, perm)
		}
	}

	return cleanPermissions
}

func FilterLakeFormationDatabasePermissions(allPermissions []*lakeformation.PrincipalResourcePermissions) []*lakeformation.PrincipalResourcePermissions {
	var cleanPermissions []*lakeformation.PrincipalResourcePermissions

	for _, perm := range allPermissions {
		if perm.Resource.Database != nil {
			cleanPermissions = append(cleanPermissions, perm)
		}
	}

	return cleanPermissions
}

func StringSlicesEqualIgnoreOrder(s1, s2 []*string) bool {
	if len(s1) != len(s2) {
		return false
	}

	v1 := aws.StringValueSlice(s1)
	v2 := aws.StringValueSlice(s2)

	sort.Strings(v1)
	sort.Strings(v2)

	return reflect.DeepEqual(v1, v2)
}
