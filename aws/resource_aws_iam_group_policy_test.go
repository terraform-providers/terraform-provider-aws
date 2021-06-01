package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSIAMGroupPolicy_basic(t *testing.T) {
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.group",
						"aws_iam_group_policy.foo",
						&groupPolicy1,
					),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMGroupPolicyConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.group",
						"aws_iam_group_policy.bar",
						&groupPolicy2,
					),
					testAccCheckAWSIAMGroupPolicyNameChanged(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIAMGroupPolicy_disappears(t *testing.T) {
	var out iam.GetGroupPolicyOutput
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.group",
						"aws_iam_group_policy.foo",
						&out,
					),
					testAccCheckIAMGroupPolicyDisappears(&out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMGroupPolicy_namePrefix(t *testing.T) {
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_iam_group_policy.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig_namePrefix(rInt, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy1,
					),
				),
			},
			{
				Config: testAccIAMGroupPolicyConfig_namePrefix(rInt, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy2,
					),
					testAccCheckAWSIAMGroupPolicyNameMatches(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:            "aws_iam_group_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSIAMGroupPolicy_generatedName(t *testing.T) {
	var groupPolicy1, groupPolicy2 iam.GetGroupPolicyOutput
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_iam_group_policy.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckIAMGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMGroupPolicyConfig_generatedName(rInt, "*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy1,
					),
				),
			},
			{
				Config: testAccIAMGroupPolicyConfig_generatedName(rInt, "ec2:*"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMGroupPolicyExists(
						"aws_iam_group.test",
						"aws_iam_group_policy.test",
						&groupPolicy2,
					),
					testAccCheckAWSIAMGroupPolicyNameMatches(&groupPolicy1, &groupPolicy2),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIAMGroupPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_group_policy" {
			continue
		}

		group, name, err := resourceAwsIamGroupPolicyParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		request := &iam.GetGroupPolicyInput{
			PolicyName: aws.String(name),
			GroupName:  aws.String(group),
		}

		getResp, err := conn.GetGroupPolicy(request)
		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				// none found, that's good
				continue
			}
			return fmt.Errorf("Error reading IAM policy %s from group %s: %s", name, group, err)
		}

		if getResp != nil {
			return fmt.Errorf("Found IAM group policy, expected none: %s", getResp)
		}
	}

	return nil
}

func testAccCheckIAMGroupPolicyDisappears(out *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		params := &iam.DeleteGroupPolicyInput{
			PolicyName: out.PolicyName,
			GroupName:  out.GroupName,
		}

		_, err := iamconn.DeleteGroupPolicy(params)
		return err
	}
}

func testAccCheckIAMGroupPolicyExists(
	iamGroupResource string,
	iamGroupPolicyResource string,
	groupPolicy *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamGroupResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamGroupResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[iamGroupPolicyResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamGroupPolicyResource)
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn
		group, name, err := resourceAwsIamGroupPolicyParseId(policy.Primary.ID)
		if err != nil {
			return err
		}

		output, err := iamconn.GetGroupPolicy(&iam.GetGroupPolicyInput{
			GroupName:  aws.String(group),
			PolicyName: aws.String(name),
		})

		if err != nil {
			return err
		}

		*groupPolicy = *output

		return nil
	}
}

func testAccCheckAWSIAMGroupPolicyNameChanged(i, j *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) == aws.StringValue(j.PolicyName) {
			return errors.New("IAM Group Policy name did not change")
		}

		return nil
	}
}

func testAccCheckAWSIAMGroupPolicyNameMatches(i, j *iam.GetGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.PolicyName) != aws.StringValue(j.PolicyName) {
			return errors.New("IAM Group Policy name did not match")
		}

		return nil
	}
}

func testAccIAMGroupPolicyConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "test_group_%d"
  path = "/"
}

resource "aws_iam_group_policy" "foo" {
  name  = "foo_policy_%d"
  group = aws_iam_group.group.name

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
`, rInt, rInt)
}

func testAccIAMGroupPolicyConfig_namePrefix(rInt int, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = "test_group_%d"
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  name_prefix = "test-%d"
  group       = aws_iam_group.test.name

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
`, rInt, rInt, policyAction)
}

func testAccIAMGroupPolicyConfig_generatedName(rInt int, policyAction string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = "test_group_%d"
  path = "/"
}

resource "aws_iam_group_policy" "test" {
  group = aws_iam_group.test.name

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
`, rInt, policyAction)
}

func testAccIAMGroupPolicyConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "test_group_%d"
  path = "/"
}

resource "aws_iam_group_policy" "foo" {
  name  = "foo_policy_%d"
  group = aws_iam_group.group.name

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

resource "aws_iam_group_policy" "bar" {
  name  = "bar_policy_%d"
  group = aws_iam_group.group.name

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
`, rInt, rInt, rInt)
}
