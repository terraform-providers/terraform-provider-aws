package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// OperationStatus fetches the Operation and its Status
func OperationStatus(conn *servicediscovery.ServiceDiscovery, operationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &servicediscovery.GetOperationInput{
			OperationId: aws.String(operationID),
		}

		output, err := conn.GetOperation(input)

		if err != nil {
			return nil, servicediscovery.OperationStatusFail, err
		}

		// Error messages can also be contained in the response with FAIL status
		//   "ErrorCode":"CANNOT_CREATE_HOSTED_ZONE",
		//   "ErrorMessage":"The VPC that you chose, vpc-xxx in region xxx, is already associated with another private hosted zone that has an overlapping name space, xxx.. (Service: AmazonRoute53; Status Code: 400; Error Code: ConflictingDomainExists; Request ID: xxx)"
		//   "Status":"FAIL",

		if aws.StringValue(output.Operation.Status) == servicediscovery.OperationStatusFail {
			return output, servicediscovery.OperationStatusFail, fmt.Errorf("%s: %s", aws.StringValue(output.Operation.ErrorCode), aws.StringValue(output.Operation.ErrorMessage))
		}

		return output, aws.StringValue(output.Operation.Status), nil
	}
}
