package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/budgets/finder"
)

func ActionStatus(conn *budgets.Budgets, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := finder.ActionById(conn, id)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, budgets.ErrCodeNotFoundException) {
				return nil, "", nil
			}
			return nil, "", err
		}

		action := out.Action
		return action, aws.StringValue(action.Status), err
	}
}
