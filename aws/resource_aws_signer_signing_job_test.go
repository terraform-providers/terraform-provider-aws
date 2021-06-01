package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSignerSigningJob_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_signer_signing_job.test"
	profileResourceName := "aws_signer_signing_profile.test"

	var job signer.DescribeSigningJobOutput
	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningJobConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(profileResourceName, &conf),
					testAccCheckAWSSignerSigningJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
					resource.TestCheckResourceAttr(resourceName, "platform_display_name", "AWS Lambda"),
					resource.TestCheckResourceAttr(resourceName, "status", "Succeeded"),
				),
			},
		},
	})

}

func testAccAWSSignerSigningJobConfig(rName string) string {
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
  bucket        = "%[1]s"
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
`, rName)
}

func testAccCheckAWSSignerSigningJobExists(res string, job *signer.DescribeSigningJobOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing job not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing job with that ID does not exist")
		}

		conn := testAccProvider.Meta().(*AWSClient).signerconn

		params := &signer.DescribeSigningJobInput{
			JobId: aws.String(rs.Primary.ID),
		}

		getJob, err := conn.DescribeSigningJob(params)
		if err != nil {
			return err
		}

		*job = *getJob

		return nil
	}
}
