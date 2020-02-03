package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCloud9EnvironmentMembership_basic(t *testing.T) {
	var conf cloud9.EnvironmentMember

	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_cloud9_environment_membership.test"
	userResourceName := "aws_iam_user.test"
	envResourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloud9(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentMembershipConfig(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentMembershipExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "permissions", "read-only"),
					resource.TestCheckResourceAttrPair(resourceName, "user_arn", userResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", envResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloud9EnvironmentMembershipConfig(rName, "read-write"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentMembershipExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "permissions", "read-write"),
					resource.TestCheckResourceAttrPair(resourceName, "user_arn", userResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "environment_id", envResourceName, "id"),
				),
			},
		},
	})
}

func TestAccAWSCloud9EnvironmentMembership_disappears(t *testing.T) {
	var conf cloud9.EnvironmentMember

	rName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_cloud9_environment_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloud9(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentMembershipConfig(rName, "read-only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentMembershipExists(resourceName, &conf),
					testAccCheckAWSCloud9EnvironmentMembershipDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloud9EnvironmentMembershipExists(n string, res *cloud9.EnvironmentMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cloud9 Environment Membership ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloud9conn

		out, err := conn.DescribeEnvironmentMemberships(&cloud9.DescribeEnvironmentMembershipsInput{
			EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
			UserArn:       aws.String(rs.Primary.Attributes["user_arn"]),
		})
		if err != nil {
			if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
				return fmt.Errorf("Cloud9 Environment Membership (%q) not found", rs.Primary.ID)
			}
			return err
		}
		if len(out.Memberships) == 0 {
			return fmt.Errorf("Cloud9 Environment Membership (%q) not found", rs.Primary.ID)
		}
		env := out.Memberships[0]

		*res = *env

		return nil
	}
}

func testAccCheckAWSCloud9EnvironmentMembershipDisappears(res *cloud9.EnvironmentMember) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloud9conn

		_, err := conn.DeleteEnvironmentMembership(&cloud9.DeleteEnvironmentMembershipInput{
			EnvironmentId: res.EnvironmentId,
			UserArn:       res.UserArn,
		})

		return err
	}
}

func testAccCheckAWSCloud9EnvironmentMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloud9conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloud9_environment_membership" {
			continue
		}

		_, err := conn.DeleteEnvironmentMembership(&cloud9.DeleteEnvironmentMembershipInput{
			EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
			UserArn:       aws.String(rs.Primary.Attributes["user_arn"]),
		})
		if err != nil {
			if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, "AccessDeniedException", "is not authorized to access this resource") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Cloud9 Environment Membership %q still exists.", rs.Primary.ID)
	}
	return nil
}

func testAccAWSCloud9EnvironmentMembershipConfig(name, permission string) string {
	return testAccAWSCloud9EnvironmentEc2Config(name) + fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_cloud9_environment_membership" "test" {
  environment_id = "${aws_cloud9_environment_ec2.test.id}"
  permissions    = %[2]q
  user_arn       = "${aws_iam_user.test.arn}"
}
`, name, permission)
}
