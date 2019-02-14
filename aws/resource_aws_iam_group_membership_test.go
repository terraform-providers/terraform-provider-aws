package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSGroupMembership_basic(t *testing.T) {
	var group iam.GetGroupOutput

	rString := acctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gm-basic-%s", rString)
	userName := fmt.Sprintf("tf-acc-user-gm-basic-%s", rString)
	userName2 := fmt.Sprintf("tf-acc-user-gm-basic-two-%s", rString)
	userName3 := fmt.Sprintf("tf-acc-user-gm-basic-three-%s", rString)
	membershipName := fmt.Sprintf("tf-acc-membership-gm-basic-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGroupMemberConfig(groupName, userName, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					testAccCheckAWSGroupMembershipAttributes(&group, groupName, []string{userName}),
				),
			},

			{
				Config: testAccAWSGroupMemberConfigUpdate(groupName, userName, userName2, userName3, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					testAccCheckAWSGroupMembershipAttributes(&group, groupName, []string{userName2, userName3}),
				),
			},

			{
				Config: testAccAWSGroupMemberConfigUpdateDown(groupName, userName3, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					testAccCheckAWSGroupMembershipAttributes(&group, groupName, []string{userName3}),
				),
			},
		},
	})
}

func TestAccAWSGroupMembership_paginatedUserList(t *testing.T) {
	var group iam.GetGroupOutput

	rString := acctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gm-pul-%s", rString)
	membershipName := fmt.Sprintf("tf-acc-membership-gm-pul-%s", rString)
	userNamePrefix := fmt.Sprintf("tf-acc-user-gm-pul-%s-", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGroupMemberConfigPaginatedUserList(groupName, membershipName, userNamePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					resource.TestCheckResourceAttr(
						"aws_iam_group_membership.team", "users.#", "101"),
				),
			},
		},
	})
}

func testAccCheckAWSGroupMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_group_membership" {
			continue
		}

		group := rs.Primary.Attributes["group"]

		_, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(group),
		})
		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "NoSuchEntity" {
				continue
			}
			return err
		}

		return fmt.Errorf("still exists")
	}

	return nil
}

func testAccCheckAWSGroupMembershipExists(n string, g *iam.GetGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		gn := rs.Primary.Attributes["group"]

		resp, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(gn),
		})

		if err != nil {
			return fmt.Errorf("Error: Group (%s) not found", gn)
		}

		*g = *resp

		return nil
	}
}

func testAccCheckAWSGroupMembershipAttributes(group *iam.GetGroupOutput, groupName string, users []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.Contains(*group.Group.GroupName, groupName) {
			return fmt.Errorf("Bad group membership: expected %s, got %s", groupName, *group.Group.GroupName)
		}

		uc := len(users)
		for _, u := range users {
			for _, gu := range group.Users {
				if u == *gu.UserName {
					uc--
				}
			}
		}

		if uc > 0 {
			return fmt.Errorf("Bad group membership count, expected (%d), but only (%d) found", len(users), uc)
		}
		return nil
	}
}

func testAccAWSGroupMemberConfig(groupName, userName, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_user" "user" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = ["${aws_iam_user.user.name}"]
	group = "${aws_iam_group.group.name}"
}
`, groupName, userName, membershipName)
}

func testAccAWSGroupMemberConfigUpdate(groupName, userName, userName2, userName3, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_user" "user" {
	name = "%s"
}

resource "aws_iam_user" "user_two" {
	name = "%s"
}

resource "aws_iam_user" "user_three" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = [
		"${aws_iam_user.user_two.name}",
		"${aws_iam_user.user_three.name}",
	]
	group = "${aws_iam_group.group.name}"
}
`, groupName, userName, userName2, userName3, membershipName)
}

func testAccAWSGroupMemberConfigUpdateDown(groupName, userName3, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_user" "user_three" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = [
		"${aws_iam_user.user_three.name}",
	]
	group = "${aws_iam_group.group.name}"
}
`, groupName, userName3, membershipName)
}

func testAccAWSGroupMemberConfigPaginatedUserList(groupName, membershipName, userNamePrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	group = "${aws_iam_group.group.name}"
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
}

resource "aws_iam_user" "user" {
	count = 101
	name = "${format("%s%%d", count.index + 1)}"
}
`, groupName, membershipName, userNamePrefix)
}
