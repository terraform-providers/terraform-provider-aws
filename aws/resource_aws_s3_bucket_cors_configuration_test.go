package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSS3BucketCorsConfiguration_CreationWithMissingField(t *testing.T) {
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSBucketCorsConfigurationCreationWithMissingRequiredField(bucketName),
				ExpectError: regexp.MustCompile(`Missing required argument: The argument "bucket" is required, but no definition was found`),
			},
		},
	})
}

func TestAccAWSS3BucketCorsConfiguration_UpdateWithMissingCorsField(t *testing.T) {
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSBucketCorsConfigurationUpdateWithMissingRequiredCorsRuleField(bucketName),
				ExpectError: regexp.MustCompile(`Missing required argument: The argument "allowed_methods" is required, but no definition was found`),
			},
		},
	})
}

func TestAccAWSS3BucketCorsConfiguration_CorsPolicyCreation(t *testing.T) {
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)
	bucket := "aws_s3_bucket.cors_test_bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBucketCorsConfigurationWithCorsPolicyCreation(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(bucket),
					testAccCheckAWSS3BucketCors(
						bucket,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("PUT"), aws.String("POST")},
								AllowedOrigins: []*string{aws.String("https://www.cors-configuration-test-create.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSS3BucketCorsConfiguration_CorsPolicyUpdate(t *testing.T) {
	rInt := acctest.RandInt()
	bucketName := fmt.Sprintf("tf-test-bucket-%d", rInt)
	bucket := "aws_s3_bucket.cors_test_bucket"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSS3BucketDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSBucketCorsConfigurationWithCorsPolicyUpdate(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketExists(bucket),
					testAccCheckAWSS3BucketCors(
						bucket,
						[]*s3.CORSRule{
							{
								AllowedHeaders: []*string{aws.String("*")},
								AllowedMethods: []*string{aws.String("GET")},
								AllowedOrigins: []*string{aws.String("https://www.cors-configuration-test-update.com")},
								ExposeHeaders:  []*string{aws.String("x-amz-server-side-encryption"), aws.String("ETag")},
								MaxAgeSeconds:  aws.Int64(3000),
							},
						},
					),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSBucketCorsConfigurationWithCorsPolicyCreation(bucketName string) string {
	return fmt.Sprintf(`
		resource "aws_s3_bucket" "cors_test_bucket" {
			bucket = "%s"
			acl    = "public-read"
		}

		resource "aws_s3_bucket_cors_configuration" "test" {
			bucket = "${aws_s3_bucket.cors_test_bucket.id}"

			cors_rule {
				allowed_headers = ["*"]
				allowed_methods = ["PUT", "POST"]
				allowed_origins = ["https://www.cors-configuration-test-create.com"]
				expose_headers  = ["x-amz-server-side-encryption", "ETag"]
				max_age_seconds = 3000
			}

		}
		`, bucketName)
}

func testAccAWSBucketCorsConfigurationCreationWithMissingRequiredField(bucketName string) string {
	return fmt.Sprintf(`
		resource "aws_s3_bucket" "cors_test_bucket" {
			bucket = "%s"
			acl    = "public-read"
		}

		resource "aws_s3_bucket_cors_configuration" "test" {
			
			cors_rule {
				allowed_headers = ["*"]
				allowed_methods = ["PUT", "POST"]
				allowed_origins = ["https://www.cors-configuration-test-create.com"]
				expose_headers  = ["x-amz-server-side-encryption", "ETag"]
				max_age_seconds = 3000
			}

		}
		`, bucketName)
}

func testAccAWSBucketCorsConfigurationWithCorsPolicyUpdate(bucketName string) string {
	return fmt.Sprintf(`
		resource "aws_s3_bucket" "cors_test_bucket" {
			bucket = "%s"
			acl    = "public-read"

			cors_rule {
				allowed_headers = ["*"]
				allowed_methods = ["PUT", "POST"]
				allowed_origins = ["https://www.cors-configuration.com"]
				expose_headers  = ["x-amz-server-side-encryption", "ETag"]
				max_age_seconds = 3000
			}
		}

		resource "aws_s3_bucket_cors_configuration" "test" {
			bucket = "${aws_s3_bucket.cors_test_bucket.id}"

			cors_rule {
				allowed_headers = ["*"]
				allowed_methods = ["GET"]
				allowed_origins = ["https://www.cors-configuration-test-update.com"]
				expose_headers  = ["x-amz-server-side-encryption", "ETag"]
				max_age_seconds = 3000
			}

		}
		`, bucketName)
}

func testAccAWSBucketCorsConfigurationUpdateWithMissingRequiredCorsRuleField(bucketName string) string {
	return fmt.Sprintf(`
		resource "aws_s3_bucket" "cors_test_bucket" {
			bucket = "%s"
			acl    = "public-read"

			cors_rule {
				allowed_headers = ["*"]
				allowed_methods = ["PUT", "POST"]
				allowed_origins = ["https://www.cors-configuration.com"]
				expose_headers  = ["x-amz-server-side-encryption", "ETag"]
				max_age_seconds = 3000
			}
		}

		resource "aws_s3_bucket_cors_configuration" "test" {
			bucket = "${aws_s3_bucket.cors_test_bucket.id}"

			# allow_methods missing
			cors_rule {
				allowed_headers = ["*"]
				allowed_origins = ["https://www.cors-configuration-test-update.com"]
				expose_headers  = ["x-amz-server-side-encryption", "ETag"]
				max_age_seconds = 3000
			}

		}
		`, bucketName)
}
