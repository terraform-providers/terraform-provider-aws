package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/datasync/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	agentStatusReady = "ready"
)

func AgentStatus(conn *datasync.DataSync, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.AgentByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, agentStatusReady, nil
	}
}

func TaskStatus(conn *datasync.DataSync, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.TaskByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
