package aws

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_timestreamwrite_table", &resource.Sweeper{
		Name: "aws_timestreamwrite_table",
		F:    testSweepTimestreamWriteTables,
	})
}

func testSweepTimestreamWriteTables(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).timestreamwriteconn
	ctx := context.Background()

	var sweeperErrs *multierror.Error

	input := &timestreamwrite.ListTablesInput{}

	err = conn.ListTablesPagesWithContext(ctx, input, func(page *timestreamwrite.ListTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, table := range page.Tables {
			if table == nil {
				continue
			}

			tableName := aws.StringValue(table.TableName)
			dbName := aws.StringValue(table.TableName)

			log.Printf("[INFO] Deleting Timestream Table (%s) from Database (%s)", tableName, dbName)
			r := resourceAwsTimestreamWriteTable()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", tableName, dbName))

			diags := r.DeleteWithoutTimeout(ctx, d, client)

			if diags != nil && diags.HasError() {
				for _, d := range diags {
					if d.Severity == diag.Error {
						sweeperErr := fmt.Errorf("error deleting Timestream Table (%s): %s", dbName, d.Summary)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Timestream Table sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Timestream Tables: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSTimestreamWriteTable_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_timestreamwrite_table.test"
	dbResourceName := "aws_timestreamwrite_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "timestream", fmt.Sprintf("database/%[1]s/table/%[1]s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "database_name", dbResourceName, "database_name"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "table_name", rName),
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

func TestAccAWSTimestreamWriteTable_disappears(t *testing.T) {
	resourceName := "aws_timestreamwrite_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsTimestreamWriteTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSTimestreamWriteTable_RetentionProperties(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteTableConfigRetentionProperties(rName, 30, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.magnetic_store_retention_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.memory_store_retention_period_in_hours", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSTimestreamWriteTableConfigRetentionProperties(rName, 300, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.magnetic_store_retention_period_in_days", "300"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.memory_store_retention_period_in_hours", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSTimestreamWriteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSTimestreamWriteTable_Tags(t *testing.T) {
	resourceName := "aws_timestreamwrite_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteTableConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccAWSTimestreamWriteTableConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccAWSTimestreamWriteTableConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
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

func testAccCheckAWSTimestreamWriteTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).timestreamwriteconn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_timestreamwrite_table" {
			continue
		}

		tableName, dbName, err := resourceAwsTimestreamWriteTableParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &timestreamwrite.DescribeTableInput{
			DatabaseName: aws.String(dbName),
			TableName:    aws.String(tableName),
		}

		output, err := conn.DescribeTableWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.Table != nil {
			return fmt.Errorf("Timestream Table (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSTimestreamWriteTableExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		tableName, dbName, err := resourceAwsTimestreamWriteTableParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).timestreamwriteconn

		input := &timestreamwrite.DescribeTableInput{
			DatabaseName: aws.String(dbName),
			TableName:    aws.String(tableName),
		}

		output, err := conn.DescribeTableWithContext(context.Background(), input)

		if err != nil {
			return err
		}

		if output == nil || output.Table == nil {
			return fmt.Errorf("Timestream Table (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSTimestreamWriteTableBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %q
}
`, rName)
}

func testAccAWSTimestreamWriteTableConfigBasic(rName string) string {
	return composeConfig(
		testAccAWSTimestreamWriteTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %q
}
`, rName))
}

func testAccAWSTimestreamWriteTableConfigRetentionProperties(rName string, magneticStoreDays, memoryStoreHours int) string {
	return composeConfig(
		testAccAWSTimestreamWriteTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %q

  retention_properties {
    magnetic_store_retention_period_in_days = %d
    memory_store_retention_period_in_hours  = %d
  }
}
`, rName, magneticStoreDays, memoryStoreHours))
}

func testAccAWSTimestreamWriteTableConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccAWSTimestreamWriteTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccAWSTimestreamWriteTableConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccAWSTimestreamWriteTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
