package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMGroupPolicyAttachment_basic(t *testing.T) {
	var out iam.ListAttachedGroupPoliciesOutput

	rString := acctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gpa-basic-%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-gpa-basic-%s", rString)
	policyName2 := fmt.Sprintf("tf-acc-policy-gpa-basic-2-%s", rString)
	policyName3 := fmt.Sprintf("tf-acc-policy-gpa-basic-3-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGroupPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGroupPolicyAttachConfig(groupName, policyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupPolicyAttachmentExists("aws_iam_group_policy_attachment.test-attach", 1, &out),
					testAccCheckAWSGroupPolicyAttachmentAttributes([]string{policyName}, &out),
				),
			},
			{
				ResourceName:      "aws_iam_group_policy_attachment.test-attach",
				ImportState:       true,
				ImportStateIdFunc: testAccAWSIAMGroupPolicyAttachmentImportStateIdFunc("aws_iam_group_policy_attachment.test-attach"),
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
				Config: testAccAWSGroupPolicyAttachConfigUpdate(groupName, policyName, policyName2, policyName3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupPolicyAttachmentExists("aws_iam_group_policy_attachment.test-attach", 2, &out),
					testAccCheckAWSGroupPolicyAttachmentAttributes([]string{policyName2, policyName3}, &out),
				),
			},
		},
	})
}
func testAccCheckAWSGroupPolicyAttachmentDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckAWSGroupPolicyAttachmentExists(n string, c int, out *iam.ListAttachedGroupPoliciesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		group := rs.Primary.Attributes["group"]

		attachedPolicies, err := conn.ListAttachedGroupPolicies(&iam.ListAttachedGroupPoliciesInput{
			GroupName: aws.String(group),
		})
		if err != nil {
			return fmt.Errorf("Error: Failed to get attached policies for group %s (%s)", group, n)
		}
		if c != len(attachedPolicies.AttachedPolicies) {
			return fmt.Errorf("Error: Group (%s) has wrong number of policies attached on initial creation", n)
		}

		*out = *attachedPolicies
		return nil
	}
}
func testAccCheckAWSGroupPolicyAttachmentAttributes(policies []string, out *iam.ListAttachedGroupPoliciesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		matched := 0

		for _, p := range policies {
			for _, ap := range out.AttachedPolicies {
				// *ap.PolicyArn like arn:aws:iam::111111111111:policy/test-policy
				parts := strings.Split(*ap.PolicyArn, "/")
				if len(parts) == 2 && p == parts[1] {
					matched++
				}
			}
		}
		if matched != len(policies) || matched != len(out.AttachedPolicies) {
			return fmt.Errorf("Error: Number of attached policies was incorrect: expected %d matched policies, matched %d of %d", len(policies), matched, len(out.AttachedPolicies))
		}
		return nil
	}
}

func testAccAWSGroupPolicyAttachConfig(groupName, policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
    name = "%s"
}

resource "aws_iam_policy" "policy" {
    name = "%s"
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

resource "aws_iam_group_policy_attachment" "test-attach" {
    group = "${aws_iam_group.group.name}"
    policy_arn = "${aws_iam_policy.policy.arn}"
}
`, groupName, policyName)
}

func testAccAWSGroupPolicyAttachConfigUpdate(groupName, policyName, policyName2, policyName3 string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
    name = "%s"
}

resource "aws_iam_policy" "policy" {
    name = "%s"
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

resource "aws_iam_policy" "policy2" {
    name = "%s"
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

resource "aws_iam_policy" "policy3" {
    name = "%s"
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

resource "aws_iam_group_policy_attachment" "test-attach" {
    group = "${aws_iam_group.group.name}"
    policy_arn = "${aws_iam_policy.policy2.arn}"
}

resource "aws_iam_group_policy_attachment" "test-attach2" {
    group = "${aws_iam_group.group.name}"
    policy_arn = "${aws_iam_policy.policy3.arn}"
}
`, groupName, policyName, policyName2, policyName3)
}

func testAccAWSIAMGroupPolicyAttachmentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["group"], rs.Primary.Attributes["policy_arn"]), nil
	}
}
