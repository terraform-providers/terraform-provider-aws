package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/emrcontainers/finder"
)

const (
	virtualClusterStatusNotFound = "NotFound"
	virtualClusterStatusUnknown  = "Unknown"
)

// VirtualClusterStatus fetches the virtual cluster and its status
func VirtualClusterStatus(conn *emrcontainers.EMRContainers, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vc, err := finder.VirtualClusterById(conn, id)

		if tfawserr.ErrCodeEquals(err, emrcontainers.ErrCodeResourceNotFoundException) {
			return nil, virtualClusterStatusNotFound, nil
		}

		if err != nil {
			return nil, virtualClusterStatusUnknown, err
		}

		if vc == nil {
			return nil, virtualClusterStatusNotFound, nil
		}

		return vc, aws.StringValue(vc.State), nil
	}
}
