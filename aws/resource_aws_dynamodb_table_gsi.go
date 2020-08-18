package aws

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsDynamoDbTableGsi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTableGsiCreate,
		Read:   resourceAwsDynamoDbTableGsiRead,
		Update: resourceAwsDynamoDbTableGsiUpdate,
		Delete: resourceAwsDynamoDbTableGsiDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute), // provisioned throughput changes only
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hash_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"range_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"attribute": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								dynamodb.ScalarAttributeTypeB,
								dynamodb.ScalarAttributeTypeN,
								dynamodb.ScalarAttributeTypeS,
							}, false),
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
					return hashcode.String(buf.String())
				},
			},
			"write_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"read_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"projection_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"non_key_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsDynamoDbTableGsiCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	indexName := d.Get("name").(string)
	keySchemaMap := map[string]interface{}{
		"hash_key": d.Get("hash_key").(string),
	}
	if v, ok := d.GetOk("range_key"); ok {
		keySchemaMap["range_key"] = v.(string)
	}

	log.Printf("[DEBUG] Creating DynamoDB table index with key schema: %#v", keySchemaMap)
	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
	}

	// TODO: add types to key attributes instead
	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandDynamoDbAttributes(aSet.List())
	}

	projection := &dynamodb.Projection{
		ProjectionType: aws.String(d.Get("projection_type").(string)),
	}

	if v, ok := d.GetOk("non_key_attributes"); ok && len(v.([]interface{})) > 0 {
		projection.NonKeyAttributes = expandStringList(v.([]interface{}))
	}

	createOp := &dynamodb.GlobalSecondaryIndexUpdate{
		Create: &dynamodb.CreateGlobalSecondaryIndexAction{
			IndexName: aws.String(indexName),
			KeySchema: expandDynamoDbKeySchema(keySchemaMap),
			ProvisionedThroughput: expandDynamoDbProvisionedThroughput(map[string]interface{}{
				"read_capacity":  d.Get("read_capacity"),
				"write_capacity": d.Get("write_capacity"),
			}),
			Projection: projection,
		},
	}
	req.GlobalSecondaryIndexUpdates = []*dynamodb.GlobalSecondaryIndexUpdate{createOp}

	var output *dynamodb.UpdateTableOutput
	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		var err error
		output, err = conn.UpdateTable(req)
		if err != nil {
			if isAWSErr(err, "ThrottlingException", "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			// Subscriber limit exceeded: Only 1 online index can be created or deleted simultaneously per table
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	gsiDescription := findDynamoDbGsi(&output.TableDescription.GlobalSecondaryIndexes, indexName)
	d.SetId(*gsiDescription.IndexName)

	err = waitForDynamoDbGSIToBeActive(d.Get("table_name").(string), d.Id(), conn)
	return err
}

func resourceAwsDynamoDbTableGsiUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Get("table_name").(string)),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
			{
				Update: &dynamodb.UpdateGlobalSecondaryIndexAction{
					IndexName: aws.String(d.Id()),
					ProvisionedThroughput: expandDynamoDbProvisionedThroughput(map[string]interface{}{
						"read_capacity":  d.Get("read_capacity"),
						"write_capacity": d.Get("write_capacity"),
					}),
				},
			},
		},
	}

	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandDynamoDbAttributes(aSet.List())
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		var err error
		_, err = conn.UpdateTable(req)
		if err != nil {
			if isAWSErr(err, "ThrottlingException", "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			// Subscriber limit exceeded: Only 1 online index can be created or deleted simultaneously per table
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = waitForDynamoDbGSIToBeActive(d.Get("table_name").(string), d.Id(), conn)
	return err
}

func resourceAwsDynamoDbTableGsiRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn
	tableName := d.Get("table_name").(string)

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return err
	}

	err = flattenAwsDynamoDbTableGsiResource(d, result.Table)
	return err
}

func resourceAwsDynamoDbTableGsiDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	req := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Get("table_name").(string)),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
			{
				Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
					IndexName: aws.String(d.Id()),
				},
			},
		},
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.UpdateTable(req)
		if err != nil {
			// Subscriber limit exceeded: Only 1 online index can be created or deleted simultaneously per table
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = waitForDynamoDbGSIToBeDeleted(d.Get("table_name").(string), d.Id(), conn)
	return err
}

// Helpers

func flattenAwsDynamoDbTableGsiResource(d *schema.ResourceData, table *dynamodb.TableDescription) error {
	d.Set("table_name", table.TableName)

	gsi := findDynamoDbGsi(&table.GlobalSecondaryIndexes, d.Id())
	d.Set("write_capacity", gsi.ProvisionedThroughput.WriteCapacityUnits)
	d.Set("read_capacity", gsi.ProvisionedThroughput.ReadCapacityUnits)
	d.Set("projection_type", gsi.Projection.ProjectionType)

	gsiAttributeNames := make(map[string]struct{}, len(gsi.KeySchema))
	for _, attribute := range gsi.KeySchema {
		if *attribute.KeyType == dynamodb.KeyTypeHash {
			d.Set("hash_key", attribute.AttributeName)
			gsiAttributeNames[*attribute.AttributeName] = struct{}{}
		}

		if *attribute.KeyType == dynamodb.KeyTypeRange {
			d.Set("range_key", attribute.AttributeName)
			gsiAttributeNames[*attribute.AttributeName] = struct{}{}
		}
	}

	attributes := []interface{}{}
	for _, attrdef := range table.AttributeDefinitions {
		if _, ok := gsiAttributeNames[*attrdef.AttributeName]; ok {
			attribute := map[string]string{
				"name": *attrdef.AttributeName,
				"type": *attrdef.AttributeType,
			}
			attributes = append(attributes, attribute)
		}
	}
	d.Set("attribute", attributes)

	nonKeyAttrs := make([]string, 0, len(gsi.Projection.NonKeyAttributes))
	for _, nonKeyAttr := range gsi.Projection.NonKeyAttributes {
		nonKeyAttrs = append(nonKeyAttrs, *nonKeyAttr)
	}
	d.Set("non_key_attributes", nonKeyAttrs)

	return nil
}

func findDynamoDbGsi(gsiList *[]*dynamodb.GlobalSecondaryIndexDescription, target string) *dynamodb.GlobalSecondaryIndexDescription {
	for _, gsiObject := range *gsiList {
		if *gsiObject.IndexName == target {
			return gsiObject
		}
	}

	return nil
}
