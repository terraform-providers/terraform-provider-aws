package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	tfevents "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/lister"
)

func Rule(conn *events.CloudWatchEvents, eventBusName, ruleName string) (*events.DescribeRuleOutput, error) {
	input := events.DescribeRuleInput{
		Name: aws.String(ruleName),
	}
	if eventBusName != "" {
		input.EventBusName = aws.String(eventBusName)
	}

	return conn.DescribeRule(&input)

}

func RuleByID(conn *events.CloudWatchEvents, ruleID string) (*events.DescribeRuleOutput, error) {
	busName, ruleName, err := tfevents.RuleParseID(ruleID)
	if err != nil {
		return nil, err
	}

	return Rule(conn, busName, ruleName)
}

func Target(conn *events.CloudWatchEvents, busName, ruleName, targetId string) (*events.Target, error) {
	var result *events.Target
	err := lister.ListAllTargetsForRulePages(conn, busName, ruleName, func(page *events.ListTargetsByRuleOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, t := range page.Targets {
			if targetId == aws.StringValue(t.Id) {
				result = t
				return false
			}
		}

		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, fmt.Errorf("CloudWatch Event Target %q (\"%s/%s\") not found", targetId, busName, ruleName)
	}
	return result, nil
}
