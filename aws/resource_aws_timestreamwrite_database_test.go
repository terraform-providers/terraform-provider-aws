package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
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
	resource.AddTestSweepers("aws_timestreamwrite_database", &resource.Sweeper{
		Name:         "aws_timestreamwrite_database",
		F:            testSweepTimestreamWriteDatabases,
		Dependencies: []string{"aws_timestreamwrite_table"},
	})
}

func testSweepTimestreamWriteDatabases(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).timestreamwriteconn
	ctx := context.Background()

	var sweeperErrs *multierror.Error

	input := &timestreamwrite.ListDatabasesInput{}

	err = conn.ListDatabasesPagesWithContext(ctx, input, func(page *timestreamwrite.ListDatabasesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, database := range page.Databases {
			if database == nil {
				continue
			}

			dbName := aws.StringValue(database.DatabaseName)

			log.Printf("[INFO] Deleting Timestream Database (%s)", dbName)
			r := resourceAwsTimestreamWriteDatabase()
			d := r.Data(nil)
			d.SetId(dbName)

			diags := r.DeleteWithoutTimeout(ctx, d, client)

			if diags != nil && diags.HasError() {
				for _, d := range diags {
					if d.Severity == diag.Error {
						sweeperErr := fmt.Errorf("error deleting Timestream Database (%s): %s", dbName, d.Summary)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					}
				}
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Timestream Database sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Timestream Databases: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSTimestreamWriteDatabase_basic(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "timestream", fmt.Sprintf("database/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
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

func TestAccAWSTimestreamWriteDatabase_disappears(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsTimestreamWriteDatabase(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSTimestreamWriteDatabase_kmsKey(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigKmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "database_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsResourceName, "arn"),
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

func TestAccAWSTimestreamWriteDatabase_updateKmsKey(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
				),
			},
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigKmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
				),
			},
		},
	})
}

func TestAccAWSTimestreamWriteDatabase_Tags(t *testing.T) {
	resourceName := "aws_timestreamwrite_database.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTimestreamWrite(t) },
		ErrorCheck:   testAccErrorCheck(t, timestreamwrite.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTimestreamWriteDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccAWSTimestreamWriteDatabaseConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTimestreamWriteDatabaseExists(resourceName),
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

func testAccCheckAWSTimestreamWriteDatabaseDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).timestreamwriteconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_timestreamwrite_database" {
			continue
		}

		output, err := conn.DescribeDatabase(&timestreamwrite.DescribeDatabaseInput{
			DatabaseName: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.Database != nil {
			return fmt.Errorf("Timestream Database (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSTimestreamWriteDatabaseExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).timestreamwriteconn

		output, err := conn.DescribeDatabase(&timestreamwrite.DescribeDatabaseInput{
			DatabaseName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if output == nil || output.Database == nil {
			return fmt.Errorf("Timestream Database (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckAWSTimestreamWrite(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).timestreamwriteconn

	input := &timestreamwrite.ListDatabasesInput{}

	_, err := conn.ListDatabases(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSTimestreamWriteDatabaseConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}
`, rName)
}

func testAccAWSTimestreamWriteDatabaseConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSTimestreamWriteDatabaseConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSTimestreamWriteDatabaseConfigKmsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Effect": "Allow",
     "Principal": {
       "AWS": "*"
     },
     "Action": "kms:*",
     "Resource": "*"
   }
 ]
}
POLICY
}

resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
  kms_key_id    = aws_kms_key.test.arn
}
`, rName)
}
