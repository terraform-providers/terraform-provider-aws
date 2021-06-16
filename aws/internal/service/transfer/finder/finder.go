package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ServerByID(conn *transfer.Transfer, id string) (*transfer.DescribedServer, error) {
	input := &transfer.DescribeServerInput{
		ServerId: aws.String(id),
	}

	output, err := conn.DescribeServer(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Server == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Server, nil
}

func UserByID(conn *transfer.Transfer, serverId, userName string) (*transfer.DescribeUserOutput, error) {
	input := &transfer.DescribeUserInput{
		ServerId: aws.String(serverId),
		UserName: aws.String(userName),
	}

	output, err := conn.DescribeUser(input)

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.User == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
