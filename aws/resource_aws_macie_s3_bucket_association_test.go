package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMacieS3BucketAssociation_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMacieS3BucketAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMacieS3BucketAssociationConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMacieS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "NONE"),
				),
			},
			{
				Config: testAccAWSMacieS3BucketAssociationConfig_basicOneTime(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMacieS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "FULL"),
				),
			},
		},
	})
}

func TestAccAWSMacieS3BucketAssociation_accountIdAndPrefix(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMacieS3BucketAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMacieS3BucketAssociationConfig_accountIdAndPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMacieS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "NONE"),
				),
			},
			{
				Config: testAccAWSMacieS3BucketAssociationConfig_accountIdAndPrefixOneTime(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMacieS3BucketAssociationExists("aws_macie_s3_bucket_association.test"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.continuous", "FULL"),
					resource.TestCheckResourceAttr("aws_macie_s3_bucket_association.test", "classification_type.0.one_time", "FULL"),
				),
			},
		},
	})
}

func testAccCheckAWSMacieS3BucketAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).macieconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_macie_s3_bucket_association" {
			continue
		}

		req := &macie.ListS3ResourcesInput{}
		acctId := rs.Primary.Attributes["member_account_id"]
		if acctId != "" {
			req.MemberAccountId = aws.String(acctId)
		}

		dissociated := true
		err := conn.ListS3ResourcesPages(req, func(page *macie.ListS3ResourcesOutput, lastPage bool) bool {
			for _, v := range page.S3Resources {
				if aws.StringValue(v.BucketName) == rs.Primary.Attributes["bucket_name"] && aws.StringValue(v.Prefix) == rs.Primary.Attributes["prefix"] {
					dissociated = false
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !dissociated {
			return fmt.Errorf("S3 resource %s/%s is not dissociated from Macie", rs.Primary.Attributes["bucket_name"], rs.Primary.Attributes["prefix"])
		}
	}
	return nil
}

func testAccCheckAWSMacieS3BucketAssociationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).macieconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		req := &macie.ListS3ResourcesInput{}
		acctId := rs.Primary.Attributes["member_account_id"]
		if acctId != "" {
			req.MemberAccountId = aws.String(acctId)
		}

		exists := false
		err := conn.ListS3ResourcesPages(req, func(page *macie.ListS3ResourcesOutput, lastPage bool) bool {
			for _, v := range page.S3Resources {
				if aws.StringValue(v.BucketName) == rs.Primary.Attributes["bucket_name"] && aws.StringValue(v.Prefix) == rs.Primary.Attributes["prefix"] {
					exists = true
					return false
				}
			}

			return true
		})
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("S3 resource %s/%s is not associated with Macie", rs.Primary.Attributes["bucket_name"], rs.Primary.Attributes["prefix"])
		}

		return nil
	}
}

func testAccAWSMacieS3BucketAssociationConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-macie-test-bucket-%d"
}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name = "${aws_s3_bucket.test.id}"
}
`, randInt)
}

func testAccAWSMacieS3BucketAssociationConfig_basicOneTime(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-macie-test-bucket-%d"
}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name = "${aws_s3_bucket.test.id}"

  classification_type {
    one_time = "FULL"
  }
}
`, randInt)
}

func testAccAWSMacieS3BucketAssociationConfig_accountIdAndPrefix(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-macie-test-bucket-%d"
}

data "aws_caller_identity" "current" {}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name       = "${aws_s3_bucket.test.id}"
  member_account_id = "${data.aws_caller_identity.current.account_id}"
  prefix            = "data"
}
`, randInt)
}

func testAccAWSMacieS3BucketAssociationConfig_accountIdAndPrefixOneTime(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = "tf-macie-test-bucket-%d"
}

data "aws_caller_identity" "current" {}

resource "aws_macie_s3_bucket_association" "test" {
  bucket_name       = "${aws_s3_bucket.test.id}"
  member_account_id = "${data.aws_caller_identity.current.account_id}"
  prefix            = "data"

  classification_type {
    one_time = "FULL"
  }
}
`, randInt)
}
