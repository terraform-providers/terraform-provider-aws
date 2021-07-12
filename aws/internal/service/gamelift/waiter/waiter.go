package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func GameServerGroupActive(conn *gamelift.GameLift, gameServerGroupID string, timeout time.Duration) (*gamelift.GameServerGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			gamelift.GameServerGroupStatusNew,
			gamelift.GameServerGroupStatusActivating,
		},
		Target: []string{
			gamelift.GameServerGroupStatusActive,
		},
		Refresh: GameServerGroupState(conn, gameServerGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*gamelift.GameServerGroup); ok {
		return v, err
	}

	return nil, err
}

func GameServerGroupDeleted(conn *gamelift.GameLift, gameServerGroupID string, timeout time.Duration) (*gamelift.GameServerGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			gamelift.GameServerGroupStatusDeleteScheduled,
			gamelift.GameServerGroupStatusDeleting,
			gamelift.GameServerGroupStatusDeleted,
		},
		Target:  []string{},
		Refresh: GameServerGroupState(conn, gameServerGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*gamelift.GameServerGroup); ok {
		return v, err
	}

	return nil, err
}
