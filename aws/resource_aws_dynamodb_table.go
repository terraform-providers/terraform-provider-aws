package aws

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDynamoDbTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTableCreate,
		Read:   resourceAwsDynamoDbTableRead,
		Update: resourceAwsDynamoDbTableUpdate,
		Delete: resourceAwsDynamoDbTableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			func(diff *schema.ResourceDiff, v interface{}) error {
				return validateDynamoDbStreamSpec(diff)
			},
			func(diff *schema.ResourceDiff, v interface{}) error {
				return validateDynamoDbTableAttributes(diff)
			},
			func(diff *schema.ResourceDiff, v interface{}) error {
				if diff.Id() != "" && diff.HasChange("server_side_encryption") {
					o, n := diff.GetChange("server_side_encryption")
					if isDynamoDbTableOptionDisabled(o) && isDynamoDbTableOptionDisabled(n) {
						return diff.Clear("server_side_encryption")
					}
				}
				return nil
			},
			func(diff *schema.ResourceDiff, v interface{}) error {
				if diff.Id() != "" && diff.HasChange("point_in_time_recovery") {
					o, n := diff.GetChange("point_in_time_recovery")
					if isDynamoDbTableOptionDisabled(o) && isDynamoDbTableOptionDisabled(n) {
						return diff.Clear("point_in_time_recovery")
					}
				}
				return nil
			},
		),

		SchemaVersion: 1,
		MigrateState:  resourceAwsDynamoDbTableMigrateState,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
			"billing_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  dynamodb.BillingModeProvisioned,
				ValidateFunc: validation.StringInSlice([]string{
					dynamodb.BillingModePayPerRequest,
					dynamodb.BillingModeProvisioned,
				}, false),
			},
			"write_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"read_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
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
			"ttl": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
			},
			"local_secondary_index": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"range_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"projection_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"non_key_attributes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
			"global_secondary_index": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"write_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"read_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"hash_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"range_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"projection_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"non_key_attributes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"stream_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"stream_view_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				StateFunc: func(v interface{}) string {
					value := v.(string)
					return strings.ToUpper(value)
				},
				ValidateFunc: validation.StringInSlice([]string{
					"",
					dynamodb.StreamViewTypeNewImage,
					dynamodb.StreamViewTypeOldImage,
					dynamodb.StreamViewTypeNewAndOldImages,
					dynamodb.StreamViewTypeKeysOnly,
				}, false),
			},
			"stream_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stream_label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_side_encryption": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"kms_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"tags": tagsSchema(),
			"point_in_time_recovery": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsDynamoDbTableCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	keySchemaMap := map[string]interface{}{
		"hash_key": d.Get("hash_key").(string),
	}
	if v, ok := d.GetOk("range_key"); ok {
		keySchemaMap["range_key"] = v.(string)
	}

	log.Printf("[DEBUG] Creating DynamoDB table with key schema: %#v", keySchemaMap)

	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DynamodbTags()

	req := &dynamodb.CreateTableInput{
		TableName:   aws.String(d.Get("name").(string)),
		BillingMode: aws.String(d.Get("billing_mode").(string)),
		KeySchema:   expandDynamoDbKeySchema(keySchemaMap),
		Tags:        tags,
	}

	billingMode := d.Get("billing_mode").(string)
	capacityMap := map[string]interface{}{
		"write_capacity": d.Get("write_capacity"),
		"read_capacity":  d.Get("read_capacity"),
	}

	if err := validateDynamoDbProvisionedThroughput(capacityMap, billingMode); err != nil {
		return err
	}

	req.ProvisionedThroughput = expandDynamoDbProvisionedThroughput(capacityMap, billingMode)

	if v, ok := d.GetOk("attribute"); ok {
		aSet := v.(*schema.Set)
		req.AttributeDefinitions = expandDynamoDbAttributes(aSet.List())
	}

	if v, ok := d.GetOk("local_secondary_index"); ok {
		lsiSet := v.(*schema.Set)
		req.LocalSecondaryIndexes = expandDynamoDbLocalSecondaryIndexes(lsiSet.List(), keySchemaMap)
	}

	if v, ok := d.GetOk("global_secondary_index"); ok {
		globalSecondaryIndexes := []*dynamodb.GlobalSecondaryIndex{}
		gsiSet := v.(*schema.Set)

		for _, gsiObject := range gsiSet.List() {
			gsi := gsiObject.(map[string]interface{})
			if err := validateDynamoDbProvisionedThroughput(gsi, billingMode); err != nil {
				return fmt.Errorf("Failed to create GSI: %v", err)
			}

			gsiObject := expandDynamoDbGlobalSecondaryIndex(gsi, billingMode)
			globalSecondaryIndexes = append(globalSecondaryIndexes, gsiObject)
		}
		req.GlobalSecondaryIndexes = globalSecondaryIndexes
	}

	if v, ok := d.GetOk("stream_enabled"); ok {
		req.StreamSpecification = &dynamodb.StreamSpecification{
			StreamEnabled:  aws.Bool(v.(bool)),
			StreamViewType: aws.String(d.Get("stream_view_type").(string)),
		}
	}

	if v, ok := d.GetOk("server_side_encryption"); ok {
		req.SSESpecification = expandDynamoDbEncryptAtRestOptions(v.([]interface{}))
	}

	var output *dynamodb.CreateTableOutput
	var requiresTagging bool
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.CreateTable(req)
		if err != nil {
			if isAWSErr(err, "ThrottlingException", "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "can be created, updated, or deleted simultaneously") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "indexed tables that can be created simultaneously") {
				return resource.RetryableError(err)
			}
			// AWS GovCloud (US) and others may reply with the following until their API is updated:
			// ValidationException: One or more parameter values were invalid: Unsupported input parameter BillingMode
			if isAWSErr(err, "ValidationException", "Unsupported input parameter BillingMode") {
				req.BillingMode = nil
				return resource.RetryableError(err)
			}
			// AWS GovCloud (US) and others may reply with the following until their API is updated:
			// ValidationException: Unsupported input parameter Tags
			if isAWSErr(err, "ValidationException", "Unsupported input parameter Tags") {
				req.Tags = nil
				requiresTagging = true
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = conn.CreateTable(req)
	}
	if err != nil {
		return fmt.Errorf("error creating DynamoDB Table: %s", err)
	}

	d.SetId(*output.TableDescription.TableName)
	d.Set("arn", output.TableDescription.TableArn)

	if err := waitForDynamoDbTableToBeActive(d.Id(), d.Timeout(schema.TimeoutCreate), conn); err != nil {
		return err
	}

	if requiresTagging {
		if err := keyvaluetags.DynamodbUpdateTags(conn, d.Get("arn").(string), nil, tags); err != nil {
			return fmt.Errorf("error adding DynamoDB Table (%s) tags: %s", d.Id(), err)
		}
	}

	if d.Get("ttl.0.enabled").(bool) {
		if err := updateDynamoDbTimeToLive(d.Id(), d.Get("ttl").([]interface{}), conn); err != nil {
			return fmt.Errorf("error enabling DynamoDB Table (%s) Time to Live: %s", d.Id(), err)
		}
	}

	if d.Get("point_in_time_recovery.0.enabled").(bool) {
		if err := updateDynamoDbPITR(d, conn); err != nil {
			return fmt.Errorf("error enabling DynamoDB Table (%s) point in time recovery: %s", d.Id(), err)
		}
	}

	return resourceAwsDynamoDbTableRead(d, meta)
}

func resourceAwsDynamoDbTableUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn
	billingMode := d.Get("billing_mode").(string)

	// Global Secondary Index operations must occur in multiple phases
	// to prevent various error scenarios. If there are no detected required
	// updates in the Terraform configuration, later validation or API errors
	// will signal the problems.
	var gsiUpdates []*dynamodb.GlobalSecondaryIndexUpdate

	if d.HasChange("global_secondary_index") {
		var err error
		o, n := d.GetChange("global_secondary_index")
		gsiUpdates, err = diffDynamoDbGSI(o.(*schema.Set).List(), n.(*schema.Set).List(), billingMode)

		if err != nil {
			return fmt.Errorf("computing difference for DynamoDB Table (%s) Global Secondary Index updates failed: %s", d.Id(), err)
		}

		log.Printf("[DEBUG] Computed DynamoDB Table (%s) Global Secondary Index updates: %s", d.Id(), gsiUpdates)
	}

	// Phase 1 of Global Secondary Index Operations: Delete Only
	//  * Delete indexes first to prevent error when simultaneously updating
	//    BillingMode to PROVISIONED, which requires updating index
	//    ProvisionedThroughput first, but we have no definition
	//  * Only 1 online index can be deleted simultaneously per table
	for _, gsiUpdate := range gsiUpdates {
		if gsiUpdate.Delete == nil {
			continue
		}

		idxName := aws.StringValue(gsiUpdate.Delete.IndexName)
		input := &dynamodb.UpdateTableInput{
			GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{gsiUpdate},
			TableName:                   aws.String(d.Id()),
		}

		if _, err := conn.UpdateTable(input); err != nil {
			return fmt.Errorf("error deleting DynamoDB Table (%s) Global Secondary Index (%s): %s", d.Id(), idxName, err)
		}

		if err := waitForDynamoDbGSIToBeDeleted(d.Id(), idxName, d.Timeout(schema.TimeoutUpdate), conn); err != nil {
			return fmt.Errorf("error waiting for DynamoDB Table (%s) Global Secondary Index (%s) deletion: %s", d.Id(), idxName, err)
		}
	}

	hasTableUpdate := false
	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(d.Id()),
	}

	if d.HasChange("billing_mode") || d.HasChange("read_capacity") || d.HasChange("write_capacity") {
		hasTableUpdate = true

		capacityMap := map[string]interface{}{
			"write_capacity": d.Get("write_capacity"),
			"read_capacity":  d.Get("read_capacity"),
		}

		if err := validateDynamoDbProvisionedThroughput(capacityMap, billingMode); err != nil {
			return err
		}

		input.BillingMode = aws.String(billingMode)
		input.ProvisionedThroughput = expandDynamoDbProvisionedThroughput(capacityMap, billingMode)
	}

	if d.HasChange("stream_enabled") || d.HasChange("stream_view_type") {
		hasTableUpdate = true

		input.StreamSpecification = &dynamodb.StreamSpecification{
			StreamEnabled: aws.Bool(d.Get("stream_enabled").(bool)),
		}
		if d.Get("stream_enabled").(bool) {
			input.StreamSpecification.StreamViewType = aws.String(d.Get("stream_view_type").(string))
		}
	}

	// Phase 2 of Global Secondary Index Operations: Update Only
	// Cannot create or delete index while updating table ProvisionedThroughput
	// Must skip all index updates when switching BillingMode from PROVISIONED to PAY_PER_REQUEST
	// Must update all indexes when switching BillingMode from PAY_PER_REQUEST to PROVISIONED
	if billingMode == dynamodb.BillingModeProvisioned {
		for _, gsiUpdate := range gsiUpdates {
			if gsiUpdate.Update == nil {
				continue
			}

			input.GlobalSecondaryIndexUpdates = append(input.GlobalSecondaryIndexUpdates, gsiUpdate)
		}
	}

	if hasTableUpdate {
		log.Printf("[DEBUG] Updating DynamoDB Table: %s", input)
		_, err := conn.UpdateTable(input)

		if err != nil {
			return fmt.Errorf("error updating DynamoDB Table (%s): %s", d.Id(), err)
		}

		if err := waitForDynamoDbTableToBeActive(d.Id(), d.Timeout(schema.TimeoutUpdate), conn); err != nil {
			return fmt.Errorf("error waiting for DynamoDB Table (%s) update: %s", d.Id(), err)
		}

		for _, gsiUpdate := range gsiUpdates {
			if gsiUpdate.Update == nil {
				continue
			}

			idxName := aws.StringValue(gsiUpdate.Update.IndexName)
			if err := waitForDynamoDbGSIToBeActive(d.Id(), idxName, d.Timeout(schema.TimeoutUpdate), conn); err != nil {
				return fmt.Errorf("error waiting for DynamoDB Table (%s) Global Secondary Index (%s) update: %s", d.Id(), idxName, err)
			}
		}
	}

	// Phase 3 of Global Secondary Index Operations: Create Only
	// Only 1 online index can be created simultaneously per table
	for _, gsiUpdate := range gsiUpdates {
		if gsiUpdate.Create == nil {
			continue
		}

		idxName := aws.StringValue(gsiUpdate.Create.IndexName)
		input := &dynamodb.UpdateTableInput{
			AttributeDefinitions:        expandDynamoDbAttributes(d.Get("attribute").(*schema.Set).List()),
			GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{gsiUpdate},
			TableName:                   aws.String(d.Id()),
		}

		if _, err := conn.UpdateTable(input); err != nil {
			return fmt.Errorf("error creating DynamoDB Table (%s) Global Secondary Index (%s): %s", d.Id(), idxName, err)
		}

		if err := waitForDynamoDbGSIToBeActive(d.Id(), idxName, d.Timeout(schema.TimeoutUpdate), conn); err != nil {
			return fmt.Errorf("error waiting for DynamoDB Table (%s) Global Secondary Index (%s) creation: %s", d.Id(), idxName, err)
		}
	}

	if d.HasChange("server_side_encryption") {
		// "ValidationException: One or more parameter values were invalid: Server-Side Encryption modification must be the only operation in the request".
		_, err := conn.UpdateTable(&dynamodb.UpdateTableInput{
			TableName:        aws.String(d.Id()),
			SSESpecification: expandDynamoDbEncryptAtRestOptions(d.Get("server_side_encryption").([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("error updating DynamoDB Table (%s) SSE: %s", d.Id(), err)
		}

		if err := waitForDynamoDbSSEUpdateToBeCompleted(d.Id(), d.Timeout(schema.TimeoutUpdate), conn); err != nil {
			return fmt.Errorf("error waiting for DynamoDB Table (%s) SSE update: %s", d.Id(), err)
		}
	}

	if d.HasChange("ttl") {
		if err := updateDynamoDbTimeToLive(d.Id(), d.Get("ttl").([]interface{}), conn); err != nil {
			return fmt.Errorf("error updating DynamoDB Table (%s) time to live: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.DynamodbUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating DynamoDB Table (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("point_in_time_recovery") {
		if err := updateDynamoDbPITR(d, conn); err != nil {
			return fmt.Errorf("error updating DynamoDB Table (%s) point in time recovery: %s", d.Id(), err)
		}
	}

	return resourceAwsDynamoDbTableRead(d, meta)
}

func resourceAwsDynamoDbTableRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(d.Id()),
	})

	if err != nil {
		if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Dynamodb Table (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	err = flattenAwsDynamoDbTableResource(d, result.Table)
	if err != nil {
		return err
	}

	ttlOut, err := conn.DescribeTimeToLive(&dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error describing DynamoDB Table (%s) Time to Live: %s", d.Id(), err)
	}
	if err := d.Set("ttl", flattenDynamoDbTtl(ttlOut)); err != nil {
		return fmt.Errorf("error setting ttl: %s", err)
	}

	tags, err := readDynamoDbTableTags(d.Get("arn").(string), conn)
	if err != nil {
		return err
	}

	d.Set("tags", tags)

	pitrOut, err := conn.DescribeContinuousBackups(&dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(d.Id()),
	})
	if err != nil && !isAWSErr(err, "UnknownOperationException", "") {
		return err
	}
	d.Set("point_in_time_recovery", flattenDynamoDbPitr(pitrOut))

	return nil
}

func resourceAwsDynamoDbTableDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	log.Printf("[DEBUG] DynamoDB delete table: %s", d.Id())

	err := deleteAwsDynamoDbTable(d.Id(), conn)
	if err != nil {
		if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "Requested resource not found: Table: ") {
			return nil
		}
		return fmt.Errorf("error deleting DynamoDB Table (%s): %s", d.Id(), err)
	}

	if err := waitForDynamodbTableDeletion(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for DynamoDB Table (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func deleteAwsDynamoDbTable(tableName string, conn *dynamodb.DynamoDB) error {
	input := &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteTable(input)
		if err != nil {
			// Subscriber limit exceeded: Only 10 tables can be created, updated, or deleted simultaneously
			if isAWSErr(err, dynamodb.ErrCodeLimitExceededException, "simultaneously") {
				return resource.RetryableError(err)
			}
			// This handles multiple scenarios in the DynamoDB API:
			// 1. Updating a table immediately before deletion may return:
			//    ResourceInUseException: Attempt to change a resource which is still in use: Table is being updated:
			// 2. Removing a table from a DynamoDB global table may return:
			//    ResourceInUseException: Attempt to change a resource which is still in use: Table is being deleted:
			if isAWSErr(err, dynamodb.ErrCodeResourceInUseException, "") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "Requested resource not found: Table: ") {
				return resource.NonRetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteTable(input)
	}

	return err
}

func waitForDynamodbTableDeletion(conn *dynamodb.DynamoDB, tableName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusActive,
			dynamodb.TableStatusDeleting,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			input := &dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			}

			output, err := conn.DescribeTable(input)

			if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
				return nil, "", nil
			}

			if err != nil {
				return 42, "", err
			}

			if output == nil {
				return nil, "", nil
			}

			return output.Table, aws.StringValue(output.Table.TableStatus), nil
		},
	}

	_, err := stateConf.WaitForState()

	return err
}

func updateDynamoDbTimeToLive(tableName string, ttlList []interface{}, conn *dynamodb.DynamoDB) error {
	ttlMap := ttlList[0].(map[string]interface{})

	input := &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(tableName),
		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
			AttributeName: aws.String(ttlMap["attribute_name"].(string)),
			Enabled:       aws.Bool(ttlMap["enabled"].(bool)),
		},
	}

	log.Printf("[DEBUG] Updating DynamoDB Table (%s) Time To Live: %s", tableName, input)
	if _, err := conn.UpdateTimeToLive(input); err != nil {
		return fmt.Errorf("error updating DynamoDB Table (%s) Time To Live: %s", tableName, err)
	}

	log.Printf("[DEBUG] Waiting for DynamoDB Table (%s) Time to Live update to complete", tableName)
	if err := waitForDynamoDbTtlUpdateToBeCompleted(tableName, ttlMap["enabled"].(bool), conn); err != nil {
		return fmt.Errorf("error waiting for DynamoDB Table (%s) Time To Live update: %s", tableName, err)
	}

	return nil
}

func updateDynamoDbPITR(d *schema.ResourceData, conn *dynamodb.DynamoDB) error {
	toEnable := d.Get("point_in_time_recovery.0.enabled").(bool)

	input := &dynamodb.UpdateContinuousBackupsInput{
		TableName: aws.String(d.Id()),
		PointInTimeRecoverySpecification: &dynamodb.PointInTimeRecoverySpecification{
			PointInTimeRecoveryEnabled: aws.Bool(toEnable),
		},
	}

	log.Printf("[DEBUG] Updating DynamoDB point in time recovery status to %v", toEnable)

	err := resource.Retry(20*time.Minute, func() *resource.RetryError {
		_, err := conn.UpdateContinuousBackups(input)
		if err != nil {
			// Backups are still being enabled for this newly created table
			if isAWSErr(err, dynamodb.ErrCodeContinuousBackupsUnavailableException, "Backups are being enabled") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.UpdateContinuousBackups(input)
	}
	if err != nil {
		return fmt.Errorf("Error updating DynamoDB PITR status: %s", err)
	}

	if err := waitForDynamoDbBackupUpdateToBeCompleted(d.Id(), toEnable, conn); err != nil {
		return fmt.Errorf("Error waiting for DynamoDB PITR update: %s", err)
	}

	return nil
}

func readDynamoDbTableTags(arn string, conn *dynamodb.DynamoDB) (map[string]string, error) {
	output, err := conn.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
		ResourceArn: aws.String(arn),
	})

	// Do not fail if interfacing with dynamodb-local
	if err != nil && !isAWSErr(err, "UnknownOperationException", "Tagging is not currently supported in DynamoDB Local.") {
		return nil, fmt.Errorf("Error reading tags from dynamodb resource: %s", err)
	}

	result := keyvaluetags.DynamodbKeyValueTags(output.Tags).IgnoreAws().Map()

	// TODO Read NextToken if available

	return result, nil
}

// Waiters

func waitForDynamoDbGSIToBeActive(tableName string, gsiName string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusCreating,
			dynamodb.IndexStatusUpdating,
		},
		Target:  []string{dynamodb.IndexStatusActive},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			table := result.Table

			// Find index
			var targetGSI *dynamodb.GlobalSecondaryIndexDescription
			for _, gsi := range table.GlobalSecondaryIndexes {
				if *gsi.IndexName == gsiName {
					targetGSI = gsi
				}
			}

			if targetGSI != nil {
				return table, *targetGSI.IndexStatus, nil
			}

			return nil, "", nil
		},
	}
	_, err := stateConf.WaitForState()
	return err
}

func waitForDynamoDbGSIToBeDeleted(tableName string, gsiName string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusActive,
			dynamodb.IndexStatusDeleting,
		},
		Target:  []string{},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			table := result.Table

			// Find index
			var targetGSI *dynamodb.GlobalSecondaryIndexDescription
			for _, gsi := range table.GlobalSecondaryIndexes {
				if *gsi.IndexName == gsiName {
					targetGSI = gsi
				}
			}

			if targetGSI == nil {
				return nil, "", nil
			}

			return targetGSI, *targetGSI.IndexStatus, nil
		},
	}
	_, err := stateConf.WaitForState()
	return err
}

func waitForDynamoDbTableToBeActive(tableName string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.TableStatusCreating, dynamodb.TableStatusUpdating},
		Target:  []string{dynamodb.TableStatusActive},
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			return result, *result.Table.TableStatus, nil
		},
	}
	_, err := stateConf.WaitForState()

	return err
}

func waitForDynamoDbBackupUpdateToBeCompleted(tableName string, toEnable bool, conn *dynamodb.DynamoDB) error {
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
		Timeout: 10 * time.Second,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeContinuousBackups(&dynamodb.DescribeContinuousBackupsInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			if result.ContinuousBackupsDescription == nil || result.ContinuousBackupsDescription.PointInTimeRecoveryDescription == nil {
				return 42, "", errors.New("Error reading backup status from dynamodb resource: empty description")
			}
			pitr := result.ContinuousBackupsDescription.PointInTimeRecoveryDescription

			return result, *pitr.PointInTimeRecoveryStatus, nil
		},
	}
	_, err := stateConf.WaitForState()
	return err
}

func waitForDynamoDbTtlUpdateToBeCompleted(tableName string, toEnable bool, conn *dynamodb.DynamoDB) error {
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
		Timeout: 10 * time.Second,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeTimeToLive(&dynamodb.DescribeTimeToLiveInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			ttlDesc := result.TimeToLiveDescription

			return result, *ttlDesc.TimeToLiveStatus, nil
		},
	}

	_, err := stateConf.WaitForState()
	return err
}

func waitForDynamoDbSSEUpdateToBeCompleted(tableName string, timeout time.Duration, conn *dynamodb.DynamoDB) error {
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
		Timeout: timeout,
		Refresh: func() (interface{}, string, error) {
			result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				return 42, "", err
			}

			// Disabling SSE returns null SSEDescription.
			if result.Table.SSEDescription == nil {
				return result, dynamodb.SSEStatusDisabled, nil
			}
			return result, aws.StringValue(result.Table.SSEDescription.Status), nil
		},
	}
	_, err := stateConf.WaitForState()

	return err
}

func isDynamoDbTableOptionDisabled(v interface{}) bool {
	options := v.([]interface{})
	if len(options) == 0 {
		return true
	}
	e := options[0].(map[string]interface{})["enabled"]
	return !e.(bool)
}
