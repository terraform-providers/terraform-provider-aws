package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesis/finder"
)

const (
	StreamConsumerStatusNotFound = "NotFound"
	StreamConsumerStatusUnknown  = "Unknown"
)

// StreamConsumerStatus fetches the StreamConsumer and its Status
func StreamConsumerStatus(conn *kinesis.Kinesis, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		consumer, err := finder.StreamConsumerByARN(conn, arn)

		if err != nil {
			return nil, StreamConsumerStatusUnknown, err
		}

		if consumer == nil {
			return nil, StreamConsumerStatusNotFound, nil
		}

		return consumer, aws.StringValue(consumer.ConsumerStatus), nil
	}
}
