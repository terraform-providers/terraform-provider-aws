package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	prefixes := []string{
		"TerraformTest",
		"tf-acc-test-",
		"tf-autoscaled-table-",
	}

	err = conn.ListTablesPages(&dynamodb.ListTablesInput{}, func(out *dynamodb.ListTablesOutput, lastPage bool) bool {
		for _, tableName := range out.TableNames {
			for _, prefix := range prefixes {
				if !strings.HasPrefix(*tableName, prefix) {
					log.Printf("[INFO] Skipping DynamoDB Table: %s", *tableName)
					continue
				}
			}
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
		{ // No-op
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
					"non_key_attributes": []interface{}{"RandomAttribute"},
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
					"non_key_attributes": []interface{}{"RandomAttribute"},
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

func TestAccAWSDynamoDbTable_importBasic(t *testing.T) {
	resourceName := "aws_dynamodb_table.basic-dynamodb-table"

	rName := acctest.RandomWithPrefix("TerraformTestTable-")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDynamoDbTable_importTags(t *testing.T) {
	resourceName := "aws_dynamodb_table.basic-dynamodb-table"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigTags(),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDynamoDbTable_importTimeToLive(t *testing.T) {
	resourceName := "aws_dynamodb_table.basic-dynamodb-table"
	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigAddTimeToLive(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDynamoDbTable_basic(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "name", rName),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "hash_key", "TestTableHashKey"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "attribute.2990477658.name", "TestTableHashKey"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "attribute.2990477658.type", "S"),
				),
			},
		},
	})
}
func TestAccAWSDynamoDbTable_extended(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					testAccCheckInitialAWSDynamoDbTableConf("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigAddSecondaryGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbTableWasUpdated("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_enablePitr(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					testAccCheckInitialAWSDynamoDbTableConf("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfig_backup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbTableHasPointInTimeRecoveryEnabled("aws_dynamodb_table.basic-dynamodb-table"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "point_in_time_recovery.#", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "point_in_time_recovery.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_BillingMode(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					testAccCheckInitialAWSDynamoDbTableConf("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequest(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbTableHasBilling_PayPerRequest("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
			{
				Config: testAccAWSDynamoDbBilling_PayPerRequestWithGSI(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbTableHasBilling_PayPerRequest("aws_dynamodb_table.basic-dynamodb-table"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "global_secondary_index.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_dynamodb_table.basic-dynamodb-table", "global_secondary_index.427641302.name", "TestTableGSI"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_streamSpecification(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	tableName := fmt.Sprintf("TerraformTestStreamTable-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigStreamSpecification(tableName, true, "KEYS_ONLY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "stream_enabled", "true"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "stream_view_type", "KEYS_ONLY"),
					resource.TestCheckResourceAttrSet("aws_dynamodb_table.basic-dynamodb-table", "stream_arn"),
					resource.TestCheckResourceAttrSet("aws_dynamodb_table.basic-dynamodb-table", "stream_label"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigStreamSpecification(tableName, false, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "stream_enabled", "false"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "stream_view_type", ""),
					resource.TestCheckResourceAttrSet("aws_dynamodb_table.basic-dynamodb-table", "stream_arn"),
					resource.TestCheckResourceAttrSet("aws_dynamodb_table.basic-dynamodb-table", "stream_label"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
					testAccCheckInitialAWSDynamoDbTableConf("aws_dynamodb_table.basic-dynamodb-table"),
					resource.TestCheckResourceAttr(
						"aws_dynamodb_table.basic-dynamodb-table", "tags.%", "3"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/13243
func TestAccAWSDynamoDbTable_gsiUpdateCapacity(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedCapacity(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.705130498.read_capacity", "2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.705130498.write_capacity", "2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1115936309.read_capacity", "2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1115936309.write_capacity", "2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.4212014188.read_capacity", "2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.4212014188.write_capacity", "2"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_gsiUpdateOtherAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2726077800.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.hash_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.3405251423.write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.hash_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.range_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.non_key_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.non_key_attributes.0", "RandomAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.range_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.write_capacity", "1"),
				),
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/566
func TestAccAWSDynamoDbTable_gsiUpdateNonKeyAttributes(t *testing.T) {
	var conf dynamodb.DescribeTableOutput
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedOtherAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.hash_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.range_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.non_key_attributes.#", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.non_key_attributes.0", "RandomAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.range_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.2311632778.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.write_capacity", "1"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigGsiUpdatedNonKeyAttributes(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.test", &conf),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.#", "3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.hash_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.name", "att2-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.range_key", "att2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1182392663.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.hash_key", "att3"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.name", "att3-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.non_key_attributes.#", "2"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.non_key_attributes.0", "RandomAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.non_key_attributes.1", "AnotherAttribute"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.projection_type", "INCLUDE"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.range_key", "att4"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.102175821.write_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.hash_key", "att1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.name", "att1-index"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.non_key_attributes.#", "0"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.projection_type", "ALL"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.range_key", ""),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.read_capacity", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.test", "global_secondary_index.1937107206.write_capacity", "1"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_ttl(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigAddTimeToLive(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamoDbTableTimeToLiveWasUpdated("aws_dynamodb_table.basic-dynamodb-table"),
				),
			},
		},
	})
}

func TestAccAWSDynamoDbTable_attributeUpdate(t *testing.T) {
	var conf dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigOneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
				),
			},
			{ // Attribute type change
				Config: testAccAWSDynamoDbConfigOneAttribute(rName, "firstKey", "firstKey", "N"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
				),
			},
			{ // New attribute addition (index update)
				Config: testAccAWSDynamoDbConfigTwoAttributes(rName, "firstKey", "secondKey", "firstKey", "N", "secondKey", "S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
				),
			},
			{ // Attribute removal (index update)
				Config: testAccAWSDynamoDbConfigOneAttribute(rName, "firstKey", "firstKey", "S"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &conf),
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
		},
	})
}

func testAccCheckDynamoDbTableTimeToLiveWasUpdated(n string) resource.TestCheckFunc {
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

		params := &dynamodb.DescribeTimeToLiveInput{
			TableName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeTimeToLive(params)

		if err != nil {
			return fmt.Errorf("Problem describing time to live for table '%s': %s", rs.Primary.ID, err)
		}

		ttlDescription := resp.TimeToLiveDescription

		log.Printf("[DEBUG] Checking on table %s", rs.Primary.ID)

		if *ttlDescription.TimeToLiveStatus != dynamodb.TimeToLiveStatusEnabled {
			return fmt.Errorf("TimeToLiveStatus %s, not ENABLED!", *ttlDescription.TimeToLiveStatus)
		}

		if *ttlDescription.AttributeName != "TestTTL" {
			return fmt.Errorf("AttributeName was %s, not TestTTL!", *ttlDescription.AttributeName)
		}

		return nil
	}
}

func TestAccAWSDynamoDbTable_encryption(t *testing.T) {
	var confEncEnabled, confEncDisabled, confBasic dynamodb.DescribeTableOutput

	rName := acctest.RandomWithPrefix("TerraformTestTable-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDynamoDbTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDynamoDbConfigInitialStateWithEncryption(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &confEncEnabled),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "server_side_encryption.0.enabled", "true"),
				),
			},
			{
				Config: testAccAWSDynamoDbConfigInitialStateWithEncryption(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &confEncDisabled),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "server_side_encryption.#", "0"),
					func(s *terraform.State) error {
						if confEncDisabled.Table.CreationDateTime.Equal(*confEncEnabled.Table.CreationDateTime) {
							return fmt.Errorf("DynamoDB table not recreated when changing SSE")
						}
						return nil
					},
				),
			},
			{
				Config: testAccAWSDynamoDbConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInitialAWSDynamoDbTableExists("aws_dynamodb_table.basic-dynamodb-table", &confBasic),
					resource.TestCheckResourceAttr("aws_dynamodb_table.basic-dynamodb-table", "server_side_encryption.#", "0"),
					func(s *terraform.State) error {
						if !confBasic.Table.CreationDateTime.Equal(*confEncDisabled.Table.CreationDateTime) {
							return fmt.Errorf("DynamoDB table was recreated unexpectedly")
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

func testAccCheckInitialAWSDynamoDbTableConf(n string) resource.TestCheckFunc {
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

		table := resp.Table

		log.Printf("[DEBUG] Checking on table %s", rs.Primary.ID)

		if table.BillingModeSummary != nil && aws.StringValue(table.BillingModeSummary.BillingMode) != dynamodb.BillingModeProvisioned {
			return fmt.Errorf("Billing Mode was %s, not %s!", aws.StringValue(table.BillingModeSummary.BillingMode), dynamodb.BillingModeProvisioned)
		}

		if *table.ProvisionedThroughput.WriteCapacityUnits != 2 {
			return fmt.Errorf("Provisioned write capacity was %d, not 2!", table.ProvisionedThroughput.WriteCapacityUnits)
		}

		if *table.ProvisionedThroughput.ReadCapacityUnits != 1 {
			return fmt.Errorf("Provisioned read capacity was %d, not 1!", table.ProvisionedThroughput.ReadCapacityUnits)
		}

		if table.SSEDescription != nil && *table.SSEDescription.Status != dynamodb.SSEStatusDisabled {
			return fmt.Errorf("SSE status was %s, not %s", *table.SSEDescription.Status, dynamodb.SSEStatusDisabled)
		}

		attrCount := len(table.AttributeDefinitions)
		gsiCount := len(table.GlobalSecondaryIndexes)
		lsiCount := len(table.LocalSecondaryIndexes)

		if attrCount != 4 {
			return fmt.Errorf("There were %d attributes, not 4 like there should have been!", attrCount)
		}

		if gsiCount != 1 {
			return fmt.Errorf("There were %d GSIs, not 1 like there should have been!", gsiCount)
		}

		if lsiCount != 1 {
			return fmt.Errorf("There were %d LSIs, not 1 like there should have been!", lsiCount)
		}

		attrmap := dynamoDbAttributesToMap(&table.AttributeDefinitions)
		if attrmap["TestTableHashKey"] != "S" {
			return fmt.Errorf("Test table hash key was of type %s instead of S!", attrmap["TestTableHashKey"])
		}
		if attrmap["TestTableRangeKey"] != "S" {
			return fmt.Errorf("Test table range key was of type %s instead of S!", attrmap["TestTableRangeKey"])
		}
		if attrmap["TestLSIRangeKey"] != "N" {
			return fmt.Errorf("Test table LSI range key was of type %s instead of N!", attrmap["TestLSIRangeKey"])
		}
		if attrmap["TestGSIRangeKey"] != "S" {
			return fmt.Errorf("Test table GSI range key was of type %s instead of S!", attrmap["TestGSIRangeKey"])
		}

		return nil
	}
}

func testAccCheckDynamoDbTableHasPointInTimeRecoveryEnabled(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table name specified!")
		}

		conn := testAccProvider.Meta().(*AWSClient).dynamodbconn

		resp, err := conn.DescribeContinuousBackups(&dynamodb.DescribeContinuousBackupsInput{
			TableName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		pitr := resp.ContinuousBackupsDescription.PointInTimeRecoveryDescription
		status := *pitr.PointInTimeRecoveryStatus
		if status != dynamodb.PointInTimeRecoveryStatusEnabled {
			return fmt.Errorf("Point in time backup had a status of %s rather than enabled", status)
		}

		return nil
	}
}

func testAccCheckDynamoDbTableHasBilling_PayPerRequest(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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
			return err
		}
		table := resp.Table

		if table.BillingModeSummary == nil {
			return fmt.Errorf("Billing Mode summary was empty, expected summary to exist and contain billing mode %s", dynamodb.BillingModePayPerRequest)
		} else if aws.StringValue(table.BillingModeSummary.BillingMode) != dynamodb.BillingModePayPerRequest {
			return fmt.Errorf("Billing Mode was %s, not %s!", aws.StringValue(table.BillingModeSummary.BillingMode), dynamodb.BillingModePayPerRequest)

		}

		return nil
	}
}

func testAccCheckDynamoDbTableWasUpdated(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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
		table := resp.Table

		if err != nil {
			return err
		}

		attrCount := len(table.AttributeDefinitions)
		gsiCount := len(table.GlobalSecondaryIndexes)
		lsiCount := len(table.LocalSecondaryIndexes)

		if attrCount != 4 {
			return fmt.Errorf("There were %d attributes, not 4 like there should have been!", attrCount)
		}

		if gsiCount != 1 {
			return fmt.Errorf("There were %d GSIs, not 1 like there should have been!", gsiCount)
		}

		if lsiCount != 1 {
			return fmt.Errorf("There were %d LSIs, not 1 like there should have been!", lsiCount)
		}

		if dynamoDbGetGSIIndex(&table.GlobalSecondaryIndexes, "ReplacementTestTableGSI") == -1 {
			return fmt.Errorf("Could not find GSI named 'ReplacementTestTableGSI' in the table!")
		}

		if dynamoDbGetGSIIndex(&table.GlobalSecondaryIndexes, "InitialTestTableGSI") != -1 {
			return fmt.Errorf("Should have removed 'InitialTestTableGSI' but it still exists!")
		}

		attrmap := dynamoDbAttributesToMap(&table.AttributeDefinitions)
		if attrmap["TestTableHashKey"] != "S" {
			return fmt.Errorf("Test table hash key was of type %s instead of S!", attrmap["TestTableHashKey"])
		}
		if attrmap["TestTableRangeKey"] != "S" {
			return fmt.Errorf("Test table range key was of type %s instead of S!", attrmap["TestTableRangeKey"])
		}
		if attrmap["TestLSIRangeKey"] != "N" {
			return fmt.Errorf("Test table LSI range key was of type %s instead of N!", attrmap["TestLSIRangeKey"])
		}
		if attrmap["ReplacementGSIRangeKey"] != "N" {
			return fmt.Errorf("Test table replacement GSI range key was of type %s instead of N!", attrmap["ReplacementGSIRangeKey"])
		}

		return nil
	}
}

func dynamoDbGetGSIIndex(gsiList *[]*dynamodb.GlobalSecondaryIndexDescription, target string) int {
	for idx, gsiObject := range *gsiList {
		if *gsiObject.IndexName == target {
			return idx
		}
	}

	return -1
}

func dynamoDbAttributesToMap(attributes *[]*dynamodb.AttributeDefinition) map[string]string {
	attrmap := make(map[string]string)

	for _, attrdef := range *attributes {
		attrmap[*attrdef.AttributeName] = *attrdef.AttributeType
	}

	return attrmap
}

func testAccAWSDynamoDbConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 1
  hash_key = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccAWSDynamoDbConfig_backup(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 1
  hash_key = "TestTableHashKey"

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
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  billing_mode = "PAY_PER_REQUEST"
  hash_key = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }
}
`, rName)
}

func testAccAWSDynamoDbBilling_PayPerRequestWithGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  billing_mode = "PAY_PER_REQUEST"
  hash_key = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  attribute {
    name = "TestTableGSIKey"
    type = "S"
  }

  global_secondary_index {
	name 			= "TestTableGSI"
	hash_key        = "TestTableGSIKey"
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigInitialState(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 2
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

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
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 1
    read_capacity = 1
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigInitialStateWithEncryption(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 1
  hash_key = "TestTableHashKey"

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

func testAccAWSDynamoDbConfigAddSecondaryGSI(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 2
  write_capacity = 2
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

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
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "ReplacementTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "ReplacementGSIRangeKey"
    write_capacity = 5
    read_capacity = 5
    projection_type = "INCLUDE"
    non_key_attributes = ["TestNonKeyAttribute"]
  }
}`, rName)
}

func testAccAWSDynamoDbConfigStreamSpecification(tableName string, enabled bool, viewType string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 2
  hash_key = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  stream_enabled = %t
  stream_view_type = "%s"
}
`, tableName, enabled, viewType)
}

func testAccAWSDynamoDbConfigTags() string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "TerraformTestTable-%d"
  read_capacity = 1
  write_capacity = 2
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

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
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 1
    read_capacity = 1
    projection_type = "KEYS_ONLY"
  }

  tags = {
    Name = "terraform-test-table-%d"
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
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
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
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
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
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
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
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
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
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
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
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    range_key       = "att4"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "INCLUDE"
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
  read_capacity  = "${var.capacity}"
  write_capacity = "${var.capacity}"
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
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att2-index"
    hash_key        = "att4"
    range_key       = "att2"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "att3-index"
    hash_key        = "att3"
    range_key       = "att4"
    write_capacity  = "${var.capacity}"
    read_capacity   = "${var.capacity}"
    projection_type = "INCLUDE"
    non_key_attributes = ["RandomAttribute", "AnotherAttribute"]
  }
}
`, name)
}

func testAccAWSDynamoDbConfigAddTimeToLive(rName string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 1
  write_capacity = 2
  hash_key = "TestTableHashKey"
  range_key = "TestTableRangeKey"

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
    name = "TestTableLSI"
    range_key = "TestLSIRangeKey"
    projection_type = "ALL"
  }

  ttl {
    attribute_name = "TestTTL"
    enabled = true
  }

  global_secondary_index {
    name = "InitialTestTableGSI"
    hash_key = "TestTableHashKey"
    range_key = "TestGSIRangeKey"
    write_capacity = 1
    read_capacity = 1
    projection_type = "KEYS_ONLY"
  }
}
`, rName)
}

func testAccAWSDynamoDbConfigOneAttribute(rName, hashKey, attrName, attrType string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 10
  write_capacity = 10
  hash_key = "staticHashKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }
  attribute {
    name = "%s"
    type = "%s"
  }

  global_secondary_index {
    name = "gsiName"
    hash_key = "%s"
    write_capacity = 10
    read_capacity = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName, attrName, attrType, hashKey)
}

func testAccAWSDynamoDbConfigTwoAttributes(rName, hashKey, rangeKey, attrName1, attrType1, attrName2, attrType2 string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "basic-dynamodb-table" {
  name = "%s"
  read_capacity = 10
  write_capacity = 10
  hash_key = "staticHashKey"

  attribute {
    name = "staticHashKey"
    type = "S"
  }

  attribute {
    name = "%s"
    type = "%s"
  }
  attribute {
    name = "%s"
    type = "%s"
  }

  global_secondary_index {
    name = "gsiName"
    hash_key = "%s"
    range_key = "%s"
    write_capacity = 10
    read_capacity = 10
    projection_type = "KEYS_ONLY"
  }
}
`, rName, attrName1, attrType1, attrName2, attrType2, hashKey, rangeKey)
}
