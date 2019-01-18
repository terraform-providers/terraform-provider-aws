package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMRolePolicy_importBasic(t *testing.T) {
	suffix := randomString(10)
	resourceName := fmt.Sprintf("aws_iam_role_policy.foo_%s", suffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamRolePolicyConfig(suffix),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIAMRolePolicy_basic(t *testing.T) {
	var rolePolicy1, rolePolicy2, rolePolicy3 iam.GetRolePolicyOutput
	role := acctest.RandString(10)
	policy1 := acctest.RandString(10)
	policy2 := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMRolePolicyConfig(role, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						"aws_iam_role.role",
						"aws_iam_role_policy.foo",
						&rolePolicy1,
					),
				),
			},
			{
				Config: testAccIAMRolePolicyConfigUpdate(role, policy1, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						"aws_iam_role.role",
						"aws_iam_role_policy.foo",
						&rolePolicy2,
					),
					testAccCheckIAMRolePolicyExists(
						"aws_iam_role.role",
						"aws_iam_role_policy.bar",
						&rolePolicy3,
					),
					testAccCheckAWSIAMRolePolicyNameMatches(&rolePolicy1, &rolePolicy2),
					testAccCheckAWSIAMRolePolicyNameChanged(&rolePolicy1, &rolePolicy3),
				),
			},
		},
	})
}

func TestAccAWSIAMRolePolicy_disappears(t *testing.T) {
	var out iam.GetRolePolicyOutput
	suffix := randomString(10)
	roleResourceName := fmt.Sprintf("aws_iam_role.role_%s", suffix)
	rolePolicyResourceName := fmt.Sprintf("aws_iam_role_policy.foo_%s", suffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamRolePolicyConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						roleResourceName,
						rolePolicyResourceName,
						&out,
					),
					testAccCheckIAMRolePolicyDisappears(&out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMRolePolicy_namePrefix(t *testing.T) {
	var rolePolicy1, rolePolicy2 iam.GetRolePolicyOutput
	role := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_iam_role_policy.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMRolePolicyConfig_namePrefix(role, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						"aws_iam_role.test",
						"aws_iam_role_policy.test",
						&rolePolicy1,
					),
					resource.TestCheckResourceAttrSet("aws_iam_role_policy.test", "name"),
				),
			},
			{
				Config: testAccIAMRolePolicyConfig_namePrefix(role, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						"aws_iam_role.test",
						"aws_iam_role_policy.test",
						&rolePolicy2,
					),
					testAccCheckAWSIAMRolePolicyNameMatches(&rolePolicy1, &rolePolicy2),
					resource.TestCheckResourceAttrSet("aws_iam_role_policy.test", "name"),
				),
			},
		},
	})
}

func TestAccAWSIAMRolePolicy_generatedName(t *testing.T) {
	var rolePolicy1, rolePolicy2 iam.GetRolePolicyOutput
	role := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_iam_role_policy.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMRolePolicyConfig_generatedName(role, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						"aws_iam_role.test",
						"aws_iam_role_policy.test",
						&rolePolicy1,
					),
					resource.TestCheckResourceAttrSet("aws_iam_role_policy.test", "name"),
				),
			},
			{
				Config: testAccIAMRolePolicyConfig_generatedName(role, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMRolePolicyExists(
						"aws_iam_role.test",
						"aws_iam_role_policy.test",
						&rolePolicy2,
					),
					testAccCheckAWSIAMRolePolicyNameMatches(&rolePolicy1, &rolePolicy2),
					resource.TestCheckResourceAttrSet("aws_iam_role_policy.test", "name"),
				),
			},
		},
	})
}

func TestAccAWSIAMRolePolicy_invalidJSON(t *testing.T) {
	role := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMRolePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccIAMRolePolicyConfig_invalidJSON(role),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
		},
	})
}

func testAccCheckIAMRolePolicyDestroy(s *terraform.State) error {
	iamconn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_role_policy" {
			continue
		}

		role, name, err := resourceAwsIamRolePolicyParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		request := &iam.GetRolePolicyInput{
			PolicyName: aws.String(name),
			RoleName:   aws.String(role),
		}

		getResp, err := iamconn.GetRolePolicy(request)
		if err != nil {
			if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
				// none found, that's good
				return nil
			}
			return fmt.Errorf("Error reading IAM policy %s from role %s: %s", name, role, err)
		}

		if getResp != nil {
			return fmt.Errorf("Found IAM Role, expected none: %s", getResp)
		}
	}

	return nil
}

func testAccCheckIAMRolePolicyDisappears(out *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		params := &iam.DeleteRolePolicyInput{
			PolicyName: out.PolicyName,
			RoleName:   out.RoleName,
		}

		_, err := iamconn.DeleteRolePolicy(params)
		return err
	}
}

func testAccCheckIAMRolePolicyExists(
	iamRoleResource string,
	iamRolePolicyResource string,
	rolePolicy *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamRoleResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamRoleResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[iamRolePolicyResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamRolePolicyResource)
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn
		role, name, err := resourceAwsIamRolePolicyParseId(policy.Primary.ID)
		if err != nil {
			return err
		}

		output, err := iamconn.GetRolePolicy(&iam.GetRolePolicyInput{
			RoleName:   aws.String(role),
			PolicyName: aws.String(name),
		})
		if err != nil {
			return err
		}

		*rolePolicy = *output

		return nil
	}
}

func testAccCheckAWSIAMRolePolicyNameChanged(i, j *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) == aws.StringValue(j.PolicyName) {
			return errors.New("IAM Role Policy name did not change")
		}

		return nil
	}
}

func testAccCheckAWSIAMRolePolicyNameMatches(i, j *iam.GetRolePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) != aws.StringValue(j.PolicyName) {
			return errors.New("IAM Role Policy name did not match")
		}

		return nil
	}
}

func testAccAwsIamRolePolicyConfig(suffix string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role_%[1]s" {
	name = "tf_test_role_test_%[1]s"
	path = "/"
	assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"ec2.amazonaws.com\"},\"Action\":\"sts:AssumeRole\"}]}"
}

resource "aws_iam_role_policy" "foo_%[1]s" {
	name = "tf_test_policy_test_%[1]s"
	role = "${aws_iam_role.role_%[1]s.name}"
	policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, suffix)
}

func testAccIAMRolePolicyConfig(role, policy1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
	name = "tf_test_role_%s"
	path = "/"
	assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf_test_policy_%s"
	role = "${aws_iam_role.role.name}"
	policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, role, policy1)
}

func testAccIAMRolePolicyConfig_namePrefix(role, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
    name = "tf_test_role_%s"
    path = "/"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
    name_prefix = "tf_test_policy_"
    role = "${aws_iam_role.test.name}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "%s",
    "Resource": "*"
  }
}
EOF
}
`, role, policyAction)
}

func testAccIAMRolePolicyConfig_generatedName(role, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
    name = "tf_test_role_%s"
    path = "/"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test" {
    role = "${aws_iam_role.test.name}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "%s",
    "Resource": "*"
  }
}
EOF
}
`, role, policyAction)
}

func testAccIAMRolePolicyConfigUpdate(role, policy1, policy2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
	name = "tf_test_role_%s"
	path = "/"
	assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf_test_policy_%s"
	role = "${aws_iam_role.role.name}"
	policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_role_policy" "bar" {
	name = "tf_test_policy_2_%s"
	role = "${aws_iam_role.role.name}"
	policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, role, policy1, policy2)
}

func testAccIAMRolePolicyConfig_invalidJSON(role string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
	name = "tf_test_role_%s"
	path = "/"
	assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
	name = "tf_test_policy_%s"
	role = "${aws_iam_role.role.name}"
	policy = <<EOF
  {
    "Version": "2012-10-17",
    "Statement": {
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  }
  EOF
}
`, role, role)
}
