package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcrRepositoryPolicy_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecr.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "aws_ecr_repository.test", "name"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrRepositoryPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "repository", "aws_ecr_repository.test", "name"),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("ecr:DescribeImages")),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
				),
			},
		},
	})
}

func TestAccAWSEcrRepositoryPolicy_iam(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecr.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyWithIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile(rName)),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("iam")),
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

func TestAccAWSEcrRepositoryPolicy_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecr.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcrRepositoryPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEcrRepositoryPolicy_disappears_repository(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ecr.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryPolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcrRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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
			RepositoryName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") ||
				isAWSErr(err, ecr.ErrCodeRepositoryPolicyNotFoundException, "") {
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

func testAccAWSEcrRepositoryPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "%[1]s",
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
`, rName)
}

func testAccAWSEcrRepositoryPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "%[1]s",
            "Effect": "Allow",
            "Principal": "*",
            "Action": [
                "ecr:ListImages",
                "ecr:DescribeImages"
            ]
        }
    ]
}
EOF
}
`, rName)
}

// testAccAWSEcrRepositoryPolicyWithIAMRoleConfig creates a new IAM Role and tries
// to use it's ARN in an ECR Repository Policy. IAM changes need some time to
// be propagated to other services - like ECR. So the following code should
// exercise our retry logic, since we try to use the new resource instantly.
func testAccAWSEcrRepositoryPolicyWithIAMRoleConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_ecr_repository_policy" "test" {
  repository = aws_ecr_repository.test.name

  policy = <<EOF
{
    "Version": "2008-10-17",
    "Statement": [
        {
            "Sid": "%[1]s",
            "Effect": "Allow",
            "Principal": {
              "AWS": "${aws_iam_role.test.arn}"
            },
            "Action": [
                "ecr:ListImages"
            ]
        }
    ]
}
EOF
}
`, rName)
}
