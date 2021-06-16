package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

// StreamConsumerByARN returns the stream consumer corresponding to the specified ARN.
// Returns nil if no stream consumer is found.
func StreamConsumerByARN(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	input := &kinesis.DescribeStreamConsumerInput{
		ConsumerARN: aws.String(arn),
	}

	output, err := conn.DescribeStreamConsumer(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ConsumerDescription, nil
}
