package lakeformation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
)

func TestFilterPermissions(t *testing.T) {
	// primitives to make test cases easier
	accountID := "481516234248"
	dbName := "Hiliji"
	altDBName := "Hiuhbum"
	tableName := "Ladocmoc"

	principal := &lakeformation.DataLakePrincipal{
		DataLakePrincipalIdentifier: aws.String(fmt.Sprintf("arn:aws-us-gov:iam::%s:role/Zepotiz-Bulgaria", accountID)),
	}

	testCases := []struct {
		Name                string
		Input               *lakeformation.ListPermissionsInput
		TableType           string
		ColumnNames         []*string
		ExcludedColumnNames []*string
		ColumnWildcard      bool
		All                 []*lakeformation.PrincipalResourcePermissions
		ExpectedClean       []*lakeformation.PrincipalResourcePermissions
	}{
		{
			Name: "empty",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource:  &lakeformation.Resource{},
			},
			All:           nil,
			ExpectedClean: nil,
		},
		{
			Name: "emptyWithInput",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All:           nil,
			ExpectedClean: nil,
		},
		{
			Name: "wrongTableResource", // this may not actually be possible but we account for it
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(altDBName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: nil,
		},
		{
			Name: "tableResource",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
		},
		{
			Name: "tableResourceSelectPerm",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &lakeformation.ColumnWildcard{},
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &lakeformation.ColumnWildcard{},
						},
					},
				},
			},
		},
		{
			Name: "tableResourceSelectPermGrant",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{lakeformation.PermissionSelect}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &lakeformation.ColumnWildcard{},
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter}),
					PermissionsWithGrantOption: aws.StringSlice([]string{lakeformation.PermissionAlter}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{lakeformation.PermissionSelect}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
							ColumnWildcard: &lakeformation.ColumnWildcard{},
						},
					},
				},
			},
		},
		{
			Name: "twcBasic",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			TableType:   TableTypeTableWithColumns,
			ColumnNames: aws.StringSlice([]string{"value"}),
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  aws.StringSlice([]string{"value"}),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  aws.StringSlice([]string{"fred"}),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  aws.StringSlice([]string{"value"}),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
		},
		{
			Name: "twcWildcard",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			TableType:      TableTypeTableWithColumns,
			ColumnWildcard: true,
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							ColumnWildcard: &lakeformation.ColumnWildcard{},
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  aws.StringSlice([]string{"fred"}),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:      aws.String(accountID),
							ColumnWildcard: &lakeformation.ColumnWildcard{},
							DatabaseName:   aws.String(dbName),
							Name:           aws.String(tableName),
						},
					},
				},
			},
		},
		{
			Name: "twcWildcardExcluded",
			Input: &lakeformation.ListPermissionsInput{
				Principal: principal,
				Resource: &lakeformation.Resource{
					Table: &lakeformation.TableResource{
						CatalogId:    aws.String(accountID),
						DatabaseName: aws.String(dbName),
						Name:         aws.String(tableName),
					},
				},
			},
			TableType:           TableTypeTableWithColumns,
			ColumnWildcard:      true,
			ExcludedColumnNames: aws.StringSlice([]string{"value"}),
			All: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId: aws.String(accountID),
							ColumnWildcard: &lakeformation.ColumnWildcard{
								ExcludedColumnNames: aws.StringSlice([]string{"value"}),
							},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId:    aws.String(accountID),
							ColumnNames:  aws.StringSlice([]string{"fred"}),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
			ExpectedClean: []*lakeformation.PrincipalResourcePermissions{
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionAlter, lakeformation.PermissionDelete}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						Table: &lakeformation.TableResource{
							CatalogId:    aws.String(accountID),
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
				{
					Permissions:                aws.StringSlice([]string{lakeformation.PermissionSelect}),
					PermissionsWithGrantOption: aws.StringSlice([]string{}),
					Principal:                  principal,
					Resource: &lakeformation.Resource{
						TableWithColumns: &lakeformation.TableWithColumnsResource{
							CatalogId: aws.String(accountID),
							ColumnWildcard: &lakeformation.ColumnWildcard{
								ExcludedColumnNames: aws.StringSlice([]string{"value"}),
							},
							DatabaseName: aws.String(dbName),
							Name:         aws.String(tableName),
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := FilterPermissions(testCase.Input, testCase.TableType, testCase.ColumnNames, testCase.ExcludedColumnNames, testCase.ColumnWildcard, testCase.All)

			if !reflect.DeepEqual(testCase.ExpectedClean, got) {
				t.Errorf("got %v, expected %v, input %v", got, testCase.ExpectedClean, testCase.Input)
			}
		})
	}
}
