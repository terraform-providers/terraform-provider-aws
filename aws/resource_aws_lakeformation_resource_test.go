package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLakeFormationResource_basic(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	roleName := acctest.RandomWithPrefix("tf-acc-test")
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"
	roleAddr := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLakeFormationResource_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lakeformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationResourceExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLakeFormationResource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLakeFormationResource_serviceLinkedRole(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t)
			testAccPreCheckIamServiceLinkedRole(t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_serviceLinkedRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
					testAccCheckResourceAttrGlobalARN(resourceAddr, "role_arn", "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
		},
	})
}

func TestAccAWSLakeFormationResource_updateRoleToRole(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	roleName1 := acctest.RandomWithPrefix("tf-acc-test")
	roleName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"
	roleAddr := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(bucketName, roleName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(bucketName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLakeFormationResource_updateSLRToRole(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	roleName := acctest.RandomWithPrefix("tf-acc-test")
	resourceAddr := "aws_lakeformation_resource.test"
	bucketAddr := "aws_s3_bucket.test"
	roleAddr := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t)
			testAccPreCheckIamServiceLinkedRole(t, "/aws-service-role/lakeformation.amazonaws.com")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLakeFormationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceConfig_serviceLinkedRole(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
					testAccCheckResourceAttrGlobalARN(resourceAddr, "role_arn", "iam", "role/aws-service-role/lakeformation.amazonaws.com/AWSServiceRoleForLakeFormationDataAccess"),
				),
			},
			{
				Config: testAccAWSLakeFormationResourceConfig_basic(bucketName, roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLakeFormationResourceExists(resourceAddr),
					resource.TestCheckResourceAttrPair(resourceAddr, "role_arn", roleAddr, "arn"),
					resource.TestCheckResourceAttrPair(resourceAddr, "arn", bucketAddr, "arn"),
				),
			},
		},
	})
}

// AWS does not support changing from an IAM role to an SLR. No error is thrown
// but the registration is not changed (the IAM role continues in the registration).
//
// func TestAccAWSLakeFormationResource_updateRoleToSLR(t *testing.T) {

func testAccCheckAWSLakeFormationResourceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lakeformation_resource" {
			continue
		}

		resourceArn := rs.Primary.Attributes["arn"]

		input := &lakeformation.DescribeResourceInput{
			ResourceArn: aws.String(resourceArn),
		}

		_, err := conn.DescribeResource(input)
		if err == nil {
			return fmt.Errorf("resource still registered: %s", resourceArn)
		}
		if !isLakeFormationResourceNotFoundErr(err) {
			return err
		}
	}

	return nil
}

func testAccCheckAWSLakeFormationResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).lakeformationconn

		input := &lakeformation.DescribeResourceInput{
			ResourceArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeResource(input)

		if err != nil {
			return fmt.Errorf("error getting Lake Formation resource (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func isLakeFormationResourceNotFoundErr(err error) bool {
	return isAWSErr(
		err,
		"EntityNotFoundException",
		"Entity not found")
}

func testAccAWSLakeFormationResourceConfig_basic(bucket, role string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[2]q
  path = "/test/"

  assume_role_policy = <<EOF
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
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[2]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:GetObject",
        "s3:GetObjectACL",
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
      ]
    }
  ]
}
EOF
}

resource "aws_lakeformation_resource" "test" {
  arn      = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}
`, bucket, role)
}

func testAccAWSLakeFormationResourceConfig_serviceLinkedRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lakeformation_resource" "test" {
  arn = aws_s3_bucket.test.arn
}
`, rName)
}
