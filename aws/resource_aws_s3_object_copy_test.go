package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSS3ObjectCopy_basic(t *testing.T) {
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_s3_object_copy.test"
	sourceName := "aws_s3_bucket_object.source"
	key := "HundBegraven"
	sourceKey := "WshngtnNtnls"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3ObjectCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSS3ObjectCopyConfig_basic(rName1, sourceKey, rName2, key),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3ObjectCopyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "bucket", rName2),
					resource.TestCheckResourceAttr(resourceName, "key", key),
					resource.TestCheckResourceAttr(resourceName, "source", fmt.Sprintf("%s/%s", rName1, sourceKey)),
					resource.TestCheckResourceAttrPair(resourceName, "etag", sourceName, "etag"),
				),
			},
		},
	})
}

func testAccCheckAWSS3ObjectCopyDestroy(s *terraform.State) error {
	s3conn := testAccProvider.Meta().(*AWSClient).s3conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_s3_object_copy" {
			continue
		}

		_, err := s3conn.HeadObject(
			&s3.HeadObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err == nil {
			return fmt.Errorf("AWS S3 Object still exists: %s", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAWSS3ObjectCopyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No S3 Bucket Object ID is set")
		}

		s3conn := testAccProvider.Meta().(*AWSClient).s3conn
		_, err := s3conn.GetObject(
			&s3.GetObjectInput{
				Bucket:  aws.String(rs.Primary.Attributes["bucket"]),
				Key:     aws.String(rs.Primary.Attributes["key"]),
				IfMatch: aws.String(rs.Primary.Attributes["etag"]),
			})
		if err != nil {
			return fmt.Errorf("S3Bucket Object error: %s", err)
		}

		return nil
	}
}

func testAccAWSS3ObjectCopyConfig_basic(rName1, sourceKey, rName2, key string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "source" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "source" {
  bucket  = aws_s3_bucket.source.bucket
  key     = %[2]q
  content = "Ingen ko på isen"
}

resource "aws_s3_bucket" "target" {
  bucket = %[3]q
}

resource "aws_s3_object_copy" "test" {
  bucket = aws_s3_bucket.target.bucket
  key    = %[4]q
  source = "${aws_s3_bucket.source.bucket}/${aws_s3_bucket_object.source.key}"

  grant {
    uri         = "http://acs.amazonaws.com/groups/global/AllUsers"
    type        = "Group"
    permissions = ["READ"]
  }
}
`, rName1, sourceKey, rName2, key)
}
