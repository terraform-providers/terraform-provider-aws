package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dynamodb_table", &resource.Sweeper{
		Name: "aws_dynamodb_table",
		F:    testSweepDynamoDbTables,
	})
}

func testSweepDynamoDbTables(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).dynamodbconn

	err = conn.ListTablesPages(&dynamodb.ListTablesInput{}, func(out *dynamodb.ListTablesOutput, lastPage bool) bool {
		for _, tableName := range out.TableNames {
			log.Printf("[INFO] Deleting DynamoDB Table: %s", *tableName)

			err := deleteAwsDynamoDbTable(*tableName, conn)
			if err != nil {
				log.Printf("[ERROR] Failed to delete DynamoDB Table %s: %s", *tableName, err)
				continue
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DynamoDB Table sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DynamoDB Tables: %s", err)
	}

	return nil
}

func TestDiffDynamoDbGSI(t *testing.T) {
	testCases := []struct {
		Old             []interface{}
		New             []interface{}
		ExpectedUpdates []*dynamodb.GlobalSecondaryIndexUpdate
	}{
		{ // No-op => no changes
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{},
		},
		{ // No-op => ignore ordering of non_key_attributes
			Old: []interface{}{
				map[string]interface{}{
					"name":               "att1-index",
					"hash_key":           "att1",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "INCLUDE",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"attr3", "attr1", "attr2"}),
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":               "att1-index",
					"hash_key":           "att1",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "INCLUDE",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"attr1", "attr2", "attr3"}),
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{},
		},

		{ // Creation
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
				map[string]interface{}{
					"name":            "att2-index",
					"hash_key":        "att2",
					"write_capacity":  12,
					"read_capacity":   11,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Create: &dynamodb.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att2-index"),
						KeySchema: []*dynamodb.KeySchemaElement{
							{
								AttributeName: aws.String("att2"),
								KeyType:       aws.String("HASH"),
							},
						},
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(12),
							ReadCapacityUnits:  aws.Int64(11),
						},
						Projection: &dynamodb.Projection{
							ProjectionType: aws.String("ALL"),
						},
					},
				},
			},
		},

		{ // Deletion
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
				map[string]interface{}{
					"name":            "att2-index",
					"hash_key":        "att2",
					"write_capacity":  12,
					"read_capacity":   11,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att2-index"),
					},
				},
			},
		},

		{ // Update
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  20,
					"read_capacity":   30,
					"projection_type": "ALL",
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Update: &dynamodb.UpdateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(20),
							ReadCapacityUnits:  aws.Int64(30),
						},
					},
				},
			},
		},

		{ // Update of non-capacity attributes
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":               "att1-index",
					"hash_key":           "att-new",
					"range_key":          "new-range-key",
					"write_capacity":     10,
					"read_capacity":      10,
					"projection_type":    "KEYS_ONLY",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"RandomAttribute"}),
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
					},
				},
				{
					Create: &dynamodb.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						KeySchema: []*dynamodb.KeySchemaElement{
							{
								AttributeName: aws.String("att-new"),
								KeyType:       aws.String("HASH"),
							},
							{
								AttributeName: aws.String("new-range-key"),
								KeyType:       aws.String("RANGE"),
							},
						},
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(10),
							ReadCapacityUnits:  aws.Int64(10),
						},
						Projection: &dynamodb.Projection{
							ProjectionType:   aws.String("KEYS_ONLY"),
							NonKeyAttributes: aws.StringSlice([]string{"RandomAttribute"}),
						},
					},
				},
			},
		},

		{ // Update of all attributes
			Old: []interface{}{
				map[string]interface{}{
					"name":            "att1-index",
					"hash_key":        "att1",
					"write_capacity":  10,
					"read_capacity":   10,
					"projection_type": "ALL",
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"name":               "att1-index",
					"hash_key":           "att-new",
					"range_key":          "new-range-key",
					"write_capacity":     12,
					"read_capacity":      12,
					"projection_type":    "INCLUDE",
					"non_key_attributes": schema.NewSet(schema.HashString, []interface{}{"RandomAttribute"}),
				},
			},
			ExpectedUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{
				{
					Delete: &dynamodb.DeleteGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
					},
				},
				{
					Create: &dynamodb.CreateGlobalSecondaryIndexAction{
						IndexName: aws.String("att1-index"),
						KeySchema: []*dynamodb.KeySchemaElement{
							{
								AttributeName: aws.String("att-new"),
								KeyType:       aws.String("HASH"),
							},
							{
								AttributeName: aws.String("new-range-key"),
								KeyType:       aws.String("RANGE"),
							},
						},
						ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
							WriteCapacityUnits: aws.Int64(12),
							ReadCapacityUnits:  aws.Int64(12),
						},
						Projection: &dynamodb.Projection{
							ProjectionType:   aws.String("INCLUDE"),
							NonKeyAttributes: aws.StringSlice([]string{"RandomAttribute"}),
						},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		ops, err := diffDynamoDbGSI(tc.Old, tc.New, dynamodb.BillingModeProvisioned)
		if err != nil {
			t.Fatal(err)
		}

		// Convert to strings to avoid dealing with pointers
		opsS := fmt.Sprintf("%v", ops)
		opsExpectedS := fmt.Sprintf("%v", tc.ExpectedUpdates)

		if opsS != opsExpectedS {
			t.Fatalf("Case #%d: Given:\n%s\n\nExpected:\n%s",
				i, opsS, opsExpectedS)
		}
	}
}

func TestAccAWSDynamoDbTable_basic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "TestTableHashKey",
						"type": "S",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDynamoDbTable_disappears(t *testing.T) {
	var table1 dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDynamoDbTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDynamoDbTable_disappears_PayPerRequestWithGSI(t *testing.T) {
	var table1, table2 dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequestWithGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDynamoDbTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequestWithGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDynamoDbTable_extended(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					testAccCheckInitialAWSDynamoDbTableConf(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfigAddSecondaryGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
					resource.TestCheckResourceAttr(resourceName, "range_key", "TestTableRangeKey"),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
					resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "read_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "attribute.#", "4"),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "TestTableHashKey",
						"type": "S",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "TestTableRangeKey",
						"type": "S",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "TestLSIRangeKey",
						"type": "N",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
						"name": "ReplacementGSIRangeKey",
						"type": "N",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"name":                 "ReplacementTestTableGSI",
						"hash_key":             "TestTableHashKey",
						"range_key":            "ReplacementGSIRangeKey",
						"write_capacity":       "5",
						"read_capacity":        "5",
						"projection_type":      "INCLUDE",
						"non_key_attributes.#": "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "TestNonKeyAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
						"name":            "TestTableLSI",
						"range_key":       "TestLSIRangeKey",
						"projection_type": "ALL",
					}),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_enablePitr(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					testAccCheckInitialAWSDynamoDbTableConf(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfig_backup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "point_in_time_recovery.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_BillingMode_PayPerRequestToProvisioned(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequest(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbBilling_Provisioned(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_BillingMode_ProvisionedToPayPerRequest(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbBilling_Provisioned(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequest(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_BillingMode_GSI_PayPerRequestToProvisioned(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequestWithGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbBilling_ProvisionedWithGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_BillingMode_GSI_ProvisionedToPayPerRequest(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbBilling_ProvisionedWithGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequestWithGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModePayPerRequest),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_streamSpecification(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	tableName := fmt.Sprintf("TerraformTestStreamTable-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigStreamSpecification(tableName, true, "KEYS_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", "KEYS_ONLY"),
					testAccMatchResourceAttrRegionalARN(resourceName, "stream_arn", "dynamodb", regexp.MustCompile(fmt.Sprintf("table/%s/stream", tableName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfigStreamSpecification(tableName, false, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "stream_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "stream_view_type", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "stream_arn", "dynamodb", regexp.MustCompile(fmt.Sprintf("table/%s/stream", tableName))),
					resource.TestCheckResourceAttrSet(resourceName, "stream_label"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_streamSpecificationValidation(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSDynamoDbConfigStreamSpecification("anything", true, ""),
				ExpectError: regexp.MustCompile(`stream_view_type is required when stream_enabled = true`),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_tags(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					testAccCheckInitialAWSDynamoDbTableConf(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/13243
func TestAccAWSDynamoDbTable_gsiUpdateCapacity(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "1",
						"write_capacity": "1",
						"name":           "att1-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "1",
						"write_capacity": "1",
						"name":           "att2-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "1",
						"write_capacity": "1",
						"name":           "att3-index",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedCapacity(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "2",
						"write_capacity": "2",
						"name":           "att1-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "2",
						"write_capacity": "2",
						"name":           "att2-index",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"read_capacity":  "2",
						"write_capacity": "2",
						"name":           "att3-index",
					}),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_gsiUpdateOtherAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att2",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "1",
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15115
func TestAccAWSDynamoDbTable_lsiNonKeyAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigLsiNonKeyAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
						"name":                 "TestTableLSI",
						"non_key_attributes.#": "1",
						"non_key_attributes.0": "TestNonKeyAttribute",
						"projection_type":      "INCLUDE",
						"range_key":            "TestLSIRangeKey",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/566
func TestAccAWSDynamoDbTable_gsiUpdateNonKeyAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "1",
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedNonKeyAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att4",
						"name":                 "att2-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att3",
						"name":                 "att3-index",
						"non_key_attributes.#": "2",
						"projection_type":      "INCLUDE",
						"range_key":            "att4",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "AnotherAttribute"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "0",
						"projection_type":      "ALL",
						"range_key":            "",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/671
func TestAccAWSDynamoDbTable_gsiUpdateNonKeyAttributes_emptyPlan(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	name := acctest.RandString(10)
	attributes := fmt.Sprintf("%q, %q", "AnotherAttribute", "RandomAttribute")
	reorderedAttributes := fmt.Sprintf("%q, %q", "RandomAttribute", "AnotherAttribute")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiMultipleNonKeyAttributes(name, attributes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
						"hash_key":             "att1",
						"name":                 "att1-index",
						"non_key_attributes.#": "2",
						"projection_type":      "INCLUDE",
						"range_key":            "att2",
						"read_capacity":        "1",
						"write_capacity":       "1",
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "AnotherAttribute"),
					resource.TestCheckTypeSetElemAttr(resourceName, "global_secondary_index.*.non_key_attributes.*", "RandomAttribute"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccAWSDynamoDbConfigGsiMultipleNonKeyAttributes(name, reorderedAttributes),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// TTL tests must be split since it can only be updated once per hour
// ValidationException: Time to live has been modified multiple times within a fixed interval
func TestAccAWSDynamoDbTable_Ttl_Enabled(t *testing.T) {
	var table dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigTimeToLive(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TTL tests must be split since it can only be updated once per hour
// ValidationException: Time to live has been modified multiple times within a fixed interval
func TestAccAWSDynamoDbTable_Ttl_Disabled(t *testing.T) {
	var table dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigTimeToLive(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfigTimeToLive(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "ttl.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ttl.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_attributeUpdate(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigOneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Attribute type change
				Config: testAccAWSDynamoDbConfigOneAttribute(rName, "firstKey", "firstKey", "N"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
				),
			},
			{ // New attribute addition (index update)
				Config: testAccAWSDynamoDbConfigTwoAttributes(rName, "firstKey", "secondKey", "firstKey", "N", "secondKey", "S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
				),
			},
			{ // Attribute removal (index update)
				Config: testAccAWSDynamoDbConfigOneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_lsiUpdate(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigLSI(rName, "lsi-original"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Change name of local secondary index
				Config: testAccAWSDynamoDbConfigLSI(rName, "lsi-changed"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_attributeUpdateValidation(t *testing.T) {
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSDynamoDbConfigOneAttribute(rName, "firstKey", "unusedKey", "S"),
				ExpectError: regexp.MustCompile(`All attributes must be indexed. Unused attributes: \["unusedKey"\]`),
			},
			{
				Config:      testAccAWSDynamoDbConfigTwoAttributes(rName, "firstKey", "secondKey", "firstUnused", "N", "secondUnused", "S"),
				ExpectError: regexp.MustCompile(`All attributes must be indexed. Unused attributes: \["firstUnused"\ \"secondUnused\"]`),
			},
			{
				Config:      testAccAWSDynamoDbConfigUnmatchedIndexes(rName, "firstUnused", "secondUnused"),
				ExpectError: regexp.MustCompile(`All indexes must match a defined attribute. Unmatched indexes: \["firstUnused"\ \"secondUnused\"]`),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_encryption(t *testing.T) {
	var confBYOK, confEncEnabled, confEncDisabled dynamodb.DescribeTableOutput
	resourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")
	kmsKeyResourceName := "aws_kms_key.test"
	kmsAliasDatasourceName := "data.aws_kms_alias.dynamodb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialStateWithEncryptionBYOK(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &confBYOK),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbConfigInitialStateWithEncryptionAmazonCMK(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &confEncDisabled),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "0"),
					func(s *terraform.State) error {
						if !confEncDisabled.Table.CreationDateTime.Equal(*confBYOK.Table.CreationDateTime) {
							return fmt.Errorf("DynamoDB table recreated when changing SSE")
						}
						return nil
					},
				),
			},
			{
				Config: testAccAWSDynamoDbConfigInitialStateWithEncryptionAmazonCMK(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &confEncEnabled),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.kms_key_arn", kmsAliasDatasourceName, "target_key_arn"),
					func(s *terraform.State) error {
						if !confEncEnabled.Table.CreationDateTime.Equal(*confEncDisabled.Table.CreationDateTime) {
							return fmt.Errorf("DynamoDB table recreated when changing SSE")
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccCheckAWSDynamoDbTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_table" {
			continue
		}

		log.Printf("[DEBUG] Checking if DynamoDB table %s exists", rs.Primary.ID)
		// Check if queue exists by checking for its attributes
		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeTable(params)
		if err == nil {
			return fmt.Errorf("DynamoDB table %s still exists. Failing!", rs.Primary.ID)
		}

		// Verify the error is what we want
		if dbErr, ok := err.(awserr.Error); ok && dbErr.Code() == "ResourceNotFoundException" {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckInitialAWSDynamoDbTableExists(n string, table *dynamodb.DescribeTableOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Printf("[DEBUG] Trying to create initial table state!")
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		params := &dynamodb.DescribeTableInput{
			TableName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeTable(params)

		if err != nil {
			return fmt.Errorf("Problem describing table '%s': %s", rs.Primary.ID, err)
		}

		*table = *resp

		return nil
	}
}

func testAccCheckInitialAWSDynamoDbTableConf(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "hash_key", "TestTableHashKey"),
		resource.TestCheckResourceAttr(resourceName, "range_key", "TestTableRangeKey"),
		resource.TestCheckResourceAttr(resourceName, "billing_mode", dynamodb.BillingModeProvisioned),
		resource.TestCheckResourceAttr(resourceName, "write_capacity", "2"),
		resource.TestCheckResourceAttr(resourceName, "read_capacity", "1"),
		resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "0"),
		resource.TestCheckResourceAttr(resourceName, "attribute.#", "4"),
		resource.TestCheckResourceAttr(resourceName, "global_secondary_index.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "local_secondary_index.#", "1"),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestTableHashKey",
			"type": "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestTableRangeKey",
			"type": "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestLSIRangeKey",
			"type": "N",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "attribute.*", map[string]string{
			"name": "TestGSIRangeKey",
			"type": "S",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "global_secondary_index.*", map[string]string{
			"name":            "InitialTestTableGSI",
			"hash_key":        "TestTableHashKey",
			"range_key":       "TestGSIRangeKey",
			"write_capacity":  "1",
			"read_capacity":   "1",
			"projection_type": "KEYS_ONLY",
		}),
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "local_secondary_index.*", map[string]string{
			"name":            "TestTableLSI",
			"range_key":       "TestLSIRangeKey",
			"projection_type": "ALL",
		}),
	)
}

func TestAccAWSDynamoDbTable_Replica_Multiple(t *testing.T) {
	var table dynamodb.DescribeTableOutput
	var providers []*schema.Provider
	resourceName := "aws_dynamodb_table.test"
	tableName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 3)
		},
		ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbTableConfigReplica2(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
				),
			},
			{
				Config:            testAccAWSDynamoDbTableConfigReplica2(tableName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbTableConfigReplica0(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "0"),
				),
			},
			{
				Config: testAccAWSDynamoDbTableConfigReplica2(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_Replica_Single(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	var providers []*schema.Provider
	resourceName := "aws_dynamodb_table.test"
	tableName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 3), // 3 due to shared test configuration
		CheckDestroy:      testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbTableConfigReplica1(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
			{
				Config:            testAccAWSDynamoDbTableConfigReplica1(tableName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDynamoDbTableConfigReplica0(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "0"),
				),
			},
			{
				Config: testAccAWSDynamoDbTableConfigReplica1(tableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "replica.#", "1"),
				),
			},
		},
	})
}

func testAccAWSDynamoDbConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccAWSDynamoDbConfig_backup(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  point_in_time_recovery {
    enabled = true
  }
}
`, rName)
}

func testAccAWSDynamoDbBilling_PayPerRequest(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = "%s"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccAWSDynamoDbBilling_Provisioned(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = "%s"
  billing_mode = "PROVISIONED"
  hash_key     = "TestTableHashKey"

  read_capacity  = 5
  write_capacity = 5

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccAWSDynamoDbBilling_PayPerRequestWithGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name         = "%s"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableGSIKey"
    type = "S"
  }

  global_secondary_index {
    name            = "TestTableGSI"
    hash_key        = "TestTableGSIKey"
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccAWSDynamoDbBilling_ProvisionedWithGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  billing_mode   = "PROVISIONED"
  hash_key       = "TestTableHashKey"
  name           = %q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableGSIKey"
    type = "S"
  }

  global_secondary_index {
    hash_key        = "TestTableGSIKey"
    name            = "TestTableGSI"
    projection_type = "KEYS_ONLY"
    read_capacity   = 1
    write_capacity  = 1
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigInitialState(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name            = "TestTableLSI"
    range_key       = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "InitialTestTableGSI"
    hash_key        = "TestTableHashKey"
    range_key       = "TestGSIRangeKey"
    write_capacity  = 1
    read_capacity   = 1
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigInitialStateWithEncryptionAmazonCMK(rName string, enabled bool) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "dynamodb" {
  name = "alias/aws/dynamodb"
}

resource "aws_kms_key" "test" {
  description = "DynamoDbTest"
}

resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  server_side_encryption {
    enabled = %t
  }
}
`, rName, enabled)
}

func testAccAWSDynamoDbConfigInitialStateWithEncryptionBYOK(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "DynamoDbTest"
}

resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  server_side_encryption {
    enabled     = true
    kms_key_arn = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigAddSecondaryGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 2
  write_capacity = 2
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "ReplacementGSIRangeKey"
    type = "N"
  }

  local_secondary_index {
    name            = "TestTableLSI"
    range_key       = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name               = "ReplacementTestTableGSI"
    hash_key           = "TestTableHashKey"
    range_key          = "ReplacementGSIRangeKey"
    write_capacity     = 5
    read_capacity      = 5
    projection_type    = "INCLUDE"
    non_key_attributes = ["TestNonKeyAttribute"]
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigStreamSpecification(tableName string, enabled bool, viewType string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  stream_enabled   = %t
  stream_view_type = "%s"
}
`, tableName, enabled, viewType)
}

func testAccAWSDynamoDbConfigTags() string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "TerraformTestTable-%d"
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  attribute {
    name = "TestGSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name            = "TestTableLSI"
    range_key       = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "InitialTestTableGSI"
    hash_key        = "TestTableHashKey"
    range_key       = "TestGSIRangeKey"
    write_capacity  = 1
    read_capacity   = 1
    projection_type = "KEYS_ONLY"
  }

  tags = {
    Name    = "terraform-test-table-%d"
    AccTest = "yes"
    Testing = "absolutely"
  }
}
`, acctest.RandInt(), acctest.RandInt())
}

func testAccAWSDynamoDbConfigGsiUpdate(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }
}
`, name)
}

func testAccAWSDynamoDbConfigGsiUpdatedCapacity(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 2
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }
}
`, name)
}

func testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  attribute {
    name = "att4"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name               = "att3-index"
    hash_key           = "att3"
    range_key          = "att4"
    write_capacity     = var.capacity
    read_capacity      = var.capacity
    projection_type    = "INCLUDE"
    non_key_attributes = ["RandomAttribute"]
  }
}
`, name)
}

func testAccAWSDynamoDbConfigGsiUpdatedNonKeyAttributes(name string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  attribute {
    name = "att3"
    type = "S"
  }

  attribute {
    name = "att4"
    type = "S"
  }

  global_secondary_index {
    name            = "att1-index"
    hash_key        = "att1"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = var.capacity
    read_capacity   = var.capacity
    projection_type = "ALL"
  }

  global_secondary_index {
    name               = "att3-index"
    hash_key           = "att3"
    range_key          = "att4"
    write_capacity     = var.capacity
    read_capacity      = var.capacity
    projection_type    = "INCLUDE"
    non_key_attributes = ["RandomAttribute", "AnotherAttribute"]
  }
}
`, name)
}

func testAccAWSDynamoDbConfigGsiMultipleNonKeyAttributes(name, attributes string) string {
	return fmt.Sprintf(`
variable "capacity" {
  default = 1
}

resource "aws_dynamodb_table" "test" {
  name           = "tf-acc-test-%s"
  read_capacity  = var.capacity
  write_capacity = var.capacity
  hash_key       = "id"

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "att1"
    type = "S"
  }

  attribute {
    name = "att2"
    type = "S"
  }

  global_secondary_index {
    name               = "att1-index"
    hash_key           = "att1"
    range_key          = "att2"
    write_capacity     = var.capacity
    read_capacity      = var.capacity
    projection_type    = "INCLUDE"
    non_key_attributes = [%s]
  }
}
`, name, attributes)
}

func testAccAWSDynamoDbConfigLsiNonKeyAttributes(name string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  hash_key       = "TestTableHashKey"
  range_key      = "TestTableRangeKey"
  write_capacity = 1
  read_capacity  = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableRangeKey"
    type = "S"
  }

  attribute {
    name = "TestLSIRangeKey"
    type = "N"
  }

  local_secondary_index {
    name               = "TestTableLSI"
    range_key          = "TestLSIRangeKey"
    projection_type    = "INCLUDE"
    non_key_attributes = ["TestNonKeyAttribute"]
  }
}
`, name)
}

func testAccAWSDynamoDbConfigTimeToLive(rName string, ttlEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  hash_key       = "TestTableHashKey"
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  ttl {
    attribute_name = %[2]t ? "TestTTL" : ""
    enabled        = %[2]t
  }
}
`, rName, ttlEnabled)
}

func testAccAWSDynamoDbConfigOneAttribute(rName, hashKey, attrName, attrType string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%[1]s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  attribute {
    name = "%[3]s"
    type = "%[4]s"
  }

  global_secondary_index {
    name            = "gsiName"
    hash_key        = "%[2]s"
    write_capacity  = 10
    read_capacity   = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName, hashKey, attrName, attrType)
}

func testAccAWSDynamoDbConfigTwoAttributes(rName, hashKey, rangeKey, attrName1, attrType1, attrName2, attrType2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%[1]s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  attribute {
    name = "%[4]s"
    type = "%[5]s"
  }

  attribute {
    name = "%[6]s"
    type = "%[7]s"
  }

  global_secondary_index {
    name            = "gsiName"
    hash_key        = "%[2]s"
    range_key       = "%[3]s"
    write_capacity  = 10
    read_capacity   = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName, hashKey, rangeKey, attrName1, attrType1, attrName2, attrType2)
}

func testAccAWSDynamoDbConfigUnmatchedIndexes(rName, attr1, attr2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"
  range_key      = %[2]q

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  local_secondary_index {
    name            = "lsiName"
    range_key       = %[3]q
    projection_type = "KEYS_ONLY"
  }
}
`, rName, attr1, attr2)
}

func testAccAWSDynamoDbTableConfigReplica0(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName))
}

func testAccAWSDynamoDbTableConfigReplica1(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(3), // Prevent "Provider configuration not present" errors
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccAWSDynamoDbTableConfigReplica2(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(3),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "third" {
  provider = "awsthird"
}

resource "aws_dynamodb_table" "test" {
  name             = %[1]q
  hash_key         = "TestTableHashKey"
  billing_mode     = "PAY_PER_REQUEST"
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.alternate.name
  }

  replica {
    region_name = data.aws_region.third.name
  }
}
`, rName))
}

func testAccAWSDynamoDbConfigLSI(rName, lsiName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "staticHashKey"
  range_key      = "staticRangeKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  attribute {
    name = "staticRangeKey"
    type = "S"
  }

  attribute {
    name = "staticLSIRangeKey"
    type = "S"
  }

  local_secondary_index {
    name            = "%s"
    range_key       = "staticLSIRangeKey"
    projection_type = "KEYS_ONLY"
  }
}
`, rName, lsiName)
}
