package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	KinesisStreamingDestinationActiveTimeout   = 5 * time.Minute
	KinesisStreamingDestinationDisabledTimeout = 5 * time.Minute
	CreateTableTimeout                         = 20 * time.Minute
	UpdateTableTimeoutTotal                    = 60 * time.Minute
	ReplicaUpdateTimeout                       = 30 * time.Minute
	UpdateTableTimeout                         = 20 * time.Minute
	UpdateTableContinuousBackupsTimeout        = 20 * time.Minute
	DeleteTableTimeout                         = 10 * time.Minute
	PITRUpdateTimeout                          = 30 * time.Second
	TTLUpdateTimeout                           = 30 * time.Second
)

func DynamoDBKinesisStreamingDestinationActive(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusDisabled, dynamodb.DestinationStatusEnabling},
		Target:  []string{dynamodb.DestinationStatusActive},
		Timeout: KinesisStreamingDestinationActiveTimeout,
		Refresh: DynamoDBKinesisStreamingDestinationStatus(ctx, conn, streamArn, tableName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func DynamoDBKinesisStreamingDestinationDisabled(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusActive, dynamodb.DestinationStatusDisabling},
		Target:  []string{dynamodb.DestinationStatusDisabled},
		Timeout: KinesisStreamingDestinationDisabledTimeout,
		Refresh: DynamoDBKinesisStreamingDestinationStatus(ctx, conn, streamArn, tableName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func DynamoDBTableActive(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusCreating,
			dynamodb.TableStatusUpdating,
		},
		Target: []string{
			dynamodb.TableStatusActive,
		},
		Timeout: CreateTableTimeout,
		Refresh: DynamoDBTableStatus(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBTableDeleted(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusActive,
			dynamodb.TableStatusDeleting,
		},
		Target:  []string{},
		Timeout: DeleteTableTimeout,
		Refresh: DynamoDBTableStatus(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBReplicaActive(conn *dynamodb.DynamoDB, tableName, region string) (*dynamodb.DescribeTableOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
		},
		Target: []string{
			dynamodb.ReplicaStatusActive,
		},
		Timeout: ReplicaUpdateTimeout,
		Refresh: DynamoDBReplicaUpdate(conn, tableName, region),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.DescribeTableOutput); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBReplicaDeleted(conn *dynamodb.DynamoDB, tableName, region string) (*dynamodb.DescribeTableOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
			dynamodb.ReplicaStatusActive,
		},
		Target:  []string{""},
		Timeout: ReplicaUpdateTimeout,
		Refresh: DynamoDBReplicaDelete(conn, tableName, region),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.DescribeTableOutput); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBGSIActive(conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusCreating,
			dynamodb.IndexStatusUpdating,
		},
		Target: []string{
			dynamodb.IndexStatusActive,
		},
		Timeout: UpdateTableTimeout,
		Refresh: DynamoDBGSIStatus(conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBGSIDeleted(conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusActive,
			dynamodb.IndexStatusDeleting,
			dynamodb.IndexStatusUpdating,
		},
		Target:  []string{},
		Timeout: UpdateTableTimeout,
		Refresh: DynamoDBGSIStatus(conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBPITRUpdated(conn *dynamodb.DynamoDB, tableName string, toEnable bool) (*dynamodb.PointInTimeRecoveryDescription, error) {
	var pending []string
	target := []string{dynamodb.TimeToLiveStatusDisabled}

	if toEnable {
		pending = []string{
			"ENABLING",
		}
		target = []string{dynamodb.PointInTimeRecoveryStatusEnabled}
	}

	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Timeout: PITRUpdateTimeout,
		Refresh: DynamoDBPITRStatus(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.PointInTimeRecoveryDescription); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBTTLUpdated(conn *dynamodb.DynamoDB, tableName string, toEnable bool) (*dynamodb.TimeToLiveDescription, error) {
	pending := []string{
		dynamodb.TimeToLiveStatusEnabled,
		dynamodb.TimeToLiveStatusDisabling,
	}
	target := []string{dynamodb.TimeToLiveStatusDisabled}

	if toEnable {
		pending = []string{
			dynamodb.TimeToLiveStatusDisabled,
			dynamodb.TimeToLiveStatusEnabling,
		}
		target = []string{dynamodb.TimeToLiveStatusEnabled}
	}

	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Timeout: TTLUpdateTimeout,
		Refresh: DynamoDBTTLStatus(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TimeToLiveDescription); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBSSEUpdated(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.SSEStatusDisabling,
			dynamodb.SSEStatusEnabling,
			dynamodb.SSEStatusUpdating,
		},
		Target: []string{
			dynamodb.SSEStatusDisabled,
			dynamodb.SSEStatusEnabled,
		},
		Timeout: UpdateTableTimeout,
		Refresh: DynamoDBTableSESStatus(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}
