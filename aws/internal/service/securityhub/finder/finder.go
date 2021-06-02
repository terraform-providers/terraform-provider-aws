package finder

import (
	"context"

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

func Insight(ctx context.Context, conn *securityhub.SecurityHub, arn string) (*securityhub.Insight, error) {
	input := &securityhub.GetInsightsInput{
		InsightArns: aws.StringSlice([]string{arn}),
		MaxResults:  aws.Int64(1),
	}

	output, err := conn.GetInsightsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Insights) == 0 {
		return nil, nil
	}

	return output.Insights[0], nil
}
