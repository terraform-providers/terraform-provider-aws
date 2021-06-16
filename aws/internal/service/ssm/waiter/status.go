package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ssm/finder"
)

const (
	DocumentStatusUnknown = "Unknown"
)

// DocumentStatus fetches the Document and its Status
func DocumentStatus(conn *ssm.SSM, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.DocumentByName(conn, name)

		if err != nil {
			return nil, ssm.DocumentStatusFailed, err
		}

		if output == nil {
			return output, DocumentStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
