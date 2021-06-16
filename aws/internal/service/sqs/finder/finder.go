package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func QueueAttributesByURL(conn *sqs.SQS, url string) (map[string]string, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{sqs.QueueAttributeNameAll}),
		QueueUrl:       aws.String(url),
	}

	output, err := conn.GetQueueAttributes(input)

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Attributes == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return aws.StringValueMap(output.Attributes), nil
}

func QueuePolicyByURL(conn *sqs.SQS, url string) (string, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{sqs.QueueAttributeNamePolicy}),
		QueueUrl:       aws.String(url),
	}

	output, err := conn.GetQueueAttributes(input)

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return "", &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Attributes == nil {
		return "", &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	v, ok := output.Attributes[sqs.QueueAttributeNamePolicy]

	if !ok || aws.StringValue(v) == "" {
		return "", &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return aws.StringValue(v), nil
}
