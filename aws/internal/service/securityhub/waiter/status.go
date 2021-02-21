package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/finder"
)

const (
	// AdminStatus NotFound
	AdminStatusNotFound = "NotFound"

	// AdminStatus Unknown
	AdminStatusUnknown = "Unknown"
)

// AdminAccountAdminStatus fetches the AdminAccount and its AdminStatus
func AdminAccountAdminStatus(conn *securityhub.SecurityHub, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := finder.AdminAccount(conn, adminAccountID)

		if err != nil {
			return nil, AdminStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, AdminStatusNotFound, nil
		}

		return adminAccount, aws.StringValue(adminAccount.Status), nil
	}
}
