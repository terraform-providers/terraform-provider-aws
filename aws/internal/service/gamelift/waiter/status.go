package waiter

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func GameServerGroupState(conn *gamelift.GameLift, gameServerGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeGameServerGroup(&gamelift.DescribeGameServerGroupInput{
			GameServerGroupName: aws.String(gameServerGroupID),
		})

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.GameServerGroup == nil {
			return nil, "", nil
		}

		status := aws.StringValue(output.GameServerGroup.Status)

		if status == gamelift.GameServerGroupStatusError {
			return nil, status, errors.New(aws.StringValue(output.GameServerGroup.StatusReason))
		}

		return output.GameServerGroup, status, nil
	}
}
