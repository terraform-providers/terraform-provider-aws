package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudFrontOriginAccessIdentity_basic(t *testing.T) {
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontOriginAccessIdentityConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontOriginAccessIdentityExistence(resourceName, &origin),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestMatchResourceAttr(resourceName, "caller_reference", regexp.MustCompile(fmt.Sprintf("^%s", resource.UniqueIdPrefix))),
					resource.TestMatchResourceAttr(resourceName, "s3_canonical_user_id", regexp.MustCompile("^[a-z0-9]+")),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_access_identity_path", regexp.MustCompile("^origin-access-identity/cloudfront/[A-Z0-9]+")),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "iam_arn", regexp.MustCompile(fmt.Sprintf("^arn:%s:iam::cloudfront:user/CloudFront Origin Access Identity [A-Z0-9]+", testAccGetPartition()))),
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

func TestAccAWSCloudFrontOriginAccessIdentity_noComment(t *testing.T) {
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontOriginAccessIdentityNoCommentConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontOriginAccessIdentityExistence(resourceName, &origin),
					resource.TestMatchResourceAttr(resourceName, "caller_reference", regexp.MustCompile(fmt.Sprintf("^%s", resource.UniqueIdPrefix))),
					resource.TestMatchResourceAttr(resourceName, "s3_canonical_user_id", regexp.MustCompile("^[a-z0-9]+")),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_access_identity_path", regexp.MustCompile("^origin-access-identity/cloudfront/[A-Z0-9]+")),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "iam_arn", regexp.MustCompile(fmt.Sprintf("^arn:%s:iam::cloudfront:user/CloudFront Origin Access Identity [A-Z0-9]+", testAccGetPartition()))),
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

func TestAccAWSCloudFrontOriginAccessIdentity_disappears(t *testing.T) {
	var origin cloudfront.GetCloudFrontOriginAccessIdentityOutput
	resourceName := "aws_cloudfront_origin_access_identity.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontOriginAccessIdentityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontOriginAccessIdentityConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontOriginAccessIdentityExistence(resourceName, &origin),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFrontOriginAccessIdentity(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudFrontOriginAccessIdentityDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_origin_access_identity" {
			continue
		}

		params := &cloudfront.GetCloudFrontOriginAccessIdentityInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCloudFrontOriginAccessIdentity(params)
		if err == nil {
			return fmt.Errorf("CloudFront origin access identity was not deleted")
		}
	}

	return nil
}

func testAccCheckCloudFrontOriginAccessIdentityExistence(r string, origin *cloudfront.GetCloudFrontOriginAccessIdentityOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Id is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		params := &cloudfront.GetCloudFrontOriginAccessIdentityInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetCloudFrontOriginAccessIdentity(params)
		if err != nil {
			return fmt.Errorf("Error retrieving CloudFront distribution: %s", err)
		}

		*origin = *resp

		return nil
	}
}

const testAccAWSCloudFrontOriginAccessIdentityConfig = `
resource "aws_cloudfront_origin_access_identity" "test" {
  comment = "some comment"
}
`

const testAccAWSCloudFrontOriginAccessIdentityNoCommentConfig = `
resource "aws_cloudfront_origin_access_identity" "test" {
}
`
