package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func ComputeEnvironmentStatus(conn *batch.Batch, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		computeEnvironmentDetail, err := finder.ComputeEnvironmentDetailByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return computeEnvironmentDetail, aws.StringValue(computeEnvironmentDetail.Status), nil
	}
}
