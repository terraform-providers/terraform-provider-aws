package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	tfbudgets "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets"
)

func ActionById(conn *budgets.Budgets, id string) (*budgets.DescribeBudgetActionOutput, error) {
	accountID, actionID, budgetName, err := tfbudgets.DecodeBudgetsBudgetActionID(id)
	if err != nil {
		return nil, err
	}

	input := &budgets.DescribeBudgetActionInput{
		BudgetName: aws.String(budgetName),
		AccountId:  aws.String(accountID),
		ActionId:   aws.String(actionID),
	}

	out, err := conn.DescribeBudgetAction(input)
	if err != nil {
		return nil, err
	}

	return out, nil
}
