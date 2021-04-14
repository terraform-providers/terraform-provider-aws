package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an AdminAccount to return Enabled
	AdminAccountEnabledTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an AdminAccount to return NotFound
	AdminAccountNotFoundTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a PublishingDestination to return Publishing
	PublishingDestinationCreatedTimeout    = 5 * time.Minute
	PublishingDestinationCreatedMinTimeout = 3 * time.Second

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

// PublishingDestinationCreated waits for GuardDuty to return Publishing
func PublishingDestinationCreated(conn *guardduty.GuardDuty, destinationID, detectorID string) (*guardduty.CreatePublishingDestinationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{guardduty.PublishingStatusPendingVerification},
		Target:  []string{guardduty.PublishingStatusPublishing},
		Refresh: PublishingDestinationStatus(conn, destinationID, detectorID),
		Timeout: PublishingDestinationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*guardduty.CreatePublishingDestinationOutput); ok {
		return v, err
	}

	return nil, err
}
