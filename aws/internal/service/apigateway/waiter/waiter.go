package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Available
	StageCacheAvailableTimeout = 90 * time.Minute
	StageCacheUpdateTimeout    = 30 * time.Minute
)

// StageCacheAvailable waits for an StageCache to return Available
func StageCacheAvailable(conn *apigateway.APIGateway, restApiId, name string) (*apigateway.Stage, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			apigateway.CacheClusterStatusCreateInProgress,
			apigateway.CacheClusterStatusDeleteInProgress,
			apigateway.CacheClusterStatusFlushInProgress,
		},
		Target:  []string{apigateway.CacheClusterStatusAvailable},
		Refresh: StageCacheStatus(conn, restApiId, name),
		Timeout: StageCacheAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*apigateway.Stage); ok {
		return output, err
	}

	return nil, err
}

// StageCacheUpdated waits for an StageCache to be Updated
func StageCacheUpdated(conn *apigateway.APIGateway, restApiId, name string) (*apigateway.Stage, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			apigateway.CacheClusterStatusCreateInProgress,
			apigateway.CacheClusterStatusFlushInProgress,
		},
		Target: []string{
			apigateway.CacheClusterStatusAvailable,
			// There's an AWS API bug (raised & confirmed in Sep 2016 by support)
			// which causes the stage to remain in deletion state forever
			apigateway.CacheClusterStatusDeleteInProgress,
		},
		Refresh: StageCacheStatus(conn, restApiId, name),
		Timeout: StageCacheUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*apigateway.Stage); ok {
		return output, err
	}

	return nil, err
}
