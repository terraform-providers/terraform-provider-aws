package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// OperationStatusNotStarted is a OperationStatus enum value
	OperationStatusNotStarted = "NotStarted"
	// OperationStatusStarted is a OperationStatus enum value
	OperationStatusStarted = "Started"
	// OperationStatusFailed is a OperationStatus enum value
	OperationStatusFailed = "Failed"
	// OperationStatusCompleted is a OperationStatus enum value
	OperationStatusCompleted = "Completed"
	// OperationStatusSucceeded is a OperationStatus enum value
	OperationStatusSucceeded = "Succeeded"

	// OperationTimeout is the Timout Value for Operations
	OperationTimeout = 10 * time.Minute
	// OperationDelay is the Delay Value for Operations
	OperationDelay = 5 * time.Second
	// OperationMinTimeout is the MinTimout Value for Operations
	OperationMinTimeout = 3 * time.Second
)

// OperationCreated waits for an Operation to return Succeeded or Compleated
func OperationCreated(conn *lightsail.Lightsail, oid *string) (*lightsail.GetOperationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{OperationStatusStarted},
		Target:     []string{OperationStatusCompleted, OperationStatusSucceeded},
		Refresh:    LightsailOperationStatus(conn, oid),
		Timeout:    OperationTimeout,
		Delay:      OperationDelay,
		MinTimeout: OperationMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lightsail.GetOperationOutput); ok {
		return output, err
	}

	return nil, err
}
