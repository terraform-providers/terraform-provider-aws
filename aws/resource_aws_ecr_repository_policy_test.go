package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEcrRepositoryPolicy_basic(t *testing.T) {
	randString := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicy(randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists("aws_ecr_repository_policy.default"),
				),
			},
		},
	})
}

func TestAccAWSEcrRepositoryPolicy_iam(t *testing.T) {
	randString := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyWithIAMRole(randString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists("aws_ecr_repository_policy.default"),
				),
			},
		},
	})
}

func testAccCheckAWSEcrRepositoryPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_repository_policy" {
			continue
		}

		_, err := conn.GetRepositoryPolicy(&ecr.GetRepositoryPolicyInput{
			RegistryId:     aws.String(rs.Primary.Attributes["registry_id"]),
			RepositoryName: aws.String(rs.Primary.Attributes["repository"]),
		})
		if err != nil {
			if ecrerr, ok := err.(awserr.Error); ok && ecrerr.Code() == "RepositoryNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSEcrRepositoryPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSEcrRepositoryPolicy(randString string) string {
	return fmt.Sprintf(`
# ECR initially only available in us-east-1
# https://aws.amazon.com/blogs/aws/ec2-container-registry-now-generally-available/
provider "aws" {
	region = "us-east-1"
}
resource "aws_ecr_repository" "foo" {
	name = "tf-acc-test-ecr-%s"
}

resource "aws_ecr_repository_policy" "default" {
	repository = "${aws_ecr_repository.foo.name}"
	policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "testpolicy",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "ecr:ListImages"
            ]
        }
    ]
}
EOF
}
`, randString)
}

// testAccAWSEcrRepositoryPolicyWithIAMRole creates a new IAM Role and tries
// to use it's ARN in an ECR Repository Policy. IAM changes need some time to
// be propagated to other services - like ECR. So the following code should
// exercise our retry logic, since we try to use the new resource instantly.
func testAccAWSEcrRepositoryPolicyWithIAMRole(randString string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}

resource "aws_ecr_repository" "foo" {
	name = "tf-acc-test-ecr-%s"
}

resource "aws_iam_role" "foo" {
  name = "tf-acc-test-ecr-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      }
    }
  ]
}
EOF
}

resource "aws_ecr_repository_policy" "default" {
	repository = "${aws_ecr_repository.foo.name}"
	policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "testpolicy",
            "Effect": "Allow",
            "Principal": {
              "AWS": "${aws_iam_role.foo.arn}"
            },
            "Action": [
                "ecr:ListImages"
            ]
        }
    ]
}
EOF
}
`, randString, randString)
}
