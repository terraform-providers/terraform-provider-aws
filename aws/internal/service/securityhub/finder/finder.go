package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
)

func AdminAccount(conn *securityhub.SecurityHub, adminAccountID string) (*securityhub.AdminAccount, error) {
	input := &securityhub.ListOrganizationAdminAccountsInput{}
	var result *securityhub.AdminAccount

	err := conn.ListOrganizationAdminAccountsPages(input, func(page *securityhub.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, adminAccount := range page.AdminAccounts {
			if adminAccount == nil {
				continue
			}

			if aws.StringValue(adminAccount.AccountId) == adminAccountID {
				result = adminAccount
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

// EnabledStandards returns a list of the standards that are currently enabled.
func EnabledStandardsSubscriptions(conn *securityhub.SecurityHub) ([]*securityhub.StandardsSubscription, error) {
	input := &securityhub.GetEnabledStandardsInput{}
	var results []*securityhub.StandardsSubscription

	err := conn.GetEnabledStandardsPages(input, func(page *securityhub.GetEnabledStandardsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		results = append(results, page.StandardsSubscriptions...)

		return !lastPage
	})

	return results, err
}
