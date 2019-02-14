package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSS3BucketInventory_basic(t *testing.T) {
	var conf s3.InventoryConfiguration
	rString := acctest.RandString(8)
	resourceName := "aws_s3_bucket_inventory.test"

	bucketName := fmt.Sprintf("tf-acc-bucket-inventory-%s", rString)
	inventoryName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketInventoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketInventoryConfig(bucketName, inventoryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketInventoryConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "bucket", bucketName),
					resource.TestCheckNoResourceAttr(resourceName, "filter"),
					resource.TestCheckResourceAttr(resourceName, "name", inventoryName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "included_object_versions", "All"),

					resource.TestCheckResourceAttr(resourceName, "optional_fields.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.frequency", "Weekly"),

					resource.TestCheckResourceAttr(resourceName, "destination.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.bucket_arn", "arn:aws:s3:::"+bucketName),
					resource.TestCheckResourceAttrSet(resourceName, "destination.0.bucket.0.account_id"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.format", "ORC"),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.prefix", "inventory"),
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

func TestAccAWSS3BucketInventory_encryptWithSSES3(t *testing.T) {
	var conf s3.InventoryConfiguration
	rString := acctest.RandString(8)
	resourceName := "aws_s3_bucket_inventory.test"

	bucketName := fmt.Sprintf("tf-acc-bucket-inventory-%s", rString)
	inventoryName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketInventoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketInventoryConfigEncryptWithSSES3(bucketName, inventoryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketInventoryConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.encryption.0.sse_s3.#", "1"),
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

func TestAccAWSS3BucketInventory_encryptWithSSEKMS(t *testing.T) {
	var conf s3.InventoryConfiguration
	rString := acctest.RandString(8)
	resourceName := "aws_s3_bucket_inventory.test"

	bucketName := fmt.Sprintf("tf-acc-bucket-inventory-%s", rString)
	inventoryName := t.Name()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketInventoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketInventoryConfigEncryptWithSSEKMS(bucketName, inventoryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketInventoryConfigExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination.0.bucket.0.encryption.0.sse_kms.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "destination.0.bucket.0.encryption.0.sse_kms.0.key_id", regexp.MustCompile("^arn:aws:kms:")),
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

func testAccCheckAWSS3BucketInventoryConfigExists(n string, res *s3.InventoryConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 bucket inventory configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).s3conn
		bucket, name, err := resourceAwsS3BucketInventoryParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &s3.GetBucketInventoryConfigurationInput{
			Bucket: aws.String(bucket),
			Id:     aws.String(name),
		}
		log.Printf("[DEBUG] Reading S3 bucket inventory configuration: %s", input)
		output, err := conn.GetBucketInventoryConfiguration(input)
		if err != nil {
			return err
		}

		*res = *output.InventoryConfiguration

		return nil
	}
}

func testAccCheckAWSS3BucketInventoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_bucket_inventory" {
			continue
		}

		bucket, name, err := resourceAwsS3BucketInventoryParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		err = resource.Retry(1*time.Minute, func() *resource.RetryError {
			input := &s3.GetBucketInventoryConfigurationInput{
				Bucket: aws.String(bucket),
				Id:     aws.String(name),
			}
			log.Printf("[DEBUG] Reading S3 bucket inventory configuration: %s", input)
			output, err := conn.GetBucketInventoryConfiguration(input)
			if err != nil {
				if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "NoSuchConfiguration", "The specified configuration does not exist.") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			if output.InventoryConfiguration != nil {
				return resource.RetryableError(fmt.Errorf("S3 bucket inventory configuration exists: %v", output))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func testAccAWSS3BucketInventoryConfigBucket(name string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "%s"
  acl    = "private"
}
`, name)
}

func testAccAWSS3BucketInventoryConfig(bucketName, inventoryName string) string {
	return fmt.Sprintf(`
%s
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket_inventory" "test" {
  bucket = "${aws_s3_bucket.bucket.id}"
  name   = "%s"

  included_object_versions = "All"

  optional_fields = [
    "Size",
    "LastModifiedDate",
  ]

  filter {
    prefix = "documents/"
  }

  schedule {
    frequency = "Weekly"
  }

  destination {
    bucket {
      format = "ORC"
      bucket_arn = "${aws_s3_bucket.bucket.arn}"
      account_id = "${data.aws_caller_identity.current.account_id}"
      prefix = "inventory"
    }
  }
}
`, testAccAWSS3BucketInventoryConfigBucket(bucketName), inventoryName)
}

func testAccAWSS3BucketInventoryConfigEncryptWithSSES3(bucketName, inventoryName string) string {
	return fmt.Sprintf(`
%s
resource "aws_s3_bucket_inventory" "test" {
  bucket = "${aws_s3_bucket.bucket.id}"
  name   = "%s"

  included_object_versions = "Current"

  schedule {
    frequency = "Daily"
  }

  destination {
    bucket {
      format = "CSV"
      bucket_arn = "${aws_s3_bucket.bucket.arn}"

      encryption {
        sse_s3 {}
      }
    }
  }
}
`, testAccAWSS3BucketInventoryConfigBucket(bucketName), inventoryName)
}

func testAccAWSS3BucketInventoryConfigEncryptWithSSEKMS(bucketName, inventoryName string) string {
	return fmt.Sprintf(`
%s
resource "aws_kms_key" "inventory" {
  description             = "Terraform acc test S3 inventory SSE-KMS encryption: %s"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_inventory" "test" {
  bucket = "${aws_s3_bucket.bucket.id}"
  name   = "%s"

  included_object_versions = "Current"

  schedule {
    frequency = "Daily"
  }

  destination {
    bucket {
      format = "Parquet"
      bucket_arn = "${aws_s3_bucket.bucket.arn}"

      encryption {
        sse_kms {
          key_id = "${aws_kms_key.inventory.arn}"
        }
      }
    }
  }
}
`, testAccAWSS3BucketInventoryConfigBucket(bucketName), bucketName, inventoryName)
}
