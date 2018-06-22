package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSS3BucketReplication_basic(t *testing.T) {
	name := fmt.Sprintf("tf-test-bucket-%d", acctest.RandInt())

	// record the initialized providers so that we can use them to check for the instances in each region
	var providers []*schema.Provider

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckWithProviders(testAccCheckAWSS3BucketDestroyWithProvider, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3BucketReplicationConfigBase(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_s3_bucket_replication.replication", "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket_replication.replication", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketHasReplicationWithProvider("aws_s3_bucket.main", name+"-replica", testAccAwsRegionProviderFunc("us-west-2", &providers)),
				),
			}, {
				Config: testAccAWSS3BucketReplicationConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_s3_bucket_replication.replication", "replication_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_s3_bucket_replication.replication", "replication_configuration.0.rules.#", "1"),
					testAccCheckAWSS3BucketHasReplicationWithProvider("aws_s3_bucket.main", name+"-replica-2", testAccAwsRegionProviderFunc("us-west-2", &providers)),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketHasReplicationWithProvider(n, b string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*AWSClient).s3conn
		rc, err := conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("GetBucketReplication error: %v", err)
		}

		destinationBucket := *rc.ReplicationConfiguration.Rules[0].Destination.Bucket
		if strings.Contains(destinationBucket, b) {
			return nil
		}
		return fmt.Errorf("No replication bucket found")
	}
}

func testAccAWSS3BucketReplicationConfig_basePrefix(bucketName string) string {
	return fmt.Sprintf(`
provider "aws" {
  alias = "usw2"
  region = "us-west-2"
}

provider "aws" {
  alias = "use2"
  region = "us-east-2"
}

resource "aws_s3_bucket" "main" {
  provider = "aws.usw2"
  bucket = "%s"
  region = "us-west-2"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination" {
  provider = "aws.use2"
  region = "us-east-2"
  bucket = "%s-replica"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket" "destination2" {
  provider = "aws.use2"
  region = "us-east-2"
  bucket = "%s-replica-2"

  versioning {
    enabled = true
  }
}

resource "aws_iam_role" "replication" {
  name = "%s-role"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_policy" "replication" {
  name = "%s-role-policy"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:GetReplicationConfiguration",
        "s3:ListBucket"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.main.arn}"
      ]
    },
    {
      "Action": [
        "s3:GetObjectVersion",
        "s3:GetObjectVersionAcl"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.main.arn}/*"
      ]
    },
    {
      "Action": [
        "s3:ReplicateObject",
        "s3:ReplicateDelete"
      ],
      "Effect": "Allow",
      "Resource": [
        "${aws_s3_bucket.destination.arn}/*",
        "${aws_s3_bucket.destination.arn}-2/*"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_policy_attachment" "replication" {
  name       = "%s-role-policy-attachment"
  roles      = ["${aws_iam_role.replication.name}"]
  policy_arn = "${aws_iam_policy.replication.arn}"
}
`, bucketName, bucketName, bucketName, bucketName, bucketName, bucketName)
}

func testAccAWSS3BucketReplicationConfigBase(bucketName string) string {
	return testAccAWSS3BucketReplicationConfig_basePrefix(bucketName) + `
resource "aws_s3_bucket_replication" "replication" {
  provider = "aws.usw2"
  bucket = "${aws_s3_bucket.main.id}"

  replication_configuration {
    role = "${aws_iam_role.replication.arn}"
    rules {
      id     = "foobar"
      prefix = "foo"
      status = "Enabled"

      destination {
        bucket        = "${aws_s3_bucket.destination.arn}"
        storage_class = "STANDARD"
      }
    }
  }
}
`
}

func testAccAWSS3BucketReplicationConfigUpdated(bucketName string) string {
	return testAccAWSS3BucketReplicationConfig_basePrefix(bucketName) + `
resource "aws_s3_bucket_replication" "replication" {
  provider = "aws.usw2"
  bucket = "${aws_s3_bucket.main.id}"

  replication_configuration {
    role = "${aws_iam_role.replication.arn}"
    rules {
      id     = "foobar"
      prefix = ""
      status = "Enabled"

      destination {
        bucket        = "${aws_s3_bucket.destination2.arn}"
        storage_class = "STANDARD"
      }
    }
  }
}
`
}
