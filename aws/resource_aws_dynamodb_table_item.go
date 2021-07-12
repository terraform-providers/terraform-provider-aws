package aws

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsDynamoDbTableItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDynamoDbTableItemCreate,
		Read:   resourceAwsDynamoDbTableItemRead,
		Update: resourceAwsDynamoDbTableItemUpdate,
		Delete: resourceAwsDynamoDbTableItemDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsDynamoDbTableItemImport,
		},

		Schema: map[string]*schema.Schema{
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
				ForceNew: true,
				Optional: true,
			},
			"item": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateDynamoDbTableItem,
			},
		},
	}
}

func validateDynamoDbTableItem(v interface{}, k string) (ws []string, errors []error) {
	_, err := expandDynamoDbTableItemAttributes(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("Invalid format of %q: %s", k, err))
	}
	return
}

func resourceAwsDynamoDbTableItemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	tableName := d.Get("table_name").(string)
	hashKey := d.Get("hash_key").(string)
	item := d.Get("item").(string)
	attributes, err := expandDynamoDbTableItemAttributes(item)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] DynamoDB item create: %s", tableName)

	_, err = conn.PutItem(&dynamodb.PutItemInput{
		Item: attributes,
		// Explode if item exists. We didn't create it.
		Expected: map[string]*dynamodb.ExpectedAttributeValue{
			hashKey: {
				Exists: aws.Bool(false),
			},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		return err
	}

	rangeKey := d.Get("range_key").(string)
	id := buildDynamoDbTableItemId(tableName, hashKey, rangeKey, attributes)

	d.SetId(id)

	return resourceAwsDynamoDbTableItemRead(d, meta)
}

func resourceAwsDynamoDbTableItemUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating DynamoDB table %s", d.Id())
	conn := meta.(*AWSClient).dynamodbconn

	if d.HasChange("item") {
		tableName := d.Get("table_name").(string)
		hashKey := d.Get("hash_key").(string)
		rangeKey := d.Get("range_key").(string)

		oldItem, newItem := d.GetChange("item")

		attributes, err := expandDynamoDbTableItemAttributes(newItem.(string))
		if err != nil {
			return err
		}
		newQueryKey := buildDynamoDbTableItemQueryKey(attributes, hashKey, rangeKey)

		updates := map[string]*dynamodb.AttributeValueUpdate{}
		for key, value := range attributes {
			// Hash keys and range keys are not updatable, so we'll basically create
			// a new record and delete the old one below
			if key == hashKey || key == rangeKey {
				continue
			}
			updates[key] = &dynamodb.AttributeValueUpdate{
				Action: aws.String(dynamodb.AttributeActionPut),
				Value:  value,
			}
		}

		_, err = conn.UpdateItem(&dynamodb.UpdateItemInput{
			AttributeUpdates: updates,
			TableName:        aws.String(tableName),
			Key:              newQueryKey,
		})
		if err != nil {
			return err
		}

		oItem := oldItem.(string)
		oldAttributes, err := expandDynamoDbTableItemAttributes(oItem)
		if err != nil {
			return err
		}

		// New record is created via UpdateItem in case we're changing hash key
		// so we need to get rid of the old one
		oldQueryKey := buildDynamoDbTableItemQueryKey(oldAttributes, hashKey, rangeKey)
		if !reflect.DeepEqual(oldQueryKey, newQueryKey) {
			log.Printf("[DEBUG] Deleting old record: %#v", oldQueryKey)
			_, err := conn.DeleteItem(&dynamodb.DeleteItemInput{
				Key:       oldQueryKey,
				TableName: aws.String(tableName),
			})
			if err != nil {
				return err
			}
		}

		id := buildDynamoDbTableItemId(tableName, hashKey, rangeKey, attributes)
		d.SetId(id)
	}

	return resourceAwsDynamoDbTableItemRead(d, meta)
}

func resourceAwsDynamoDbTableItemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	log.Printf("[DEBUG] Loading data for DynamoDB table item '%s'", d.Id())

	tableName := d.Get("table_name").(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	attributes, err := expandDynamoDbTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return err
	}

	result, err := conn.GetItem(&dynamodb.GetItemInput{
		TableName:                aws.String(tableName),
		ConsistentRead:           aws.Bool(true),
		Key:                      buildDynamoDbTableItemQueryKey(attributes, hashKey, rangeKey),
		ProjectionExpression:     buildDynamoDbProjectionExpression(attributes),
		ExpressionAttributeNames: buildDynamoDbExpressionAttributeNames(attributes),
	})
	if err != nil {
		if isAWSErr(err, dynamodb.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Dynamodb Table Item (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving DynamoDB table item: %s", err)
	}

	if result.Item == nil {
		log.Printf("[WARN] Dynamodb Table Item (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	// The record exists, now test if it differs from what is desired
	if !reflect.DeepEqual(result.Item, attributes) {
		itemAttrs, err := flattenDynamoDbTableItemAttributes(result.Item)
		if err != nil {
			return err
		}
		d.Set("item", itemAttrs)
		id := buildDynamoDbTableItemId(tableName, hashKey, rangeKey, result.Item)
		d.SetId(id)
	}

	return nil
}

func resourceAwsDynamoDbTableItemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dynamodbconn

	attributes, err := expandDynamoDbTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return err
	}
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	queryKey := buildDynamoDbTableItemQueryKey(attributes, hashKey, rangeKey)

	_, err = conn.DeleteItem(&dynamodb.DeleteItemInput{
		Key:       queryKey,
		TableName: aws.String(d.Get("table_name").(string)),
	})
	return err
}

// Helpers

func buildDynamoDbExpressionAttributeNames(attrs map[string]*dynamodb.AttributeValue) map[string]*string {
	names := map[string]*string{}
	for key := range attrs {
		names["#a_"+key] = aws.String(key)
	}

	return names
}

func buildDynamoDbProjectionExpression(attrs map[string]*dynamodb.AttributeValue) *string {
	keys := []string{}
	for key := range attrs {
		keys = append(keys, key)
	}
	return aws.String("#a_" + strings.Join(keys, ", #a_"))
}

func buildDynamoDbTableItemId(tableName string, hashKey string, rangeKey string, attrs map[string]*dynamodb.AttributeValue) string {
	id := []string{tableName, hashKey}

	if hashVal, ok := attrs[hashKey]; ok {
		id = append(id, base64Encode(hashVal.B))
		id = append(id, aws.StringValue(hashVal.S))
		id = append(id, aws.StringValue(hashVal.N))
	}
	if rangeVal, ok := attrs[rangeKey]; ok && rangeKey != "" {
		id = append(id, rangeKey, base64Encode(rangeVal.B))
		id = append(id, aws.StringValue(rangeVal.S))
		id = append(id, aws.StringValue(rangeVal.N))
	}
	return strings.Join(id, "|")
}

func buildDynamoDbTableItemQueryKey(attrs map[string]*dynamodb.AttributeValue, hashKey string, rangeKey string) map[string]*dynamodb.AttributeValue {
	queryKey := map[string]*dynamodb.AttributeValue{
		hashKey: attrs[hashKey],
	}
	if rangeKey != "" {
		queryKey[rangeKey] = attrs[rangeKey]
	}

	return queryKey
}

func parseDynamoDbTableItemQueryKey(d *schema.ResourceData, keyType string, keyName string, keyValue string, tableDescription *dynamodb.TableDescription, attrs map[string]*dynamodb.AttributeValue) error {
	if keyValue == "" {
		return fmt.Errorf("No value given for %s %s", keyType, keyName)
	}

	var value *dynamodb.AttributeValue
	found := false

	// Find the matching attribute definition and construct an appropriate attribute value based on the type
	for _, attr := range tableDescription.AttributeDefinitions {
		if *attr.AttributeName == keyName {
			found = true
			switch *attr.AttributeType {
			case "B":
				data, err := base64.StdEncoding.DecodeString(keyValue)
				if err != nil {
					return err
				}
				value = &dynamodb.AttributeValue{
					B: data,
				}
			case "S":
				value = &dynamodb.AttributeValue{
					S: aws.String(keyValue),
				}
			case "N":
				value = &dynamodb.AttributeValue{
					N: aws.String(keyValue),
				}
			default:
				return fmt.Errorf("%s %s has invalid type %s", keyType, keyName, *attr.AttributeType)
			}
		}
	}

	// Set both the resource attribute and the attribute query parameter
	if !found {
		return fmt.Errorf("%s %s not found in table", keyType, keyName)
	}
	d.Set(keyType, keyName)
	attrs[keyName] = value

	return nil
}

func resourceAwsDynamoDbTableItemImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Parse given id string as either a pipe-delimited string or json
	var id []string
	if strings.HasPrefix(d.Id(), "[") {
		err := json.Unmarshal([]byte(d.Id()), &id)
		if err != nil {
			return nil, err
		}
	} else {
		id = strings.Split(d.Id(), "|")
	}

	// Check for proper number of array elements
	if len(id) == 2 {
		id = append(id, "")
	}
	if len(id) != 3 || id[0] == "" || id[1] == "" {
		return nil, errors.New("Invalid id, must be of the form table_name|hash_key_value[|range_key_value] or json [ \"table_name\", \"hash_key_value\", \"range_key_value\" ]")
	}

	// Initialize table query parameters
	tableName := id[0]
	hashKey := ""
	rangeKey := ""
	hashKeyValueString := id[1]
	rangeKeyValueString := id[2]
	params := &dynamodb.GetItemInput{
		TableName:      aws.String(tableName),
		ConsistentRead: aws.Bool(true),
		Key:            map[string]*dynamodb.AttributeValue{},
	}

	// Query table description to determine its hash/range key attributes
	conn := meta.(*AWSClient).dynamodbconn
	tableResult, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, err
	}
	tableDescription := tableResult.Table

	// Build attribute query using given hash/range key values
	for _, key := range tableDescription.KeySchema {
		switch *key.KeyType {
		case "HASH":
			hashKey = *key.AttributeName
			err := parseDynamoDbTableItemQueryKey(d, "hash_key", hashKey, hashKeyValueString, tableDescription, params.Key)
			if err != nil {
				return nil, err
			}
		case "RANGE":
			rangeKey = *key.AttributeName
			err := parseDynamoDbTableItemQueryKey(d, "range_key", rangeKey, rangeKeyValueString, tableDescription, params.Key)
			if err != nil {
				return nil, err
			}
		}
	}

	// Error if we were given a range key value but the table has no range key
	if rangeKey == "" && rangeKeyValueString != "" {
		return nil, fmt.Errorf("Table %s has no range key but a range key value was given", tableName)
	}

	// Query table for matching record
	result, err := conn.GetItem(params)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, fmt.Errorf("No item matching %s found to import", d.Id())
	}
	itemAttrs, err := flattenDynamoDbTableItemAttributes(result.Item)
	if err != nil {
		return nil, err
	}

	// Set required resource attributes
	d.Set("table_name", tableName)
	d.Set("hash_key", hashKey)
	if rangeKey != "" {
		d.Set("range_key", rangeKey)
	}
	d.Set("item", itemAttrs)

	// Always set id to canonical format
	d.SetId(buildDynamoDbTableItemId(tableName, hashKey, rangeKey, params.Key))

	return []*schema.ResourceData{d}, nil
}
