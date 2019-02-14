package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIAMPolicyAttachment_basic(t *testing.T) {
	var out iam.ListEntitiesForPolicyOutput

	rString := acctest.RandString(8)
	userName := fmt.Sprintf("tf-acc-user-pa-basic-%s", rString)
	userName2 := fmt.Sprintf("tf-acc-user-pa-basic-2-%s", rString)
	userName3 := fmt.Sprintf("tf-acc-user-pa-basic-3-%s", rString)
	roleName := fmt.Sprintf("tf-acc-role-pa-basic-%s", rString)
	roleName2 := fmt.Sprintf("tf-acc-role-pa-basic-2-%s", rString)
	roleName3 := fmt.Sprintf("tf-acc-role-pa-basic-3-%s", rString)
	groupName := fmt.Sprintf("tf-acc-group-pa-basic-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-group-pa-basic-2-%s", rString)
	groupName3 := fmt.Sprintf("tf-acc-group-pa-basic-3-%s", rString)
	policyName := fmt.Sprintf("tf-acc-policy-pa-basic-%s", rString)
	attachmentName := fmt.Sprintf("tf-acc-attachment-pa-basic-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPolicyAttachConfig(userName, roleName, groupName, policyName, attachmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists("aws_iam_policy_attachment.test-attach", 3, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{userName}, []string{roleName}, []string{groupName}, &out),
				),
			},
			{
				Config: testAccAWSPolicyAttachConfigUpdate(userName, userName2, userName3,
					roleName, roleName2, roleName3,
					groupName, groupName2, groupName3,
					policyName, attachmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists("aws_iam_policy_attachment.test-attach", 6, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{userName2, userName3},
						[]string{roleName2, roleName3}, []string{groupName2, groupName3}, &out),
				),
			},
		},
	})
}

func TestAccAWSIAMPolicyAttachment_paginatedEntities(t *testing.T) {
	var out iam.ListEntitiesForPolicyOutput

	rString := acctest.RandString(8)
	userNamePrefix := fmt.Sprintf("tf-acc-user-pa-pe-%s-", rString)
	policyName := fmt.Sprintf("tf-acc-policy-pa-pe-%s-", rString)
	attachmentName := fmt.Sprintf("tf-acc-attachment-pa-pe-%s-", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSPolicyPaginatedAttachConfig(userNamePrefix, policyName, attachmentName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists("aws_iam_policy_attachment.test-paginated-attach", 101, &out),
				),
			},
		},
	})
}

func testAccCheckAWSPolicyAttachmentDestroy(s *terraform.State) error {
	return nil
}

func testAccCheckAWSPolicyAttachmentExists(n string, c int64, out *iam.ListEntitiesForPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		arn := rs.Primary.Attributes["policy_arn"]

		resp, err := conn.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: aws.String(arn),
		})
		if err != nil {
			return fmt.Errorf("Error: Policy (%s) not found", n)
		}
		if c != *resp.Policy.AttachmentCount {
			return fmt.Errorf("Error: Policy (%s) has wrong number of entities attached on initial creation", n)
		}
		resp2, err := conn.ListEntitiesForPolicy(&iam.ListEntitiesForPolicyInput{
			PolicyArn: aws.String(arn),
		})
		if err != nil {
			return fmt.Errorf("Error: Failed to get entities for Policy (%s)", arn)
		}

		*out = *resp2
		return nil
	}
}

func testAccCheckAWSPolicyAttachmentAttributes(users []string, roles []string, groups []string, out *iam.ListEntitiesForPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		uc := len(users)
		rc := len(roles)
		gc := len(groups)

		for _, u := range users {
			for _, pu := range out.PolicyUsers {
				if u == *pu.UserName {
					uc--
				}
			}
		}
		for _, r := range roles {
			for _, pr := range out.PolicyRoles {
				if r == *pr.RoleName {
					rc--
				}
			}
		}
		for _, g := range groups {
			for _, pg := range out.PolicyGroups {
				if g == *pg.GroupName {
					gc--
				}
			}
		}
		if uc != 0 || rc != 0 || gc != 0 {
			return fmt.Errorf("Error: Number of attached users, roles, or groups was incorrect:\n expected %d users and found %d\nexpected %d roles and found %d\nexpected %d groups and found %d", len(users), len(users)-uc, len(roles), len(roles)-rc, len(groups), len(groups)-gc)
		}
		return nil
	}
}

func testAccAWSPolicyAttachConfig(userName, roleName, groupName, policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
    name = "%s"
}
resource "aws_iam_role" "role" {
    name = "%s"
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

resource "aws_iam_policy_attachment" "test-attach" {
    name = "%s"
    users = ["${aws_iam_user.user.name}"]
    roles = ["${aws_iam_role.role.name}"]
    groups = ["${aws_iam_group.group.name}"]
    policy_arn = "${aws_iam_policy.policy.arn}"
}`, userName, roleName, groupName, policyName, attachmentName)
}

func testAccAWSPolicyAttachConfigUpdate(userName, userName2, userName3,
	roleName, roleName2, roleName3,
	groupName, groupName2, groupName3,
	policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
    name = "%s"
}
resource "aws_iam_user" "user2" {
    name = "%s"
}
resource "aws_iam_user" "user3" {
    name = "%s"
}
resource "aws_iam_role" "role" {
    name = "%s"
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

resource "aws_iam_role" "role2" {
    name = "%s"
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
resource "aws_iam_role" "role3" {
    name = "%s"
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
resource "aws_iam_group" "group" {
    name = "%s"
}
resource "aws_iam_group" "group2" {
    name = "%s"
}
resource "aws_iam_group" "group3" {
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

resource "aws_iam_policy_attachment" "test-attach" {
    name = "%s"
    users = [
        "${aws_iam_user.user2.name}",
        "${aws_iam_user.user3.name}"
    ]
    roles = [
        "${aws_iam_role.role2.name}",
        "${aws_iam_role.role3.name}"
    ]
    groups = [
        "${aws_iam_group.group2.name}",
        "${aws_iam_group.group3.name}"
    ]
    policy_arn = "${aws_iam_policy.policy.arn}"
}`, userName, userName2, userName3,
		roleName, roleName2, roleName3,
		groupName, groupName2, groupName3,
		policyName, attachmentName)
}

func testAccAWSPolicyPaginatedAttachConfig(userNamePrefix, policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
	count = 101
	name = "${format("%s%%d", count.index + 1)}"
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
resource "aws_iam_policy_attachment" "test-paginated-attach" {
	name = "%s"
	policy_arn = "${aws_iam_policy.policy.arn}"
	# TODO: Switch back to simple list reference when test configurations are upgraded to 0.12 syntax
	users = [
		"${aws_iam_user.user.*.name[0]}",
		"${aws_iam_user.user.*.name[1]}",
		"${aws_iam_user.user.*.name[2]}",
		"${aws_iam_user.user.*.name[3]}",
		"${aws_iam_user.user.*.name[4]}",
		"${aws_iam_user.user.*.name[5]}",
		"${aws_iam_user.user.*.name[6]}",
		"${aws_iam_user.user.*.name[7]}",
		"${aws_iam_user.user.*.name[8]}",
		"${aws_iam_user.user.*.name[9]}",
		"${aws_iam_user.user.*.name[10]}",
		"${aws_iam_user.user.*.name[11]}",
		"${aws_iam_user.user.*.name[12]}",
		"${aws_iam_user.user.*.name[13]}",
		"${aws_iam_user.user.*.name[14]}",
		"${aws_iam_user.user.*.name[15]}",
		"${aws_iam_user.user.*.name[16]}",
		"${aws_iam_user.user.*.name[17]}",
		"${aws_iam_user.user.*.name[18]}",
		"${aws_iam_user.user.*.name[19]}",
		"${aws_iam_user.user.*.name[20]}",
		"${aws_iam_user.user.*.name[21]}",
		"${aws_iam_user.user.*.name[22]}",
		"${aws_iam_user.user.*.name[23]}",
		"${aws_iam_user.user.*.name[24]}",
		"${aws_iam_user.user.*.name[25]}",
		"${aws_iam_user.user.*.name[26]}",
		"${aws_iam_user.user.*.name[27]}",
		"${aws_iam_user.user.*.name[28]}",
		"${aws_iam_user.user.*.name[29]}",
		"${aws_iam_user.user.*.name[30]}",
		"${aws_iam_user.user.*.name[31]}",
		"${aws_iam_user.user.*.name[32]}",
		"${aws_iam_user.user.*.name[33]}",
		"${aws_iam_user.user.*.name[34]}",
		"${aws_iam_user.user.*.name[35]}",
		"${aws_iam_user.user.*.name[36]}",
		"${aws_iam_user.user.*.name[37]}",
		"${aws_iam_user.user.*.name[38]}",
		"${aws_iam_user.user.*.name[39]}",
		"${aws_iam_user.user.*.name[40]}",
		"${aws_iam_user.user.*.name[41]}",
		"${aws_iam_user.user.*.name[42]}",
		"${aws_iam_user.user.*.name[43]}",
		"${aws_iam_user.user.*.name[44]}",
		"${aws_iam_user.user.*.name[45]}",
		"${aws_iam_user.user.*.name[46]}",
		"${aws_iam_user.user.*.name[47]}",
		"${aws_iam_user.user.*.name[48]}",
		"${aws_iam_user.user.*.name[49]}",
		"${aws_iam_user.user.*.name[50]}",
		"${aws_iam_user.user.*.name[51]}",
		"${aws_iam_user.user.*.name[52]}",
		"${aws_iam_user.user.*.name[53]}",
		"${aws_iam_user.user.*.name[54]}",
		"${aws_iam_user.user.*.name[55]}",
		"${aws_iam_user.user.*.name[56]}",
		"${aws_iam_user.user.*.name[57]}",
		"${aws_iam_user.user.*.name[58]}",
		"${aws_iam_user.user.*.name[59]}",
		"${aws_iam_user.user.*.name[60]}",
		"${aws_iam_user.user.*.name[61]}",
		"${aws_iam_user.user.*.name[62]}",
		"${aws_iam_user.user.*.name[63]}",
		"${aws_iam_user.user.*.name[64]}",
		"${aws_iam_user.user.*.name[65]}",
		"${aws_iam_user.user.*.name[66]}",
		"${aws_iam_user.user.*.name[67]}",
		"${aws_iam_user.user.*.name[68]}",
		"${aws_iam_user.user.*.name[69]}",
		"${aws_iam_user.user.*.name[70]}",
		"${aws_iam_user.user.*.name[71]}",
		"${aws_iam_user.user.*.name[72]}",
		"${aws_iam_user.user.*.name[73]}",
		"${aws_iam_user.user.*.name[74]}",
		"${aws_iam_user.user.*.name[75]}",
		"${aws_iam_user.user.*.name[76]}",
		"${aws_iam_user.user.*.name[77]}",
		"${aws_iam_user.user.*.name[78]}",
		"${aws_iam_user.user.*.name[79]}",
		"${aws_iam_user.user.*.name[80]}",
		"${aws_iam_user.user.*.name[81]}",
		"${aws_iam_user.user.*.name[82]}",
		"${aws_iam_user.user.*.name[83]}",
		"${aws_iam_user.user.*.name[84]}",
		"${aws_iam_user.user.*.name[85]}",
		"${aws_iam_user.user.*.name[86]}",
		"${aws_iam_user.user.*.name[87]}",
		"${aws_iam_user.user.*.name[88]}",
		"${aws_iam_user.user.*.name[89]}",
		"${aws_iam_user.user.*.name[90]}",
		"${aws_iam_user.user.*.name[91]}",
		"${aws_iam_user.user.*.name[92]}",
		"${aws_iam_user.user.*.name[93]}",
		"${aws_iam_user.user.*.name[94]}",
		"${aws_iam_user.user.*.name[95]}",
		"${aws_iam_user.user.*.name[96]}",
		"${aws_iam_user.user.*.name[97]}",
		"${aws_iam_user.user.*.name[98]}",
		"${aws_iam_user.user.*.name[99]}",
		"${aws_iam_user.user.*.name[100]}",
	]
}`, userNamePrefix, policyName, attachmentName)
}
