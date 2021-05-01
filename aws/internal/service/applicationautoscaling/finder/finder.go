package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ScheduledAction(conn *applicationautoscaling.ApplicationAutoScaling, name, serviceNamespace, resourceId string) (*applicationautoscaling.ScheduledAction, error) {
	var result *applicationautoscaling.ScheduledAction

	input := &applicationautoscaling.DescribeScheduledActionsInput{
		ScheduledActionNames: []*string{aws.String(name)},
		ServiceNamespace:     aws.String(serviceNamespace),
		ResourceId:           aws.String(resourceId),
	}
	err := conn.DescribeScheduledActionsPages(input, func(page *applicationautoscaling.DescribeScheduledActionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.ScheduledActions {
			if item == nil {
				continue
			}

			if name == aws.StringValue(item.ScheduledActionName) {
				result = item
				return false
			}
		}

		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return result, nil
}
