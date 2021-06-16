package finder

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
)

func PortfolioShare(conn *servicecatalog.ServiceCatalog, portfolioID, shareType, principalID string) (*servicecatalog.PortfolioShareDetail, error) {
	input := &servicecatalog.DescribePortfolioSharesInput{
		PortfolioId: aws.String(portfolioID),
		Type:        aws.String(shareType),
	}
	var result *servicecatalog.PortfolioShareDetail

	err := conn.DescribePortfolioSharesPages(input, func(page *servicecatalog.DescribePortfolioSharesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.PortfolioShareDetails {
			if deet == nil {
				continue
			}

			if strings.Contains(principalID, aws.StringValue(deet.PrincipalId)) {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func ProductPortfolioAssociation(conn *servicecatalog.ServiceCatalog, acceptLanguage, portfolioID, productID string) (*servicecatalog.PortfolioDetail, error) {
	// seems odd that the sourcePortfolioID is not returned or searchable...
	input := &servicecatalog.ListPortfoliosForProductInput{
		ProductId: aws.String(productID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	var result *servicecatalog.PortfolioDetail

	err := conn.ListPortfoliosForProductPages(input, func(page *servicecatalog.ListPortfoliosForProductOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.PortfolioDetails {
			if deet == nil {
				continue
			}

			if aws.StringValue(deet.Id) == portfolioID {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func BudgetResourceAssociation(conn *servicecatalog.ServiceCatalog, budgetName, resourceID string) (*servicecatalog.BudgetDetail, error) {
	input := &servicecatalog.ListBudgetsForResourceInput{
		ResourceId: aws.String(resourceID),
	}

	var result *servicecatalog.BudgetDetail

	err := conn.ListBudgetsForResourcePages(input, func(page *servicecatalog.ListBudgetsForResourceOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, budget := range page.Budgets {
			if budget == nil {
				continue
			}

			if aws.StringValue(budget.BudgetName) == budgetName {
				result = budget
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func TagOptionResourceAssociation(conn *servicecatalog.ServiceCatalog, tagOptionID, resourceID string) (*servicecatalog.ResourceDetail, error) {
	input := &servicecatalog.ListResourcesForTagOptionInput{
		TagOptionId: aws.String(tagOptionID),
	}

	var result *servicecatalog.ResourceDetail

	err := conn.ListResourcesForTagOptionPages(input, func(page *servicecatalog.ListResourcesForTagOptionOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.ResourceDetails {
			if deet == nil {
				continue
			}

			if aws.StringValue(deet.Id) == resourceID {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func PrincipalPortfolioAssociation(conn *servicecatalog.ServiceCatalog, acceptLanguage, principalARN, portfolioID string) (*servicecatalog.Principal, error) {
	input := &servicecatalog.ListPrincipalsForPortfolioInput{
		PortfolioId: aws.String(portfolioID),
	}

	if acceptLanguage != "" {
		input.AcceptLanguage = aws.String(acceptLanguage)
	}

	var result *servicecatalog.Principal

	err := conn.ListPrincipalsForPortfolioPages(input, func(page *servicecatalog.ListPrincipalsForPortfolioOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, deet := range page.Principals {
			if deet == nil {
				continue
			}

			if aws.StringValue(deet.PrincipalARN) == principalARN {
				result = deet
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
