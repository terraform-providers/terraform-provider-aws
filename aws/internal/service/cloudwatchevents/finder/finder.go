package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfevents "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents/lister"
)

func ConnectionByName(conn *events.CloudWatchEvents, name string) (*events.DescribeConnectionOutput, error) {
	input := &events.DescribeConnectionInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeConnection(input)

	if tfawserr.ErrCodeEquals(err, events.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}

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
