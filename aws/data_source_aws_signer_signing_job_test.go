package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSSignerSigningJob_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_signer_signing_job.test"
	resourceName := "aws_signer_signing_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSSignerSigningJobConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "job_owner", resourceName, "job_owner"),
					resource.TestCheckResourceAttrPair(dataSourceName, "job_invoker", resourceName, "job_invoker"),
					resource.TestCheckResourceAttrPair(dataSourceName, "profile_name", resourceName, "profile_name"),
				),
			},
		},
	})
}

func testAccDataSourceAWSSignerSigningJobConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_s3_bucket" "source" {
  bucket = "%[1]s-source"

  versioning {
    enabled = true
  }

  force_destroy = true
}

resource "aws_s3_bucket" "destination" {
  bucket        = "%[1]s-destination"
  force_destroy = true
}

resource "aws_s3_bucket_object" "source" {
  bucket = aws_s3_bucket.source.bucket
  key    = "lambdatest.zip"
  source = "test-fixtures/lambdatest.zip"
}

resource "aws_signer_signing_job" "test" {
  profile_name = aws_signer_signing_profile.test.name

  source {
    s3 {
      bucket  = aws_s3_bucket_object.source.bucket
      key     = aws_s3_bucket_object.source.key
      version = aws_s3_bucket_object.source.version_id
    }
  }

  destination {
    s3 {
      bucket = aws_s3_bucket.destination.bucket
    }
  }
}

data "aws_signer_signing_job" "test" {
  job_id = aws_signer_signing_job.test.job_id
}
`, rName)
}
