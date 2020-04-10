package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceS3Bucket_basic(t *testing.T) {
	rInt := acctest.RandInt()
	arnRegexp := regexp.MustCompile(`^arn:aws[\w-]*:s3:::`)
	region := testAccGetRegion()
	hostedZoneID, _ := HostedZoneIDForRegion(region)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket.bucket"),
					resource.TestMatchResourceAttr("data.aws_s3_bucket.bucket", "arn", arnRegexp),
					resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "region", region),
					testAccCheckS3BucketDomainName(
						"data.aws_s3_bucket.bucket", "bucket_domain_name", testAccBucketName(rInt)),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "bucket_regional_domain_name", testAccBucketRegionalDomainName(rInt, region)),
					resource.TestCheckResourceAttr(
						"data.aws_s3_bucket.bucket", "hosted_zone_id", hostedZoneID),
					resource.TestCheckNoResourceAttr("data.aws_s3_bucket.bucket", "website_endpoint"),
				),
			},
		},
	})
}

func TestAccDataSourceS3Bucket_website(t *testing.T) {
	rInt := acctest.RandInt()
	region := testAccGetRegion()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketWebsiteConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists("data.aws_s3_bucket.bucket"),
					testAccCheckAWSS3BucketWebsite(
						"data.aws_s3_bucket.bucket", "index.html", "error.html", "", ""),
					testAccCheckS3BucketWebsiteEndpoint(
						"data.aws_s3_bucket.bucket", "website_endpoint", testAccBucketName(rInt), region),
				),
			},
		},
	})
}

func testAccAWSDataSourceS3BucketConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
}

data "aws_s3_bucket" "bucket" {
  bucket = "${aws_s3_bucket.bucket.id}"
}
`, randInt)
}

func testAccAWSDataSourceS3BucketWebsiteConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}

data "aws_s3_bucket" "bucket" {
  bucket = "${aws_s3_bucket.bucket.id}"
}
`, randInt)
}
