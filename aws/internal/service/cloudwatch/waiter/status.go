package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func MetricStreamState(ctx context.Context, conn *cloudwatch.CloudWatch, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := cloudwatch.GetMetricStreamInput{
			Name: aws.String(name),
		}

		metricStream, err := conn.GetMetricStreamWithContext(ctx, &input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFoundException) {
				return nil, "", nil
			}
			return nil, "", err
		}

		return metricStream, aws.StringValue(metricStream.State), err
	}
}
