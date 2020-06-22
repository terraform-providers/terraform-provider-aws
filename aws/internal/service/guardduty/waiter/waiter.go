package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// Maximum amount of time to wait for an AdminAccount to return Enabled
	AdminAccountEnabledTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an AdminAccount to return NotFound
	AdminAccountNotFoundTimeout = 5 * time.Minute

	// Maximum amount of time to wait for membership to propagate
	// When removing Organization Admin Accounts, there is eventual
	// consistency even after the account is no longer listed.
	// Reference error message:
	// BadRequestException: The request is rejected because the current account cannot delete detector while it has invited or associated members.
	MembershipPropagationTimeout = 2 * time.Minute
)

// AdminAccountEnabled waits for an AdminAccount to return Enabled
func AdminAccountEnabled(conn *guardduty.GuardDuty, adminAccountID string) (*guardduty.AdminAccount, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{AdminStatusNotFound},
		Target:  []string{guardduty.AdminStatusEnabled},
		Refresh: AdminAccountAdminStatus(conn, adminAccountID),
		Timeout: AdminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*guardduty.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// AdminAccountNotFound waits for an AdminAccount to return NotFound
func AdminAccountNotFound(conn *guardduty.GuardDuty, adminAccountID string) (*guardduty.AdminAccount, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{guardduty.AdminStatusDisableInProgress},
		Target:  []string{AdminStatusNotFound},
		Refresh: AdminAccountAdminStatus(conn, adminAccountID),
		Timeout: AdminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*guardduty.AdminAccount); ok {
		return output, err
	}

	return nil, err
}
