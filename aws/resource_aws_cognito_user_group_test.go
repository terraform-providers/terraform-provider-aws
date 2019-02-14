package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCognitoUserGroup_basic(t *testing.T) {
	poolName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	groupName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	updatedGroupName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserGroupConfig_basic(poolName, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserGroupExists("aws_cognito_user_group.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "name", groupName),
				),
			},
			{
				Config: testAccAWSCognitoUserGroupConfig_basic(poolName, updatedGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserGroupExists("aws_cognito_user_group.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "name", updatedGroupName),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserGroup_complex(t *testing.T) {
	poolName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	groupName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	updatedGroupName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserGroupConfig_complex(poolName, groupName, "This is the user group description", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserGroupExists("aws_cognito_user_group.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "name", groupName),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "description", "This is the user group description"),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "precedence", "1"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_group.main", "role_arn"),
				),
			},
			{
				Config: testAccAWSCognitoUserGroupConfig_complex(poolName, updatedGroupName, "This is the updated user group description", 42),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserGroupExists("aws_cognito_user_group.main"),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "name", updatedGroupName),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "description", "This is the updated user group description"),
					resource.TestCheckResourceAttr("aws_cognito_user_group.main", "precedence", "42"),
					resource.TestCheckResourceAttrSet("aws_cognito_user_group.main", "role_arn"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserGroup_RoleArn(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_cognito_user_group.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserGroupConfig_RoleArn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserGroupExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
				),
			},
			{
				Config: testAccAWSCognitoUserGroupConfig_RoleArn_Updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCognitoUserGroupExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "role_arn"),
				),
			},
		},
	})
}

func TestAccAWSCognitoUserGroup_importBasic(t *testing.T) {
	resourceName := "aws_cognito_user_group.main"
	poolName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	groupName := fmt.Sprintf("tf-acc-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCognitoUserGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserGroupConfig_basic(poolName, groupName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSCognitoUserGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		id := rs.Primary.ID
		name := rs.Primary.Attributes["name"]
		userPoolId := rs.Primary.Attributes["user_pool_id"]

		if name == "" {
			return errors.New("No Cognito User Group Name set")
		}

		if userPoolId == "" {
			return errors.New("No Cognito User Pool Id set")
		}

		if id != fmt.Sprintf("%s/%s", userPoolId, name) {
			return fmt.Errorf(fmt.Sprintf("ID should be user_pool_id/name. ID was %s. name was %s, user_pool_id was %s", id, name, userPoolId))
		}

		conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

		params := &cognitoidentityprovider.GetGroupInput{
			GroupName:  aws.String(rs.Primary.Attributes["name"]),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.GetGroup(params)
		return err
	}
}

func testAccCheckAWSCognitoUserGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cognitoidpconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cognito_user_group" {
			continue
		}

		params := &cognitoidentityprovider.GetGroupInput{
			GroupName:  aws.String(rs.Primary.ID),
			UserPoolId: aws.String(rs.Primary.Attributes["user_pool_id"]),
		}

		_, err := conn.GetGroup(params)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccAWSCognitoUserGroupConfig_basic(poolName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
  name = "%s"
}

resource "aws_cognito_user_group" "main" {
  name         = "%s"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
}
`, poolName, groupName)
}

func testAccAWSCognitoUserGroupConfig_complex(poolName, groupName, groupDescription string, precedence int) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
  name = "%s"
}

resource "aws_iam_role" "group_role" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "cognito-identity.amazonaws.com:aud": "us-east-1:12345678-dead-beef-cafe-123456790ab"
        },
        "ForAnyValue:StringLike": {
          "cognito-identity.amazonaws.com:amr": "authenticated"
        }
      }
    }
  ]
}
EOF
}

resource "aws_cognito_user_group" "main" {
  name         = "%s"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
  description  = "%s"
  precedence   = %v
  role_arn     = "${aws_iam_role.group_role.arn}"
}
`, poolName, groupName, groupName, groupDescription, precedence)
}

func testAccAWSCognitoUserGroupConfig_RoleArn(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
  name = "%[1]s"
}

resource "aws_iam_role" "group_role" {
  name = "%[1]s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity"
    }
  ]
}
EOF
}

resource "aws_cognito_user_group" "main" {
  name         = "%[1]s"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
  role_arn     = "${aws_iam_role.group_role.arn}"
}
`, rName)
}

func testAccAWSCognitoUserGroupConfig_RoleArn_Updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
  name = "%[1]s"
}

resource "aws_iam_role" "group_role_updated" {
  name = "%[1]s-updated"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Federated": "cognito-identity.amazonaws.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity"
    }
  ]
}
EOF
}

resource "aws_cognito_user_group" "main" {
  name         = "%[1]s"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
  role_arn     = "${aws_iam_role.group_role_updated.arn}"
}
`, rName)
}
