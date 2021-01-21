package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceS3Bucket_basic(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	region := testAccGetRegion()
	hostedZoneID, _ := HostedZoneIDForRegion(region)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket.bucket"),
					resource.TestCheckResourceAttrPair("data.aws_s3_bucket.bucket", "arn", "aws_s3_bucket.bucket", "arn"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "region", region),
					testAccCheckS3BucketDomainName("data.aws_s3_bucket.bucket", "bucket_domain_name", bucketName),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "bucket_regional_domain_name", testAccBucketRegionalDomainName(bucketName, region)),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "hosted_zone_id", hostedZoneID),
					resource.TestCheckNoResourceAttr("data.aws_s3_bucket.bucket", "website_endpoint"),
				),
			},
		},
	})
}

func TestAccDataSourceS3Bucket_website(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	region := testAccGetRegion()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketWebsiteConfig(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite("data.aws_s3_bucket.bucket", "index.html", "error.html", "", ""),
					testAccCheckS3BucketWebsiteEndpoint("data.aws_s3_bucket.bucket", "website_endpoint", bucketName, region),
				),
			},
		},
	})
}

func testAccAWSDataSourceS3BucketConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

data "aws_s3_bucket" "bucket" {
  bucket = aws_s3_bucket.bucket.id
}
`, bucketName)
}

func testAccAWSDataSourceS3BucketWebsiteConfig(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}

data "aws_s3_bucket" "bucket" {
  bucket = aws_s3_bucket.bucket.id
}
`, bucketName)
}
