package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/dynamodb/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/dynamodb/waiter"
)

func resourceAwsDynamoDbKinesisStreamingDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsDynamoDbKinesisStreamingDestinationCreate,
		ReadWithoutTimeout:   resourceAwsDynamoDbKinesisStreamingDestinationRead,
		DeleteWithoutTimeout: resourceAwsDynamoDbKinesisStreamingDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"stream_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDynamoDbKinesisStreamingDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dynamodbconn

	streamArn := d.Get("stream_arn").(string)
	tableName := d.Get("table_name").(string)

	input := &dynamodb.EnableKinesisStreamingDestinationInput{
		StreamArn: aws.String(streamArn),
		TableName: aws.String(tableName),
	}

	output, err := conn.EnableKinesisStreamingDestinationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error enabling DynamoDB Kinesis streaming destination (stream: %s, table: %s): %w", streamArn, tableName, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error enabling DynamoDB Kinesis streaming destination (stream: %s, table: %s): empty output", streamArn, tableName))
	}

	if err := waiter.DynamoDBKinesisStreamingDestinationActive(ctx, conn, streamArn, tableName); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DynamoDB Kinesis streaming destination (stream: %s, table: %s) to be active: %w", streamArn, tableName, err))
	}

	d.SetId(fmt.Sprintf("%s,%s", aws.StringValue(output.TableName), aws.StringValue(output.StreamArn)))

	return resourceAwsDynamoDbKinesisStreamingDestinationRead(ctx, d, meta)
}

func resourceAwsDynamoDbKinesisStreamingDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dynamodbconn

	tableName, streamArn, err := dynamoDbKinesisStreamingDestinationParseId(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	output, err := finder.DynamoDBKinesisDataStreamDestination(ctx, conn, streamArn, tableName)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] DynamoDB Kinesis Streaming Destination (stream: %s, table: %s) not found, removing from state", streamArn, tableName)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving DynamoDB Kinesis streaming destination (stream: %s, table: %s): %w", streamArn, tableName, err))
	}

	if output == nil || aws.StringValue(output.DestinationStatus) == dynamodb.DestinationStatusDisabled {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error retrieving DynamoDB Kinesis streaming destination (stream: %s, table: %s): empty output after creation", streamArn, tableName))
		}
		log.Printf("[WARN] DynamoDB Kinesis Streaming Destination (stream: %s, table: %s) not found, removing from state", streamArn, tableName)
		d.SetId("")
		return nil
	}

	d.Set("stream_arn", output.StreamArn)
	d.Set("table_name", tableName)

	return nil
}

func resourceAwsDynamoDbKinesisStreamingDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).dynamodbconn

	tableName, streamArn, err := dynamoDbKinesisStreamingDestinationParseId(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &dynamodb.DisableKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
		StreamArn: aws.String(streamArn),
	}

	_, err = conn.DisableKinesisStreamingDestinationWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error disabling DynamoDB Kinesis streaming destination (stream: %s, table: %s): %w", streamArn, tableName, err))
	}

	if err := waiter.DynamoDBKinesisStreamingDestinationDisabled(ctx, conn, streamArn, tableName); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for DynamoDB Kinesis streaming destination (stream: %s, table: %s) to be disabled: %w", streamArn, tableName, err))
	}

	return nil
}

func dynamoDbKinesisStreamingDestinationParseId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected TABLE_NAME,STREAM_ARN", id)
	}

	return parts[0], parts[1], nil
}
