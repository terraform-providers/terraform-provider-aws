package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const rfc1123RegexPattern = `^[a-zA-Z]{3}, [0-9]+ [a-zA-Z]+ [0-9]{4} [0-9:]+ [A-Z]+$`

func TestAccDataSourceAWSS3BucketObject_basic(t *testing.T) {
	rInt := acctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_bucket_object.object"
	dataSourceName := "data.aws_s3_bucket_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_basicViaAccessPoint(t *testing.T) {
	var dsObj, rObj s3.GetObjectOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")

	dataSourceName := "data.aws_s3_bucket_object.test"
	resourceName := "aws_s3_bucket_object.test"
	accessPointResourceName := "aws_s3_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfig_basicViaAccessPoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					resource.TestCheckResourceAttrPair(dataSourceName, "bucket", accessPointResourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "key", resourceName, "key"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_readableBody(t *testing.T) {
	rInt := acctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_bucket_object.object"
	dataSourceName := "data.aws_s3_bucket_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfig_readableBody(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, "body", "yes"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_kmsEncrypted(t *testing.T) {
	rInt := acctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_bucket_object.object"
	dataSourceName := "data.aws_s3_bucket_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfig_kmsEncrypted(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "22"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sse_kms_key_id", resourceName, "kms_key_id"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckResourceAttr(dataSourceName, "body", "Keep Calm and Carry On"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_allParams(t *testing.T) {
	rInt := acctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_bucket_object.object"
	dataSourceName := "data.aws_s3_bucket_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfig_allParams(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "25"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "version_id", resourceName, "version_id"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cache_control", resourceName, "cache_control"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_disposition", resourceName, "content_disposition"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_encoding", resourceName, "content_encoding"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_language", resourceName, "content_language"),
					// Encryption is off
					resource.TestCheckResourceAttrPair(dataSourceName, "server_side_encryption", resourceName, "server_side_encryption"),
					resource.TestCheckResourceAttr(dataSourceName, "sse_kms_key_id", ""),
					// Supported, but difficult to reproduce in short testing time
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_class", resourceName, "storage_class"),
					resource.TestCheckResourceAttr(dataSourceName, "expiration", ""),
					// Currently unsupported in aws_s3_bucket_object resource
					resource.TestCheckResourceAttr(dataSourceName, "expires", ""),
					resource.TestCheckResourceAttrPair(dataSourceName, "website_redirect_location", resourceName, "website_redirect"),
					resource.TestCheckResourceAttr(dataSourceName, "metadata.%", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_ObjectLockLegalHoldOff(t *testing.T) {
	rInt := acctest.RandInt()

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_bucket_object.object"
	dataSourceName := "data.aws_s3_bucket_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfig_objectLockLegalHoldOff(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_ObjectLockLegalHoldOn(t *testing.T) {
	rInt := acctest.RandInt()
	retainUntilDate := time.Now().UTC().AddDate(0, 0, 10).Format(time.RFC3339)

	var rObj s3.GetObjectOutput
	var dsObj s3.GetObjectOutput

	resourceName := "aws_s3_bucket_object.object"
	dataSourceName := "data.aws_s3_bucket_object.obj"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfig_objectLockLegalHoldOn(rInt, retainUntilDate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
					resource.TestCheckResourceAttr(dataSourceName, "content_length", "11"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_legal_hold_status", resourceName, "object_lock_legal_hold_status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_mode", resourceName, "object_lock_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "object_lock_retain_until_date", resourceName, "object_lock_retain_until_date"),
					resource.TestCheckNoResourceAttr(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_LeadingSlash(t *testing.T) {
	var rObj s3.GetObjectOutput
	var dsObj1, dsObj2, dsObj3 s3.GetObjectOutput

	resourceName := "aws_s3_bucket_object.object"
	dataSourceName1 := "data.aws_s3_bucket_object.obj1"
	dataSourceName2 := "data.aws_s3_bucket_object.obj2"
	dataSourceName3 := "data.aws_s3_bucket_object.obj3"

	rInt := acctest.RandInt()
	resourceOnlyConf, conf := testAccAWSDataSourceS3ObjectConfig_leadingSlash(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: resourceOnlyConf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName, &rObj),
				),
			},
			{
				Config: conf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName1, &dsObj1),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName1, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),

					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName2, &dsObj2),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName2, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),

					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName3, &dsObj3),
					resource.TestCheckResourceAttr(dataSourceName3, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "content_type", resourceName, "content_type"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "etag", resourceName, "etag"),
					resource.TestMatchResourceAttr(dataSourceName3, "last_modified", regexp.MustCompile(rfc1123RegexPattern)),
					resource.TestCheckResourceAttr(dataSourceName3, "body", "yes"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_MultipleSlashes(t *testing.T) {
	var rObj1, rObj2 s3.GetObjectOutput
	var dsObj1, dsObj2, dsObj3 s3.GetObjectOutput

	resourceName1 := "aws_s3_bucket_object.object1"
	resourceName2 := "aws_s3_bucket_object.object2"
	dataSourceName1 := "data.aws_s3_bucket_object.obj1"
	dataSourceName2 := "data.aws_s3_bucket_object.obj2"
	dataSourceName3 := "data.aws_s3_bucket_object.obj3"

	rInt := acctest.RandInt()
	resourceOnlyConf, conf := testAccAWSDataSourceS3ObjectConfig_multipleSlashes(rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: resourceOnlyConf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists(resourceName1, &rObj1),
					testAccCheckAWSS3BucketObjectExists(resourceName2, &rObj2),
				),
			},
			{
				Config: conf,
				Check: resource.ComposeTestCheckFunc(

					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName1, &dsObj1),
					resource.TestCheckResourceAttr(dataSourceName1, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "content_type", resourceName1, "content_type"),
					resource.TestCheckResourceAttr(dataSourceName1, "body", "yes"),

					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName2, &dsObj2),
					resource.TestCheckResourceAttr(dataSourceName2, "content_length", "3"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "content_type", resourceName1, "content_type"),
					resource.TestCheckResourceAttr(dataSourceName2, "body", "yes"),

					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName3, &dsObj3),
					resource.TestCheckResourceAttr(dataSourceName3, "content_length", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName3, "content_type", resourceName2, "content_type"),
					resource.TestCheckResourceAttr(dataSourceName3, "body", "no"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObject_SingleSlashAsKey(t *testing.T) {
	var dsObj s3.GetObjectOutput
	dataSourceName := "data.aws_s3_bucket_object.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectConfigSingleSlashAsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectDataSourceExists(dataSourceName, &dsObj),
				),
			},
		},
	})
}

func testAccCheckAwsS3ObjectDataSourceExists(n string, obj *s3.GetObjectOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find S3 object data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("S3 object data source ID not set")
		}

		s3conn := testAccProvider.Meta().(*AWSClient).s3conn
		out, err := s3conn.GetObject(
			&s3.GetObjectInput{
				Bucket: aws.String(rs.Primary.Attributes["bucket"]),
				Key:    aws.String(rs.Primary.Attributes["key"]),
			})
		if err != nil {
			return fmt.Errorf("Failed getting S3 Object from %s: %s",
				rs.Primary.Attributes["bucket"]+"/"+rs.Primary.Attributes["key"], err)
		}

		*obj = *out

		return nil
	}
}

func testAccAWSDataSourceS3ObjectConfig_basic(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_bucket_object" "object" {
  bucket  = aws_s3_bucket.object_bucket.bucket
  key     = "tf-testing-obj-%[1]d"
  content = "Hello World"
}

data "aws_s3_bucket_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_bucket_object.object.key
}
`, randInt)
}

func testAccAWSDataSourceS3ObjectConfig_basicViaAccessPoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.test.bucket
  name   = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = %[1]q
  content = "Hello World"
}

data "aws_s3_bucket_object" "test" {
  bucket = aws_s3_access_point.test.arn
  key    = aws_s3_bucket_object.test.key
}
`, rName)
}

func testAccAWSDataSourceS3ObjectConfig_readableBody(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_bucket_object" "object" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "tf-testing-obj-%[1]d-readable"
  content      = "yes"
  content_type = "text/plain"
}

data "aws_s3_bucket_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_bucket_object.object.key
}
`, randInt)
}

func testAccAWSDataSourceS3ObjectConfig_kmsEncrypted(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_kms_key" "example" {
  description             = "TF Acceptance Test KMS key"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket_object" "object" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "tf-testing-obj-%[1]d-encrypted"
  content      = "Keep Calm and Carry On"
  content_type = "text/plain"
  kms_key_id   = aws_kms_key.example.arn
}

data "aws_s3_bucket_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_bucket_object.object.key
}
`, randInt)
}

func testAccAWSDataSourceS3ObjectConfig_allParams(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "tf-testing-obj-%[1]d-all-params"

  content             = <<CONTENT
{
  "msg": "Hi there!"
}
CONTENT
  content_type        = "application/unknown"
  cache_control       = "no-cache"
  content_disposition = "attachment"
  content_encoding    = "identity"
  content_language    = "en-GB"

  tags = {
    Key1 = "Value 1"
  }
}

data "aws_s3_bucket_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_bucket_object.object.key
}
`, randInt)
}

func testAccAWSDataSourceS3ObjectConfig_objectLockLegalHoldOff(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket                        = aws_s3_bucket.object_bucket.bucket
  key                           = "tf-testing-obj-%[1]d"
  content                       = "Hello World"
  object_lock_legal_hold_status = "OFF"
}

data "aws_s3_bucket_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_bucket_object.object.key
}
`, randInt)
}

func testAccAWSDataSourceS3ObjectConfig_objectLockLegalHoldOn(randInt int, retainUntilDate string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"

  versioning {
    enabled = true
  }

  object_lock_configuration {
    object_lock_enabled = "Enabled"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket                        = aws_s3_bucket.object_bucket.bucket
  key                           = "tf-testing-obj-%[1]d"
  content                       = "Hello World"
  force_destroy                 = true
  object_lock_legal_hold_status = "ON"
  object_lock_mode              = "GOVERNANCE"
  object_lock_retain_until_date = "%[2]s"
}

data "aws_s3_bucket_object" "obj" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = aws_s3_bucket_object.object.key
}
`, randInt, retainUntilDate)
}

func testAccAWSDataSourceS3ObjectConfig_leadingSlash(randInt int) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_bucket_object" "object" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "//tf-testing-obj-%[1]d-readable"
  content      = "yes"
  content_type = "text/plain"
}
`, randInt)

	both := fmt.Sprintf(`
%[1]s

data "aws_s3_bucket_object" "obj1" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "tf-testing-obj-%[2]d-readable"
}

data "aws_s3_bucket_object" "obj2" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "/tf-testing-obj-%[2]d-readable"
}

data "aws_s3_bucket_object" "obj3" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "//tf-testing-obj-%[2]d-readable"
}
`, resources, randInt)

	return resources, both
}

func testAccAWSDataSourceS3ObjectConfig_multipleSlashes(randInt int) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
  bucket = "tf-object-test-bucket-%[1]d"
}

resource "aws_s3_bucket_object" "object1" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "first//second///third//"
  content      = "yes"
  content_type = "text/plain"
}

# Without a trailing slash.
resource "aws_s3_bucket_object" "object2" {
  bucket       = aws_s3_bucket.object_bucket.bucket
  key          = "/first////second/third"
  content      = "no"
  content_type = "text/plain"
}
`, randInt)

	both := fmt.Sprintf(`
%s

data "aws_s3_bucket_object" "obj1" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "first/second/third/"
}

data "aws_s3_bucket_object" "obj2" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "first//second///third//"
}

data "aws_s3_bucket_object" "obj3" {
  bucket = aws_s3_bucket.object_bucket.bucket
  key    = "first/second/third"
}
`, resources)

	return resources, both
}

func testAccAWSDataSourceS3ObjectConfigSingleSlashAsKey(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "/"
}
`, rName)
}
