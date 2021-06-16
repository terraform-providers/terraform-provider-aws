package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	StreamConsumerCreatedTimeout = 5 * time.Minute
	StreamConsumerDeletedTimeout = 5 * time.Minute
)

// StreamConsumerCreated waits for an Stream Consumer to return Active
func StreamConsumerCreated(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusCreating},
		Target:  []string{kinesis.ConsumerStatusActive},
		Refresh: StreamConsumerStatus(conn, arn),
		Timeout: StreamConsumerCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return v, err
	}

	return nil, err
}

// StreamConsumerDeleted waits for a Stream Consumer to be deleted
func StreamConsumerDeleted(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kinesis.ConsumerStatusDeleting},
		Target:  []string{},
		Refresh: StreamConsumerStatus(conn, arn),
		Timeout: StreamConsumerDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*kinesis.ConsumerDescription); ok {
		return v, err
	}

	return nil, err
}
