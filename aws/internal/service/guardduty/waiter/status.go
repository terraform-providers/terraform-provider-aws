package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// AdminStatus NotFound
	AdminStatusNotFound = "NotFound"

	// AdminStatus Unknown
	AdminStatusUnknown = "Unknown"

	// Constants not currently provided by the AWS Go SDK
	PublishingStatusFailed  = "Failed"
	PublishingStatusUnknown = "Unknown"
)

// AdminAccountAdminStatus fetches the AdminAccount and its AdminStatus
func AdminAccountAdminStatus(conn *guardduty.GuardDuty, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := getOrganizationAdminAccount(conn, adminAccountID)

		if err != nil {
			return nil, AdminStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, AdminStatusNotFound, nil
		}

		return adminAccount, aws.StringValue(adminAccount.AdminStatus), nil
	}
}

// PublishingDestinationStatus fetches the PublishingDestination and its Status
func PublishingDestinationStatus(conn *guardduty.GuardDuty, destinationID, detectorID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.DescribePublishingDestinationInput{
			DetectorId:    aws.String(detectorID),
			DestinationId: aws.String(destinationID),
		}

		output, err := conn.DescribePublishingDestination(input)

		if err != nil {
			return output, PublishingStatusFailed, err
		}

		if output == nil {
			return output, PublishingStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// TODO: Migrate to shared internal package for aws package and this package
func getOrganizationAdminAccount(conn *guardduty.GuardDuty, adminAccountID string) (*guardduty.AdminAccount, error) {
	input := &guardduty.ListOrganizationAdminAccountsInput{}
	var result *guardduty.AdminAccount

	err := conn.ListOrganizationAdminAccountsPages(input, func(page *guardduty.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, adminAccount := range page.AdminAccounts {
			if adminAccount == nil {
				continue
			}

			if aws.StringValue(adminAccount.AdminAccountId) == adminAccountID {
				result = adminAccount
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
