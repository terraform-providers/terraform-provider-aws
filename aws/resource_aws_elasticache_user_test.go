package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSElasticacheUser_basic(t *testing.T) {
	var u elasticache.User
	resourceName := "aws_elasticache_user.test-basic"
	rName := fmt.Sprintf("a-user-test-tf-basic")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(&u, resourceName),
					testAccCheckAWSElasticacheUserAttributes(&u, rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* +@all"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheUser_updateAccessString(t *testing.T) {
	var u elasticache.User
	resourceName := "aws_elasticache_user.test-access-string"
	rName := fmt.Sprintf("a-user-test-tf-update-access-string")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfigUpdateAccessStringPre(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(&u, resourceName),
					testAccCheckAWSElasticacheUserAttributes(&u, rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* +@all"),
				),
			},
			{
				Config: testAccAWSElasticacheUserConfigUpdateAccessStringPost(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(&u, resourceName),
					testAccCheckAWSElasticacheUserAttributes(&u, rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "off ~* +@all"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheUser_addPasswords(t *testing.T) {
	var u elasticache.User
	resourceName := "aws_elasticache_user.test-passwords-add"
	rName := fmt.Sprintf("a-user-test-tf-update-passwords-add")
	pass1 := fmt.Sprintf("strongpassword12345678")
	pass2 := fmt.Sprintf("strongpassword23456789")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfigNoPasswords("add", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(&u, resourceName),
					testAccCheckAWSElasticacheUserAttributes(&u, rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* +@all"),
				),
			},
			{
				Config: testAccAWSElasticacheUserConfigPasswords("add", rName, pass1, pass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(&u, resourceName),
					testAccCheckAWSElasticacheUserAttributes(&u, rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* +@all"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheUser_removePasswords(t *testing.T) {
	var u elasticache.User
	resourceName := "aws_elasticache_user.test-passwords-remove"
	rName := fmt.Sprintf("a-user-test-tf-update-passwords-remove")
	pass1 := fmt.Sprintf("strongpassw0rd12345678")
	pass2 := fmt.Sprintf("strongpassw0rd23456789")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheUserConfigPasswords("remove", rName, pass1, pass2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(&u, resourceName),
					testAccCheckAWSElasticacheUserAttributes(&u, rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* +@all"),
				),
			},
			{
				Config: testAccAWSElasticacheUserConfigNoPasswords("remove", rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheUserExists(&u, resourceName),
					testAccCheckAWSElasticacheUserAttributes(&u, rName),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "user_id", rName),
					resource.TestCheckResourceAttr(resourceName, "user_name", rName),
					resource.TestCheckResourceAttr(resourceName, "access_string", "on ~* +@all"),
				),
			},
		},
	})
}

func testAccCheckAWSElasticacheUserDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_user" {
			continue
		}

		params := elasticache.DescribeUsersInput{
			UserId: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeUsers(&params)

		if isAWSErr(err, elasticache.ErrCodeUserNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if response != nil {
			for _, users := range response.Users {
				if aws.StringValue(users.UserId) == rs.Primary.ID {
					return fmt.Errorf("[ERROR] ElastiCache User (%s) still exists", rs.Primary.ID)
				}
			}
		}
	}
	return nil
}

// ElastiCache User Validations

func testAccCheckAWSElasticacheUserExists(u *elasticache.User, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("[ERROR] No ElastiCache User ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

		params := elasticache.DescribeUsersInput{
			UserId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeUsers(&params)
		if err != nil {
			return fmt.Errorf("[ERROR] ElastiCache User: %v", err)
		}

		if len(resp.Users) != 1 ||
			*resp.Users[0].UserId != rs.Primary.ID {
			return fmt.Errorf("[ERROR] ElastiCache User not found")
		}

		*u = *resp.Users[0]

		return nil
	}

}

func testAccCheckAWSElasticacheUserAttributes(u *elasticache.User, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *u.UserId != rName {
			return fmt.Errorf("Bad ElastiCache User ID: %#v", u.UserId)
		}

		if *u.UserName != rName {
			return fmt.Errorf("Bad ElastiCache User Name: %#v", u.UserName)
		}

		if *u.Engine != "redis" {
			return fmt.Errorf("Bad ElastiCache User Engine: %#v", u.Engine)
		}

		return nil
	}
}

// Basic Resource

func testAccAWSElasticacheUserConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test-basic" {
  user_id              = %[1]q
  user_name            = %[1]q
  access_string        = "on ~* +@all"
  engine               = "redis"
  no_password_required = true
}
`, rName)
}

// Updates for Passwords

func testAccAWSElasticacheUserConfigNoPasswords(operation string, rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test-passwords-%[1]s" {
  user_id       = %[2]q
  user_name     = %[2]q
  access_string = "on ~* +@all"
  engine        = "redis"
}
`, operation, rName)
}

func testAccAWSElasticacheUserConfigPasswords(operation string, rName string, pass1 string, pass2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test-passwords-%[1]s" {
  user_id              = %[2]q
  user_name            = %[2]q
  access_string        = "on ~* +@all"
  engine               = "redis"
  no_password_required = false
  passwords            = [%[3]q, %[4]q]
}
`, operation, rName, pass1, pass2)
}

// Updates for Access String

func testAccAWSElasticacheUserConfigUpdateAccessStringPre(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test-access-string" {
  user_id       = %[1]q
  user_name     = %[1]q
  access_string = "on ~* +@all"
  engine        = "redis"
}
`, rName)
}

func testAccAWSElasticacheUserConfigUpdateAccessStringPost(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test-access-string" {
  user_id       = %[1]q
  user_name     = %[1]q
  access_string = "off ~* +@all"
  engine        = "redis"
}
`, rName)
}
