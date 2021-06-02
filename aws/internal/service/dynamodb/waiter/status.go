package waiter

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/dynamodb/finder"
)

func DynamoDBKinesisStreamingDestinationStatus(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := finder.DynamoDBKinesisDataStreamDestination(ctx, conn, streamArn, tableName)

		if err != nil {
			return nil, "", err
		}

		if result == nil {
			return nil, "", nil
		}

		return result, aws.StringValue(result.DestinationStatus), nil
	}
}

func DynamoDBTableStatus(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := finder.DynamoDBTableByName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if table == nil {
			return nil, "", nil
		}

		return table, aws.StringValue(table.TableStatus), nil
	}
}

func DynamoDBReplicaUpdate(conn *dynamodb.DynamoDB, tableName, region string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return 42, "", err
		}
		log.Printf("[DEBUG] DynamoDB replicas: %s", result.Table.Replicas)

		var targetReplica *dynamodb.ReplicaDescription

		for _, replica := range result.Table.Replicas {
			if aws.StringValue(replica.RegionName) == region {
				targetReplica = replica
				break
			}
		}

		if targetReplica == nil {
			return result, dynamodb.ReplicaStatusCreating, nil
		}

		return result, aws.StringValue(targetReplica.ReplicaStatus), nil
	}
}

func DynamoDBReplicaDelete(conn *dynamodb.DynamoDB, tableName, region string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return 42, "", err
		}

		log.Printf("[DEBUG] all replicas for waiting: %s", result.Table.Replicas)
		var targetReplica *dynamodb.ReplicaDescription

		for _, replica := range result.Table.Replicas {
			if aws.StringValue(replica.RegionName) == region {
				targetReplica = replica
				break
			}
		}

		if targetReplica == nil {
			return result, "", nil
		}

		return result, aws.StringValue(targetReplica.ReplicaStatus), nil
	}
}

func DynamoDBGSIStatus(conn *dynamodb.DynamoDB, tableName, indexName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		gsi, err := finder.DynamoDBGSIByTableNameIndexName(conn, tableName, indexName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if gsi == nil {
			return nil, "", nil
		}

		return gsi, aws.StringValue(gsi.IndexStatus), nil
	}
}

func DynamoDBPITRStatus(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pitr, err := finder.DynamoDBPITRDescriptionByTableName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if pitr == nil {
			return nil, "", nil
		}

		return pitr, aws.StringValue(pitr.PointInTimeRecoveryStatus), nil
	}
}

func DynamoDBTTLStatus(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ttl, err := finder.DynamoDBTTLRDescriptionByTableName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if ttl == nil {
			return nil, "", nil
		}

		return ttl, aws.StringValue(ttl.TimeToLiveStatus), nil
	}
}

func DynamoDBTableSESStatus(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := finder.DynamoDBTableByName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if table == nil {
			return nil, "", nil
		}

		// Disabling SSE returns null SSEDescription
		if table.SSEDescription == nil {
			return table, dynamodb.SSEStatusDisabled, nil
		}

		return table, aws.StringValue(table.SSEDescription.Status), nil
	}
}
