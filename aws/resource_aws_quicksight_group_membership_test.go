package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

func TestAccAWSQuickSightGroupMembership_basic(t *testing.T) {
	groupName := acctest.RandomWithPrefix("tf-acc-test")
	memberName := "tfacctest" + acctest.RandString(10)
	resourceName := "aws_quicksight_group_membership.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckQuickSightGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightGroupMembershipConfig(groupName, memberName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupMembershipExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSQuickSightGroupMembership_disappears(t *testing.T) {
	groupName := acctest.RandomWithPrefix("tf-acc-test")
	memberName := "tfacctest" + acctest.RandString(10)
	resourceName := "aws_quicksight_group_membership.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckQuickSightGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightGroupMembershipConfig(groupName, memberName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightGroupMembershipExists(resourceName),
					testAccCheckQuickSightGroupMembershipDisappears(groupName, memberName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickSightGroupMembershipExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).quicksightconn

		input := &quicksight.ListUserGroupsInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			UserName:     aws.String(userName),
		}

		output, err := conn.ListUserGroups(input)

		if err != nil {
			return err
		}

		if output == nil || output.GroupList == nil {
			return fmt.Errorf("QuickSight Group (%s) not found", rs.Primary.ID)
		}

		for _, group := range output.GroupList {
			if *group.GroupName == groupName {
				return nil
			}
		}

		return fmt.Errorf("QuickSight Group (%s) not found in user's group list", rs.Primary.ID)
	}
}

func testAccCheckQuickSightGroupMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).quicksightconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_group_membership" {
			continue
		}

		awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &quicksight.ListUserGroupsInput{
			AwsAccountId: aws.String(awsAccountID),
			Namespace:    aws.String(namespace),
			UserName:     aws.String(userName),
		}

		output, err := conn.ListUserGroups(input)

		if err != nil {
			return err
		}

		if output == nil || output.GroupList == nil {
			return fmt.Errorf("QuickSight Group (%s) not found", rs.Primary.ID)
		}

		for _, group := range output.GroupList {
			if *group.GroupName == groupName {
				return fmt.Errorf("QuickSight Group membership (%s) was not deleted properly", rs.Primary.ID)
			}
		}

	}

	return nil
}

func testAccCheckQuickSightGroupMembershipDisappears(groupName string, memberName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).quicksightconn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_group_membership" {
				continue
			}

			awsAccountID, namespace, groupName, userName, err := resourceAwsQuickSightGroupMembershipParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			input := &quicksight.DeleteGroupMembershipInput{
				AwsAccountId: aws.String(awsAccountID),
				Namespace:    aws.String(namespace),
				MemberName:   aws.String(userName),
				GroupName:    aws.String(groupName),
			}

			if _, err := conn.DeleteGroupMembership(input); err != nil {
				return err
			}
		}
		return nil
	}
}

func testAccAWSQuickSightGroupMembershipConfig(groupName string, memberName string) string {
	return fmt.Sprintf(`
%s

%s

resource "aws_quicksight_group_membership" "default" {
  group_name  = aws_quicksight_group.default.group_name
  member_name = aws_quicksight_user.%s.user_name
}
`, testAccAWSQuickSightGroupConfig(groupName), testAccAWSQuickSightUserConfig(memberName), memberName)
}
