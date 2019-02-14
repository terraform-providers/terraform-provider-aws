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

func TestAccAWSUserGroupMembership_basic(t *testing.T) {
	rString := acctest.RandString(8)
	userName1 := fmt.Sprintf("tf-acc-ugm-basic-user1-%s", rString)
	userName2 := fmt.Sprintf("tf-acc-ugm-basic-user2-%s", rString)
	groupName1 := fmt.Sprintf("tf-acc-ugm-basic-group1-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-ugm-basic-group2-%s", rString)
	groupName3 := fmt.Sprintf("tf-acc-ugm-basic-group3-%s", rString)

	usersAndGroupsConfig := testAccAWSUserGroupMembershipConfigUsersAndGroups(userName1, userName2, groupName1, groupName2, groupName3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSUserGroupMembershipDestroy,
		Steps: []resource.TestStep{
			// simplest test
			{
				Config: usersAndGroupsConfig + testAccAWSUserGroupMembershipConfigInit,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1}, []string{groupName2, groupName3}),
				),
			},
			// test adding an additional group to an existing resource
			{
				Config: usersAndGroupsConfig + testAccAWSUserGroupMembershipConfigAddOne,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName2}, []string{groupName3}),
				),
			},
			// test adding multiple resources for the same user, and resources with the same groups for another user
			{
				Config: usersAndGroupsConfig + testAccAWSUserGroupMembershipConfigAddAll,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test2", "user", userName2),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName2, groupName3}, []string{}),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1, groupName2, groupName3}, []string{}),
				),
			},
			// test that nothing happens when we apply the same config again
			{
				Config: usersAndGroupsConfig + testAccAWSUserGroupMembershipConfigAddAll,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test2", "user", userName2),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName2, groupName3}, []string{}),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1, groupName2, groupName3}, []string{}),
				),
			},
			// test removing a group
			{
				Config: usersAndGroupsConfig + testAccAWSUserGroupMembershipConfigRemoveGroup,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test2", "user", userName2),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName3}, []string{groupName2}),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1, groupName2}, []string{groupName3}),
				),
			},
			// test removing a resource
			{
				Config: usersAndGroupsConfig + testAccAWSUserGroupMembershipConfigDeleteResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName3}, []string{groupName2}),
					testAccAWSUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1}, []string{groupName2, groupName3}),
				),
			},
		},
	})
}

func testAccAWSUserGroupMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "aws_iam_user_group_membership" {
			input := &iam.ListGroupsForUserInput{
				UserName: aws.String(rs.Primary.Attributes["user"]),
			}
			foundGroups := 0
			err := conn.ListGroupsForUserPages(input, func(page *iam.ListGroupsForUserOutput, lastPage bool) bool {
				if len(page.Groups) > 0 {
					foundGroups = foundGroups + len(page.Groups)
				}
				return !lastPage
			})
			if err != nil {
				if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
					continue
				}
				return err
			}
			if foundGroups > 0 {
				return fmt.Errorf("Expected all group membership for user to be removed, found: %d", foundGroups)
			}
		}
	}

	return nil
}

func testAccAWSUserGroupMembershipCheckGroupListForUser(userName string, groups []string, groupsNeg []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iamconn

		// get list of groups for user
		userGroupList, err := conn.ListGroupsForUser(&iam.ListGroupsForUserInput{
			UserName: &userName,
		})
		if err != nil {
			return fmt.Errorf("Error validing user group list for %s: %s", userName, err)
		}

		// check required groups
	GROUP:
		for _, group := range groups {
			for _, groupFound := range userGroupList.Groups {
				if group == *groupFound.GroupName {
					continue GROUP // found our group, start checking the next one
				}
			}
			// group not found, return an error
			return fmt.Errorf("Required group not found for %s: %s", userName, group)
		}

		// check that none of groupsNeg are present
		for _, group := range groupsNeg {
			for _, groupFound := range userGroupList.Groups {
				if group == *groupFound.GroupName {
					return fmt.Errorf("Unexpected group found for %s: %s", userName, group)
				}
			}
		}

		return nil
	}
}

// users and groups for all other tests
func testAccAWSUserGroupMembershipConfigUsersAndGroups(userName1, userName2, groupName1, groupName2, groupName3 string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user1" {
	name          = "%s"
	force_destroy = true
}

resource "aws_iam_user" "user2" {
	name          = "%s"
	force_destroy = true
}

resource "aws_iam_group" "group1" {
	name = "%s"
}

resource "aws_iam_group" "group2" {
	name = "%s"
}

resource "aws_iam_group" "group3" {
	name = "%s"
}
`, userName1, userName2, groupName1, groupName2, groupName3)
}

// associate users and groups
const testAccAWSUserGroupMembershipConfigInit = `
resource "aws_iam_user_group_membership" "user1_test1" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group1.name}",
	]
}
`

const testAccAWSUserGroupMembershipConfigAddOne = `
resource "aws_iam_user_group_membership" "user1_test1" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group1.name}",
		"${aws_iam_group.group2.name}",
	]
}
`

const testAccAWSUserGroupMembershipConfigAddAll = `
resource "aws_iam_user_group_membership" "user1_test1" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group1.name}",
		"${aws_iam_group.group2.name}",
	]
}

resource "aws_iam_user_group_membership" "user1_test2" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group3.name}",
	]
}

resource "aws_iam_user_group_membership" "user2_test1" {
	user = "${aws_iam_user.user2.name}"
	groups = [
		"${aws_iam_group.group1.name}",
	]
}

resource "aws_iam_user_group_membership" "user2_test2" {
	user = "${aws_iam_user.user2.name}"
	groups = [
		"${aws_iam_group.group2.name}",
		"${aws_iam_group.group3.name}",
	]
}
`

// test removing a group
const testAccAWSUserGroupMembershipConfigRemoveGroup = `
resource "aws_iam_user_group_membership" "user1_test1" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group1.name}",
	]
}

resource "aws_iam_user_group_membership" "user1_test2" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group3.name}",
	]
}

resource "aws_iam_user_group_membership" "user2_test1" {
	user = "${aws_iam_user.user2.name}"
	groups = [
		"${aws_iam_group.group1.name}",
	]
}

resource "aws_iam_user_group_membership" "user2_test2" {
	user = "${aws_iam_user.user2.name}"
	groups = [
		"${aws_iam_group.group2.name}",
	]
}
`

// test deleting an entity
const testAccAWSUserGroupMembershipConfigDeleteResource = `
resource "aws_iam_user_group_membership" "user1_test1" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group1.name}",
	]
}

resource "aws_iam_user_group_membership" "user1_test2" {
	user = "${aws_iam_user.user1.name}"
	groups = [
		"${aws_iam_group.group3.name}",
	]
}

resource "aws_iam_user_group_membership" "user2_test1" {
	user = "${aws_iam_user.user2.name}"
	groups = [
		"${aws_iam_group.group1.name}",
	]
}
`
