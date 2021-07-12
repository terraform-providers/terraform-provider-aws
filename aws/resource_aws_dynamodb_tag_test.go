package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccAWSDynamodbTag_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dynamodb_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dynamodb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDynamodbTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamodbTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
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

func TestAccAWSDynamodbTag_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dynamodb_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dynamodb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDynamodbTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamodbTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbTagExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDynamodbTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/13725
func TestAccAWSDynamodbTag_ResourceArn_TableReplica(t *testing.T) {
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dynamodb_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ErrorCheck:        testAccErrorCheck(t, dynamodb.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckDynamodbTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamodbTagConfigResourceArnTableReplica(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbTagExists(resourceName),
				),
			},
			{
				Config:            testAccDynamodbTagConfigResourceArnTableReplica(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDynamodbTag_Value(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dynamodb_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dynamodb.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDynamodbTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDynamodbTagConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDynamodbTagConfig(rName, "key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDynamodbTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "value", "value1updated"),
				),
			},
		},
	})
}

func testAccDynamodbTagConfig(rName string, key string, value string) string {
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

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_dynamodb_tag" "test" {
  resource_arn = aws_dynamodb_table.test.arn
  key          = %[2]q
  value        = %[3]q
}
`, rName, key, value)
}

func testAccDynamodbTagConfigResourceArnTableReplica(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(2),
		fmt.Sprintf(`
data "aws_region" "alternate" {
  provider = "awsalternate"
}

data "aws_region" "current" {}

resource "aws_dynamodb_table" "test" {
  provider = "awsalternate"

  billing_mode     = "PAY_PER_REQUEST"
  hash_key         = "TestTableHashKey"
  name             = %[1]q
  stream_enabled   = true
  stream_view_type = "NEW_AND_OLD_IMAGES"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  replica {
    region_name = data.aws_region.current.name
  }
}

resource "aws_dynamodb_tag" "test" {
  resource_arn = replace(aws_dynamodb_table.test.arn, data.aws_region.alternate.name, data.aws_region.current.name)
  key          = "testkey"
  value        = "testvalue"
}
`, rName))
}
