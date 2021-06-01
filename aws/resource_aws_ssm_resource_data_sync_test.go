package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSsmResourceDataSync_basic(t *testing.T) {
	resourceName := "aws_ssm_resource_data_sync.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSsmResourceDataSyncDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSsmResourceDataSyncConfig(acctest.RandInt(), acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSsmResourceDataSyncExists(resourceName),
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

func TestAccAWSSsmResourceDataSync_update(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_ssm_resource_data_sync.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSsmResourceDataSyncDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSsmResourceDataSyncConfig(acctest.RandInt(), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSsmResourceDataSyncExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSsmResourceDataSyncConfigUpdate(acctest.RandInt(), rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSsmResourceDataSyncExists(resourceName),
				),
			},
		},
	})
}

func testAccCheckAWSSsmResourceDataSyncDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_resource_data_sync" {
			continue
		}
		syncItem, err := findResourceDataSyncItem(conn, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}
		if syncItem != nil {
			return fmt.Errorf("Resource Data Sync (%s) found", rs.Primary.Attributes["name"])
		}
	}
	return nil
}

func testAccCheckAWSSsmResourceDataSyncExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Println(s.RootModule().Resources)
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		return nil
	}
}

func testAccSsmResourceDataSyncConfig(rInt int, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "hoge" {
  bucket        = "tf-test-bucket-%[1]d"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "hoge" {
  bucket = aws_s3_bucket.hoge.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMBucketPermissionsCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d"
    },
    {
      "Sid": " SSMBucketDelivery",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
      EOF

}

resource "aws_ssm_resource_data_sync" "test" {
  name = "tf-test-ssm-%[2]s"

  s3_destination {
    bucket_name = aws_s3_bucket.hoge.bucket
    region      = aws_s3_bucket.hoge.region
  }
}
`, rInt, rName)
}

func testAccSsmResourceDataSyncConfigUpdate(rInt int, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "hoge" {
  bucket        = "tf-test-bucket-%[1]d"
  force_destroy = true
}

data "aws_partition" "current" {}

resource "aws_s3_bucket_policy" "hoge" {
  bucket = aws_s3_bucket.hoge.bucket

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMBucketPermissionsCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d"
    },
    {
      "Sid": " SSMBucketDelivery",
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::tf-test-bucket-%[1]d/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
      EOF

}

resource "aws_ssm_resource_data_sync" "test" {
  name = "tf-test-ssm-%[2]s"

  s3_destination {
    bucket_name = aws_s3_bucket.hoge.bucket
    region      = aws_s3_bucket.hoge.region
    prefix      = "test-"
  }
}
`, rInt, rName)
}
