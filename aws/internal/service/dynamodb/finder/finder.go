package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func DynamoDBTableByName(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeTable(input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Table == nil {
		return nil, nil
	}

	return output.Table, nil
}

func DynamoDBGSIByTableNameIndexName(conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	table, err := DynamoDBTableByName(conn, tableName)

	if err != nil {
		return nil, err
	}

	if table == nil {
		return nil, nil
	}

	for _, gsi := range table.GlobalSecondaryIndexes {
		if aws.StringValue(gsi.IndexName) == indexName {
			return gsi, nil
		}
	}

	return nil, nil
}

func DynamoDBPITRDescriptionByTableName(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.PointInTimeRecoveryDescription, error) {
	input := &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeContinuousBackups(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	if output.ContinuousBackupsDescription == nil || output.ContinuousBackupsDescription.PointInTimeRecoveryDescription == nil {
		return nil, nil
	}

	return output.ContinuousBackupsDescription.PointInTimeRecoveryDescription, nil
}

func DynamoDBTTLRDescriptionByTableName(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TimeToLiveDescription, error) {
	input := &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeTimeToLive(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	if output.TimeToLiveDescription == nil {
		return nil, nil
	}

	return output.TimeToLiveDescription, nil
}
