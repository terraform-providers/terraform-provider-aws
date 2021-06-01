package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestAccAWSIAMPolicyAttachment_Groups_RenamedGroup(t *testing.T) {
	var out iam.ListEntitiesForPolicyOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	groupName1 := fmt.Sprintf("%s-1", rName)
	groupName2 := fmt.Sprintf("%s-2", rName)
	resourceName := "aws_iam_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIamPolicyAttachmentConfigGroupsRenamedGroup(rName, groupName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists(resourceName, 1, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{}, []string{}, []string{groupName1}, &out),
				),
			},
			{
				Config: testAccAWSIamPolicyAttachmentConfigGroupsRenamedGroup(rName, groupName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists(resourceName, 1, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{}, []string{}, []string{groupName2}, &out),
				),
			},
		},
	})
}

func TestAccAWSIAMPolicyAttachment_Roles_RenamedRole(t *testing.T) {
	var out iam.ListEntitiesForPolicyOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	roleName1 := fmt.Sprintf("%s-1", rName)
	roleName2 := fmt.Sprintf("%s-2", rName)
	resourceName := "aws_iam_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIamPolicyAttachmentConfigRolesRenamedRole(rName, roleName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists(resourceName, 1, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{}, []string{roleName1}, []string{}, &out),
				),
			},
			{
				Config: testAccAWSIamPolicyAttachmentConfigRolesRenamedRole(rName, roleName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists(resourceName, 1, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{}, []string{roleName2}, []string{}, &out),
				),
			},
		},
	})
}

func TestAccAWSIAMPolicyAttachment_Users_RenamedUser(t *testing.T) {
	var out iam.ListEntitiesForPolicyOutput

	rName := acctest.RandomWithPrefix("tf-acc-test")
	userName1 := fmt.Sprintf("%s-1", rName)
	userName2 := fmt.Sprintf("%s-2", rName)
	resourceName := "aws_iam_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIamPolicyAttachmentConfigUsersRenamedUser(rName, userName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists(resourceName, 1, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{userName1}, []string{}, []string{}, &out),
				),
			},
			{
				Config: testAccAWSIamPolicyAttachmentConfigUsersRenamedUser(rName, userName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSPolicyAttachmentExists(resourceName, 1, &out),
					testAccCheckAWSPolicyAttachmentAttributes([]string{userName2}, []string{}, []string{}, &out),
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

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "%s"
  users      = [aws_iam_user.user.name]
  roles      = [aws_iam_role.role.name]
  groups     = [aws_iam_group.group.name]
  policy_arn = aws_iam_policy.policy.arn
}
`, userName, roleName, groupName, policyName, attachmentName)
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

resource "aws_iam_policy_attachment" "test-attach" {
  name = "%s"

  users = [
    aws_iam_user.user2.name,
    aws_iam_user.user3.name,
  ]

  roles = [
    aws_iam_role.role2.name,
    aws_iam_role.role3.name,
  ]

  groups = [
    aws_iam_group.group2.name,
    aws_iam_group.group3.name,
  ]

  policy_arn = aws_iam_policy.policy.arn
}
`, userName, userName2, userName3,
		roleName, roleName2, roleName3,
		groupName, groupName2, groupName3,
		policyName, attachmentName)
}

func testAccAWSPolicyPaginatedAttachConfig(userNamePrefix, policyName, attachmentName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  count = 101
  name  = format("%s%%d", count.index + 1)
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

resource "aws_iam_policy_attachment" "test-paginated-attach" {
  name       = "%s"
  policy_arn = aws_iam_policy.policy.arn

  # TODO: Switch back to simple list reference when test configurations are upgraded to 0.12 syntax
  users = [
    aws_iam_user.user[0].name,
    aws_iam_user.user[1].name,
    aws_iam_user.user[2].name,
    aws_iam_user.user[3].name,
    aws_iam_user.user[4].name,
    aws_iam_user.user[5].name,
    aws_iam_user.user[6].name,
    aws_iam_user.user[7].name,
    aws_iam_user.user[8].name,
    aws_iam_user.user[9].name,
    aws_iam_user.user[10].name,
    aws_iam_user.user[11].name,
    aws_iam_user.user[12].name,
    aws_iam_user.user[13].name,
    aws_iam_user.user[14].name,
    aws_iam_user.user[15].name,
    aws_iam_user.user[16].name,
    aws_iam_user.user[17].name,
    aws_iam_user.user[18].name,
    aws_iam_user.user[19].name,
    aws_iam_user.user[20].name,
    aws_iam_user.user[21].name,
    aws_iam_user.user[22].name,
    aws_iam_user.user[23].name,
    aws_iam_user.user[24].name,
    aws_iam_user.user[25].name,
    aws_iam_user.user[26].name,
    aws_iam_user.user[27].name,
    aws_iam_user.user[28].name,
    aws_iam_user.user[29].name,
    aws_iam_user.user[30].name,
    aws_iam_user.user[31].name,
    aws_iam_user.user[32].name,
    aws_iam_user.user[33].name,
    aws_iam_user.user[34].name,
    aws_iam_user.user[35].name,
    aws_iam_user.user[36].name,
    aws_iam_user.user[37].name,
    aws_iam_user.user[38].name,
    aws_iam_user.user[39].name,
    aws_iam_user.user[40].name,
    aws_iam_user.user[41].name,
    aws_iam_user.user[42].name,
    aws_iam_user.user[43].name,
    aws_iam_user.user[44].name,
    aws_iam_user.user[45].name,
    aws_iam_user.user[46].name,
    aws_iam_user.user[47].name,
    aws_iam_user.user[48].name,
    aws_iam_user.user[49].name,
    aws_iam_user.user[50].name,
    aws_iam_user.user[51].name,
    aws_iam_user.user[52].name,
    aws_iam_user.user[53].name,
    aws_iam_user.user[54].name,
    aws_iam_user.user[55].name,
    aws_iam_user.user[56].name,
    aws_iam_user.user[57].name,
    aws_iam_user.user[58].name,
    aws_iam_user.user[59].name,
    aws_iam_user.user[60].name,
    aws_iam_user.user[61].name,
    aws_iam_user.user[62].name,
    aws_iam_user.user[63].name,
    aws_iam_user.user[64].name,
    aws_iam_user.user[65].name,
    aws_iam_user.user[66].name,
    aws_iam_user.user[67].name,
    aws_iam_user.user[68].name,
    aws_iam_user.user[69].name,
    aws_iam_user.user[70].name,
    aws_iam_user.user[71].name,
    aws_iam_user.user[72].name,
    aws_iam_user.user[73].name,
    aws_iam_user.user[74].name,
    aws_iam_user.user[75].name,
    aws_iam_user.user[76].name,
    aws_iam_user.user[77].name,
    aws_iam_user.user[78].name,
    aws_iam_user.user[79].name,
    aws_iam_user.user[80].name,
    aws_iam_user.user[81].name,
    aws_iam_user.user[82].name,
    aws_iam_user.user[83].name,
    aws_iam_user.user[84].name,
    aws_iam_user.user[85].name,
    aws_iam_user.user[86].name,
    aws_iam_user.user[87].name,
    aws_iam_user.user[88].name,
    aws_iam_user.user[89].name,
    aws_iam_user.user[90].name,
    aws_iam_user.user[91].name,
    aws_iam_user.user[92].name,
    aws_iam_user.user[93].name,
    aws_iam_user.user[94].name,
    aws_iam_user.user[95].name,
    aws_iam_user.user[96].name,
    aws_iam_user.user[97].name,
    aws_iam_user.user[98].name,
    aws_iam_user.user[99].name,
    aws_iam_user.user[100].name,
  ]
}
`, userNamePrefix, policyName, attachmentName)
}

func testAccAWSIamPolicyAttachmentConfigGroupsRenamedGroup(rName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "*",
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_group" "test" {
  name = %[2]q
}

resource "aws_iam_policy_attachment" "test" {
  groups     = [aws_iam_group.test.name]
  name       = %[1]q
  policy_arn = aws_iam_policy.test.arn
}
`, rName, groupName)
}

func testAccAWSIamPolicyAttachmentConfigRolesRenamedRole(rName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "*",
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  force_detach_policies = true
  name                  = %[2]q

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

resource "aws_iam_policy_attachment" "test" {
  name       = %[1]q
  policy_arn = aws_iam_policy.test.arn
  roles      = [aws_iam_role.test.name]
}
`, rName, roleName)
}

func testAccAWSIamPolicyAttachmentConfigUsersRenamedUser(rName, userName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "*",
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_user" "test" {
  force_destroy = true
  name          = %[2]q
}

resource "aws_iam_policy_attachment" "test" {
  name       = %[1]q
  policy_arn = aws_iam_policy.test.arn
  users      = [aws_iam_user.test.name]
}
`, rName, userName)
}
