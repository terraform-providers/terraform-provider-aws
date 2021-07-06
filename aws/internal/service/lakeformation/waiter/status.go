package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tflakeformation "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lakeformation"
)

func PermissionsStatus(conn *lakeformation.LakeFormation, input *lakeformation.ListPermissionsInput, tableType string, columnNames []*string, excludedColumnNames []*string, columnWildcard bool) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		var permissions []*lakeformation.PrincipalResourcePermissions

		err := conn.ListPermissionsPages(input, func(resp *lakeformation.ListPermissionsOutput, lastPage bool) bool {
			for _, permission := range resp.PrincipalResourcePermissions {
				if permission == nil {
					continue
				}

				permissions = append(permissions, permission)
			}
			return !lastPage
		})

		if tfawserr.ErrCodeEquals(err, lakeformation.ErrCodeEntityNotFoundException) {
			return nil, StatusNotFound, err
		}

		if tfawserr.ErrMessageContains(err, lakeformation.ErrCodeInvalidInputException, "Invalid principal") {
			return nil, StatusIAMDelay, nil
		}

		if err != nil {
			return nil, StatusFailed, fmt.Errorf("error listing permissions: %w", err)
		}

		// clean permissions = filter out permissions that do not pertain to this specific resource
		cleanPermissions := tflakeformation.FilterPermissions(input, tableType, columnNames, excludedColumnNames, columnWildcard, permissions)

		if len(cleanPermissions) == 0 {
			return nil, StatusNotFound, nil
		}

		return permissions, StatusAvailable, nil
	}
}
