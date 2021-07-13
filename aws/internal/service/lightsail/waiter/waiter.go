package waiter

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	OperationTimeout = 20 * time.Minute
	// OperationDelay is the Delay Value for Operations
	OperationDelay = 5 * time.Second
	// OperationMinTimeout is the MinTimout Value for Operations
	OperationMinTimeout = 3 * time.Second

	// DatabaseStateModifying is a state value for a Relational Database undergoing a modification
	DatabaseStateModifying = "modifying"
	// DatabaseStateAvailable is a state value for a Relational Database available for modification
	DatabaseStateAvailable = "available"

	// DatabaseTimeout is the Timout Value for Relational Database Modifications
	DatabaseTimeout = 20 * time.Minute
	// DatabaseDelay is the Delay Value for Relational Database Modifications
	DatabaseDelay = 5 * time.Second
	// DatabaseMinTimeout is the MinTimout Value for Relational Database Modifications
	DatabaseMinTimeout = 3 * time.Second
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

func DatabaseModified(conn *lightsail.Lightsail, db *string) (*lightsail.GetRelationalDatabaseOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{DatabaseStateModifying},
		Target:     []string{DatabaseStateAvailable},
		Refresh:    LightsailDatabaseStatus(conn, db),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return output, err
	}

	return nil, err
}

func DatabaseBackupRetentionModified(conn *lightsail.Lightsail, db *string, status *bool) (*lightsail.GetRelationalDatabaseOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{strconv.FormatBool(!aws.BoolValue(status))},
		Target:     []string{strconv.FormatBool(aws.BoolValue(status))},
		Refresh:    LightsailDatabaseBackupRetentionStatus(conn, db, status),
		Timeout:    DatabaseTimeout,
		Delay:      DatabaseDelay,
		MinTimeout: DatabaseMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lightsail.GetRelationalDatabaseOutput); ok {
		return output, err
	}

	return nil, err
}
