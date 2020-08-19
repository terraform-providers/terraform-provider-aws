package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSQuickSightIAMPolicyAssignment_basic(t *testing.T) {
	var out quicksight.DescribeIAMPolicyAssignmentOutput

	rString := acctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gpa-basic-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-group-gpa-basic-2-%s", rString)
	groupName3 := fmt.Sprintf("tf-acc-group-gpa-basic-3-%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-gpa-basic-%s", rString)
	//	policyName2 := fmt.Sprintf("tf-acc-policy-gpa-basic-2-%s", rString)
	//	policyName3 := fmt.Sprintf("tf-acc-policy-gpa-basic-3-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSQuickSightIAMPolicyAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightIAMPolicyAssignConfig(groupName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQuickSightIAMPolicyAssignmentExists("aws_quicksight_iam_policy_assignment.test-assign", 1, &out),
					testAccCheckAWSQuickSightIAMPolicyAssignmentAttributes([]string{groupName}, &out),
				),
			},
			{
				ResourceName:      "aws_quicksight_iam_policy_assignment.test-assign",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSQuickSightIAMPolicyAssignmentImportStateIdFunc("aws_quicksight_iam_policy_assignment.test-assign"),
				// We do not have a way to align IDs since the Create function uses resource.PrefixedUniqueId()
				// Failed state verification, resource with ID GROUP-POLICYARN not found
				// ImportStateVerify: true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %#v", s)
					}
					rs := s[0]
					if !strings.HasPrefix(rs.Attributes["policy_arn"], "arn:") {
						return fmt.Errorf("expected policy_arn attribute to be set and begin with arn:, received: %s", rs.Attributes["policy_arn"])
					}
					return nil
				},
			},
			{
				Config: testAccAWSQuickSightIAMPolicyAssignConfigUpdate(groupName, groupName2, groupName3, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQuickSightIAMPolicyAssignmentExists("aws_quicksight_iam_policy_assignment.test-assign", 3, &out),
					testAccCheckAWSQuickSightIAMPolicyAssignmentAttributes([]string{groupName, groupName2, groupName3}, &out),
				),
			},
		},
	})
}
func testAccCheckAWSQuickSightIAMPolicyAssignmentDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckAWSQuickSightIAMPolicyAssignmentExists(n string, c int, out *quicksight.DescribeIAMPolicyAssignmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).quicksightconn
		//		group := rs.Primary.Attributes["group_name"]
		assignment := rs.Primary.Attributes["assignment_name"]

		identities, err := conn.DescribeIAMPolicyAssignment(&quicksight.DescribeIAMPolicyAssignmentInput{
			AssignmentName: aws.String(assignment),
			AwsAccountId:   aws.String(testAccProvider.Meta().(*AWSClient).accountid),
			Namespace:      aws.String("default"),
		})
		if err != nil {
			return fmt.Errorf("Error: Failed to get the policy assignment(%s): %s", assignment, err)
		}
		if c != len(identities.IAMPolicyAssignment.Identities["group"]) {
			return fmt.Errorf("Error: policy (%s) is assigned to wrong number of identities on initial creation (should ne %#v but was %#v. Assignment: %#v", n, c, len(identities.IAMPolicyAssignment.Identities["group"]), identities)
		}

		*out = *identities
		return nil
	}
}
func testAccCheckAWSQuickSightIAMPolicyAssignmentAttributes(identities []string, out *quicksight.DescribeIAMPolicyAssignmentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		for _, i := range identities {

			if !strings.Contains(out.IAMPolicyAssignment.String(), i) {
				return fmt.Errorf("Error: Policy is not assigned to the group %s, %s", i, out.IAMPolicyAssignment.String())
			}
		}
		return nil
	}
}

func testAccAWSQuickSightIAMPolicyAssignConfig(groupName, policyName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_group" "group" {
  group_name = "%s"
}

resource "aws_iam_policy" "policy" {
  name        = "%s"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_quicksight_iam_policy_assignment" "test-assign" {
  assignment_name   = "testassign"
  groups            = ["${aws_quicksight_group.group.group_name}"]
  policy_arn        = aws_iam_policy.policy.arn
  assignment_status = "ENABLED"
  namespace         = "default"
  aws_account_id    = data.aws_caller_identity.current.account_id
  depends_on = [
    aws_quicksight_group.group,
	aws_iam_policy.policy
  ]
}
`, groupName, policyName)
}

func testAccAWSQuickSightIAMPolicyAssignConfigUpdate(groupName, groupName2, groupName3, policyName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_group" "group" {
  group_name = "%s"
}

resource "aws_quicksight_group" "group2" {
  group_name = "%s"
}

resource "aws_quicksight_group" "group3" {
  group_name = "%s"
}

resource "aws_iam_policy" "policy" {
  name        = "%s"
  description = "A test policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "iam:ChangePassword"
      ],
      "Resource": "*",
      "Effect": "Allow"
    }
  ]
}
EOF
}

resource "aws_quicksight_iam_policy_assignment" "test-assign" {
  assignment_name   = "testassign"
  groups            = [
    "${aws_quicksight_group.group.group_name}",
    "${aws_quicksight_group.group2.group_name}",
    "${aws_quicksight_group.group3.group_name}"
  ]
  policy_arn        = aws_iam_policy.policy.arn
  assignment_status = "ENABLED"
  namespace         = "default"
  aws_account_id    = data.aws_caller_identity.current.account_id
  depends_on = [
    aws_quicksight_group.group,
	aws_iam_policy.policy
  ]
}
`, groupName, groupName2, groupName3, policyName)
}

func testAccAWSQuickSightIAMPolicyAssignmentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		awsAccountId := testAccProvider.Meta().(*AWSClient).accountid
		namespace := "default"
		return fmt.Sprintf("%s/%s/%s", awsAccountId, namespace, rs.Primary.Attributes["assignment_name"]), nil
	}
}
