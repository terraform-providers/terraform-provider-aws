package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func TestAccAWSS3BucketVersioning_basic(t *testing.T) {
	rString := acctest.RandString(8)
	bucketName := fmt.Sprintf("tf-acc-s3bv-basic-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSS3BucketVersioningDestroy,
		Steps: []resource.TestStep{
			// step 0 - initialize a bucket with versioning and make sure it applies
			{
				Config: testAccAWSS3BucketVersioningConfig(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_s3_bucket_versioning.test", "bucket", bucketName),
					testAccAWSS3BucketVersioningCheckStatus(bucketName, true),
				),
			},
			// step 1 - disable versioning test
			{
				Config: testAccAWSS3BucketVersioningConfig(bucketName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_s3_bucket_versioning.test", "bucket", bucketName),
					testAccAWSS3BucketVersioningCheckStatus(bucketName, false),
				),
			},
			// step 2 - re-enable
			{
				Config: testAccAWSS3BucketVersioningConfig(bucketName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_s3_bucket_versioning.test", "bucket", bucketName),
					testAccAWSS3BucketVersioningCheckStatus(bucketName, true),
				),
			},
			// step 3 - test deleting
			{
				Config: testAccAWSS3BucketVersioningConfig_bucket(bucketName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_s3_bucket.test", "bucket", bucketName),
					testAccAWSS3BucketVersioningCheckStatus(bucketName, false),
				),
			},
		},
	})
}

func TestS3VersioningStatusToBool(t *testing.T) {
	var enabled *string = nil
	if s3VersioningStatusToBool(enabled) != false {
		t.Errorf("Expected s3VersioningStatusToBool to return false for nil input")
	}

	enabled = new(string)

	*enabled = s3.BucketVersioningStatusEnabled
	if s3VersioningStatusToBool(enabled) != true {
		t.Errorf("Expected s3VersioningStatusToBool to return true for s3.BucketVersioningStatusEnabled")
	}

	*enabled = s3.BucketVersioningStatusSuspended
	if s3VersioningStatusToBool(enabled) != false {
		t.Errorf("Expected s3VersioningStatusToBool to return false for s3.BucketVersioningStatusSuspended")
	}
}

func testAccAWSS3BucketVersioningDestroy(s *terraform.State) error {
	// nothing to do - we get destroyed along with the aws_s3_bucket
	return nil
}

func testAccAWSS3BucketVersioningCheckStatus(bucketName string, versioningEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).s3conn

		bucketVersioningStatus, err := conn.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			return fmt.Errorf("Error getting versioning config for %s: %s", bucketName, err)
		}

		var bucketVersioningEnabled bool
		if bucketVersioningStatus == nil {
			bucketVersioningEnabled = false
		} else {
			bucketVersioningEnabled = s3VersioningStatusToBool(bucketVersioningStatus.Status)
		}

		if bucketVersioningEnabled != versioningEnabled {
			return fmt.Errorf("Expected versioning for %s = %t but got %t", bucketName, versioningEnabled, bucketVersioningEnabled)
		}

		return nil
	}
}

func testAccAWSS3BucketVersioningConfig_bucket(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "%s"
  acl    = "private"
}
`, bucketName)
}

func testAccAWSS3BucketVersioningConfig(bucketName string, versioningEnabled bool) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_versioning" "test" {
  bucket  = "${aws_s3_bucket.test.id}"
  enabled = %t
}
`, testAccAWSS3BucketVersioningConfig_bucket(bucketName), versioningEnabled)
}
