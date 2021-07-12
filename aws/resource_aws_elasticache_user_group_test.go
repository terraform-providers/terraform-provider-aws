package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSElasticacheUserGroup_basic(t *testing.T) {
	var ug elasticache.UserGroup
	resourceName := "aws_elasticache_user_group.test-basic"
	rName := fmt.Sprintf("a-user-group-test-tf-basic")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserGroupExists(&ug, resourceName),
					testAccCheckAWSElasticacheUserGroupAttributes(&ug, rName),
					// Add User ID Count Test
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
		},
	})
}

func TestAccAWSElasticacheUserGroup_addUserId(t *testing.T) {
	var ug elasticache.UserGroup
	operation := "add"
	resourceName := "aws_elasticache_user_group.test-user-id-add"
	rName := fmt.Sprintf("a-user-group-test-tf-add-user-id")
	testuser1 := "test-user-1"
	testuser2 := "test-user-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserGroupConfigUserIdSingle(operation, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserGroupExists(&ug, resourceName),
					testAccCheckAWSElasticacheUserGroupAttributes(&ug, rName),
					// Add User ID Count Test
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
			{
				Config: testAccAWSElasticacheUserGroupConfigUserIdMulti(operation, rName, testuser1, testuser2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserGroupExists(&ug, resourceName),
					testAccCheckAWSElasticacheUserGroupAttributes(&ug, rName),
					// Add User ID Count Test
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
		},
	})
}

func TestAccAWSElasticacheUserGroup_changeUserId(t *testing.T) {
	var ug elasticache.UserGroup
	operation := "change"
	resourceName := "aws_elasticache_user_group.test-user-id-change"
	rName := fmt.Sprintf("a-user-group-test-tf-change-user-id")
	testuser1 := "test-user-1"
	testuser2 := "test-user-2"
	testuser3 := "test-user-3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserGroupConfigUserIdMulti(operation, rName, testuser1, testuser2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserGroupExists(&ug, resourceName),
					testAccCheckAWSElasticacheUserGroupAttributes(&ug, rName),
					// Add User ID Count Test
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
			{
				Config: testAccAWSElasticacheUserGroupConfigUserIdMulti(operation, rName, testuser1, testuser3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserGroupExists(&ug, resourceName),
					testAccCheckAWSElasticacheUserGroupAttributes(&ug, rName),
					// Add User ID Count Test
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
		},
	})
}

func TestAccAWSElasticacheUserGroup_removeUserId(t *testing.T) {
	var ug elasticache.UserGroup
	operation := "remove"
	resourceName := "aws_elasticache_user_group.test-user-id-remove"
	rName := fmt.Sprintf("a-user-group-test-tf-remove-user-id")
	testuser1 := "test-user-1"
	testuser2 := "test-user-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserGroupConfigUserIdMulti(operation, rName, testuser1, testuser2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserGroupExists(&ug, resourceName),
					testAccCheckAWSElasticacheUserGroupAttributes(&ug, rName),
					// Add User ID Count Test
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
			{
				Config: testAccAWSElasticacheUserGroupConfigUserIdSingle(operation, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserGroupExists(&ug, resourceName),
					testAccCheckAWSElasticacheUserGroupAttributes(&ug, rName),
					// Add User ID Count Test
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", rName),
				),
			},
		},
	})
}

func testAccCheckAWSElasticacheUserGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_user_group" {
			continue
		}

		params := &elasticache.DescribeUserGroupsInput{
			UserGroupId: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeUserGroups(params)

		if isAWSErr(err, elasticache.ErrCodeUserGroupNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if response != nil {
			for _, userGroups := range response.UserGroups {
				if aws.StringValue(userGroups.UserGroupId) == rs.Primary.ID {
					return fmt.Errorf("[ERROR] ElastiCache User Group (%s) still exists", rs.Primary.ID)
				}
			}
		}
	}
	return nil
}

func testAccCheckAWSElasticacheUserGroupExists(ug *elasticache.UserGroup, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("[ERROR] No ElastiCache User Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

		params := elasticache.DescribeUserGroupsInput{
			UserGroupId: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeUserGroups(&params)
		if err != nil {
			return fmt.Errorf("[ERROR] ElastiCache User Group: %v", err)
		}

		if len(response.UserGroups) != 1 ||
			*response.UserGroups[0].UserGroupId != rs.Primary.ID {
			return fmt.Errorf("[ERROR] ElastiCache User Group not found")
		}

		*ug = *response.UserGroups[0]

		return nil
	}
}

func testAccCheckAWSElasticacheUserGroupAttributes(ug *elasticache.UserGroup, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *ug.UserGroupId != rName {
			return fmt.Errorf("Bad ElastiCache User Group ID: %#v", ug.UserGroupId)
		}

		if *ug.Engine != "redis" {
			return fmt.Errorf("Bad ElastiCache User Group Engine: %#v", ug.Engine)
		}

		return nil
	}
}

func testAccAWSElasticacheUserGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user_group" "test-basic" {
  user_group_id = %[1]q
  engine        = "redis"
  user_ids      = ["default"]
}
`, rName)
}

func testAccAWSElasticacheUserGroupConfigUserIdSingle(operation string, rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user_group" "test-user-id-%[1]s" {
  user_group_id = %[2]q
  engine        = "redis"
  user_ids      = ["default"]
}
`, operation, rName)
}

// https://github.com/hashicorp/terraform-provider-aws/pull/16629 is Merged
// Add following resources:
// resource "aws_elasticache_user_group" "test-user-1" {
//   user_id       = "test-user-1"
//   user_name     = "test-user-1"
//   access-string = "off ~* +@all"
// }
//
// resource "aws_elasticache_user_group" "test-user-2" {
//   user_id       = "test-user-2"
//   user_name     = "test-user-2"
//   access-string = "off ~* +@all"
// }
//
// resource "aws_elasticache_user_group" "test-user-3" {
//   user_id       = "test-user-3"
//   user_name     = "test-user-3"
//   access-string = "off ~* +@all"
// }

// Need to manually create the new users in AWS Console

func testAccAWSElasticacheUserGroupConfigUserIdMulti(operation, rName, testuser1, testuser2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user_group" "test-user-id-%[1]s" {
  user_group_id = %[2]q
  engine        = "redis"
  user_ids      = ["default", %[3]q, %[4]q]
}
`, operation, rName, testuser1, testuser2)
}
