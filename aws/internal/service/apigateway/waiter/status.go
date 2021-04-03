package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/apigateway/finder"
)

const (
	StageCacheStatusUnknown = "Unknown"
)

// StageCacheStatus fetches the StageCache and its Status
func StageCacheStatus(conn *apigateway.APIGateway, restApiId, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.StageByName(conn, restApiId, name)

		if err != nil {
			return nil, StageCacheStatusUnknown, err
		}

		if output == nil {
			return output, StageCacheStatusUnknown, nil
		}

		return output, aws.StringValue(output.CacheClusterStatus), nil
	}
}
