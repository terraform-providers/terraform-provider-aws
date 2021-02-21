package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// API model constant is incorrectly AVAILABLE
	EndpointStatusAvailable = "Available"

	// API model constant is incorrectly PENDING
	EndpointStatusPending = "Pending"

	// Maximum amount of time to wait for Endpoint to return Available on creation
	EndpointStatusCreatedTimeout = 5 * time.Minute
)

// EndpointStatusCreated waits for Endpoint to return Available
func EndpointStatusCreated(conn *s3outposts.S3Outposts, endpointArn string) (*s3outposts.Endpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{EndpointStatusPending, EndpointStatusNotFound},
		Target:  []string{EndpointStatusAvailable},
		Refresh: EndpointStatus(conn, endpointArn),
		Timeout: EndpointStatusCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*s3outposts.Endpoint); ok {
		return v, err
	}

	return nil, err
}
