package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/finder"
)

const (
	// AdminStatus NotFound
	AdminStatusNotFound = "NotFound"

	// AdminStatus Unknown
	AdminStatusUnknown = "Unknown"

	// StandardsStatus NotFound
	StandardsStatusNotFound = "NotFound"

	// StandardsStatus Unknown
	StandardsStatusUnknown = "Unknown"
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

// StandardsSubscriptionsStatus fetches the enabled Standards Subscriptions
func StandardsSubscriptionsStatus(conn *securityhub.SecurityHub) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		standardsSubscriptions, err := finder.EnabledStandardsSubscriptions(conn)

		if err != nil {
			return nil, StandardsStatusUnknown, err
		}

		if len(standardsSubscriptions) == 0 {
			return standardsSubscriptions, StandardsStatusNotFound, nil
		}

		statuses := schema.NewSet(schema.HashString, nil)

		for _, subscription := range standardsSubscriptions {
			if subscription == nil {
				continue
			}

			statuses.Add(aws.StringValue(subscription.StandardsStatus))
		}

		var status string

		// If any status is "FAILED", it should take precedence over others.
		// If all statuses are INCOMPLETE, mark as READY as they are still enabled by the API.
		if statuses.Contains(securityhub.StandardsStatusFailed) {
			status = securityhub.StandardsStatusFailed
		} else if statuses.Contains(securityhub.StandardsStatusIncomplete) {
			if statuses.Len() == 1 {
				status = securityhub.StandardsStatusReady
			} else {
				status = securityhub.StandardsStatusIncomplete
			}
		} else if statuses.Contains(securityhub.StandardsStatusPending) {
			status = securityhub.StandardsStatusPending
		} else if statuses.Contains(securityhub.StandardsStatusDeleting) {
			status = securityhub.StandardsStatusDeleting
		} else if statuses.Contains(securityhub.StandardsStatusReady) {
			status = securityhub.StandardsStatusReady
		} else {
			status = StandardsStatusUnknown
		}

		return standardsSubscriptions, status, nil
	}
}
