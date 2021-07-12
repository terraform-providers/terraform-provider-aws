package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsDataSourceSsoAdminRole_Group(t *testing.T) {
	permissionSetName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ssoadmin_role.test"
	roleNamePrefix := fmt.Sprintf("AWSReservedSSO_%s", permissionSetName)
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		ErrorCheck: testAccErrorCheck(t, ssoadmin.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSsoAdminRoleGroupAttachmentConfig(permissionSetName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(fmt.Sprintf("%s_[a-f0-9]{16}", roleNamePrefix))),
					resource.TestCheckResourceAttr(dataSourceName, "path", "/aws-reserved/sso.amazonaws.com/"),
				),
			},
		},
	})
}

func TestAccAwsDataSourceSsoAdminRole_User(t *testing.T) {
	permissionSetName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ssoadmin_role.test"
	roleNamePrefix := fmt.Sprintf("AWSReservedSSO_%s", permissionSetName)
	userName := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ssoadmin.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSsoAdminRoleUserAttachmentConfig(permissionSetName, userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(fmt.Sprintf("%s_[a-f0-9]{16}", roleNamePrefix))),
					resource.TestCheckResourceAttr(dataSourceName, "path", "/aws-reserved/sso.amazonaws.com/"),
				),
			},
		},
	})
}

func TestAccAwsDataSourceSsoAdminRole_UserAndGroup(t *testing.T) {
	permissionSetName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ssoadmin_role.test"
	roleNamePrefix := fmt.Sprintf("AWSReservedSSO_%s", permissionSetName)
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")
	userName := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, ssoadmin.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsSsoAdminRoleUserAndGroupAttachmentsConfig(permissionSetName, groupName, userName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(fmt.Sprintf("%s_[a-f0-9]{16}", roleNamePrefix))),
					resource.TestCheckResourceAttr(dataSourceName, "path", "/aws-reserved/sso.amazonaws.com/"),
				),
			},
		},
	})
}

func testAccAwsSsoAdminRoleBaseConfig(permissionSetName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  description  = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state  = "https://example.com"
}

data "aws_caller_identity" "current" {}
`, permissionSetName)
}

func testAccAwsSsoAdminRoleGroupAttachmentBaseConfig(groupName string) string {
	return fmt.Sprintf(`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %[1]q
  }
}

resource "aws_ssoadmin_account_assignment" "grouptest" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "GROUP"
  principal_id       = data.aws_identitystore_group.test.group_id
}
`, groupName)
}

func testAccAwsSsoAdminRoleUserAttachmentBaseConfig(userName string) string {
	return fmt.Sprintf(`
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  filter {
    attribute_path  = "UserName"
    attribute_value = %[1]q
  }
}

resource "aws_ssoadmin_account_assignment" "usertest" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "USER"
  principal_id       = data.aws_identitystore_user.test.user_id
}
`, userName)
}

func testAccAwsSsoAdminRoleGroupAttachmentConfig(permissionSetName, groupName string) string {
	return composeConfig(
		testAccAwsSsoAdminRoleBaseConfig(permissionSetName),
		testAccAwsSsoAdminRoleGroupAttachmentBaseConfig(groupName),
		`
data "aws_ssoadmin_role" "test" {
  permission_set_name = aws_ssoadmin_permission_set.test.name
  depends_on = [
    aws_ssoadmin_account_assignment.grouptest
  ]
}
`)
}

func testAccAwsSsoAdminRoleUserAttachmentConfig(permissionSetName, userName string) string {
	return composeConfig(
		testAccAwsSsoAdminRoleBaseConfig(permissionSetName),
		testAccAwsSsoAdminRoleUserAttachmentBaseConfig(userName),
		`
data "aws_ssoadmin_role" "test" {
  permission_set_name = aws_ssoadmin_permission_set.test.name
  depends_on = [
    aws_ssoadmin_account_assignment.usertest
  ]
}
`)
}

func testAccAwsSsoAdminRoleUserAndGroupAttachmentsConfig(permissionSetName, groupName, userName string) string {
	return composeConfig(
		testAccAwsSsoAdminRoleBaseConfig(permissionSetName),
		testAccAwsSsoAdminRoleGroupAttachmentBaseConfig(groupName),
		testAccAwsSsoAdminRoleUserAttachmentBaseConfig(userName),
		`
data "aws_ssoadmin_role" "test" {
  permission_set_name = aws_ssoadmin_permission_set.test.name
  depends_on = [
    aws_ssoadmin_account_assignment.grouptest,
    aws_ssoadmin_account_assignment.usertest
  ]
}
`)
}
