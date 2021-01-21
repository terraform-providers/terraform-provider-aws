package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSOpsworksPermission_basic(t *testing.T) {
	sName := fmt.Sprintf("tf-ops-perm-%d", acctest.RandInt())
	var opsperm opsworks.Permission
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(opsworks.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksPermissionCreate(sName, "true", "true", "iam_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksPermissionExists(
						"aws_opsworks_permission.tf-acc-perm", &opsperm),
					testAccCheckAWSOpsworksCreatePermissionAttributes(&opsperm, true, true, "iam_only"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_ssh", "true",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_sudo", "true",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "level", "iam_only",
					),
				),
			},
			{
				Config: testAccAwsOpsworksPermissionCreate(sName, "true", "false", "iam_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksPermissionExists(
						"aws_opsworks_permission.tf-acc-perm", &opsperm),
					testAccCheckAWSOpsworksCreatePermissionAttributes(&opsperm, true, false, "iam_only"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_ssh", "true",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_sudo", "false",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "level", "iam_only",
					),
				),
			},
			{
				Config: testAccAwsOpsworksPermissionCreate(sName, "false", "false", "deny"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksPermissionExists(
						"aws_opsworks_permission.tf-acc-perm", &opsperm),
					testAccCheckAWSOpsworksCreatePermissionAttributes(&opsperm, false, false, "deny"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_ssh", "false",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_sudo", "false",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "level", "deny",
					),
				),
			},
			{
				Config: testAccAwsOpsworksPermissionCreate(sName, "false", "false", "show"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksPermissionExists(
						"aws_opsworks_permission.tf-acc-perm", &opsperm),
					testAccCheckAWSOpsworksCreatePermissionAttributes(&opsperm, false, false, "show"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_ssh", "false",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "allow_sudo", "false",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_permission.tf-acc-perm", "level", "show",
					),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4804
func TestAccAWSOpsworksPermission_Self(t *testing.T) {
	var opsperm opsworks.Permission
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(opsworks.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: nil, // Cannot delete own OpsWorks Permission
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksPermissionSelf(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksPermissionExists(resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "true"),
				),
			},
			{
				Config: testAccAwsOpsworksPermissionSelf(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksPermissionExists(resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSOpsworksPermissionExists(
	n string, opsperm *opsworks.Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).opsworksconn

		params := &opsworks.DescribePermissionsInput{
			StackId:    aws.String(rs.Primary.Attributes["stack_id"]),
			IamUserArn: aws.String(rs.Primary.Attributes["user_arn"]),
		}
		resp, err := conn.DescribePermissions(params)

		if err != nil {
			return err
		}

		if v := len(resp.Permissions); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opsperm = *resp.Permissions[0]

		return nil
	}
}

func testAccCheckAWSOpsworksCreatePermissionAttributes(
	opsperm *opsworks.Permission, allowSsh bool, allowSudo bool, level string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsperm.AllowSsh != allowSsh {
			return fmt.Errorf("Unnexpected allowSsh: %t", *opsperm.AllowSsh)
		}

		if *opsperm.AllowSudo != allowSudo {
			return fmt.Errorf("Unnexpected allowSudo: %t", *opsperm.AllowSudo)
		}

		if *opsperm.Level != level {
			return fmt.Errorf("Unnexpected level: %s", *opsperm.Level)
		}

		return nil
	}
}

func testAccCheckAwsOpsworksPermissionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*AWSClient).opsworksconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_permission" {
			continue
		}

		req := &opsworks.DescribePermissionsInput{
			IamUserArn: aws.String(rs.Primary.Attributes["user_arn"]),
		}

		resp, err := client.DescribePermissions(req)
		if err == nil {
			if len(resp.Permissions) > 0 {
				return fmt.Errorf("OpsWorks Permissions still exist.")
			}
		}

		if awserr, ok := err.(awserr.Error); ok {
			if awserr.Code() != "ResourceNotFoundException" {
				return err
			}
		}
	}
	return nil
}

func testAccAwsOpsworksPermissionBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = "tf-acc-test-opsworks-permission"
  }
}

resource "aws_subnet" "test" {
  cidr_block = aws_vpc.test.cidr_block
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-opsworks-permissions"
  }
}

resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = data.aws_region.current.name
  vpc_id                       = aws_vpc.test.id
  default_subnet_id            = aws_subnet.test.id
  service_role_arn             = aws_iam_role.service.arn
  default_instance_profile_arn = aws_iam_instance_profile.test.arn
  default_os                   = "Amazon Linux 2016.09"
  default_root_device_type     = "ebs"

  custom_json = <<EOF
{
  "key": "value"
}
EOF

  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}

resource "aws_iam_role" "service" {
  name = "%[1]s-service"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "service" {
  name = %[1]q
  role = aws_iam_role.service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.instance.name
}
`, rName)
}

func testAccAwsOpsworksPermissionCreate(name, ssh, sudo, level string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_permission" "tf-acc-perm" {
  stack_id = aws_opsworks_stack.tf-acc.id

  allow_ssh  = %s
  allow_sudo = %s
  user_arn   = aws_opsworks_user_profile.user.user_arn
  level      = "%s"
}

resource "aws_opsworks_user_profile" "user" {
  user_arn     = aws_iam_user.user.arn
  ssh_username = aws_iam_user.user.name
}

resource "aws_iam_user" "user" {
  name = "%s"
  path = "/"
}
%s
`, ssh, sudo, level, name, testAccAwsOpsworksStackConfigVpcCreate(name))
}

func testAccAwsOpsworksPermissionSelf(rName string, allowSsh bool, allowSudo bool) string {
	return testAccAwsOpsworksPermissionBase(rName) + fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_opsworks_permission" "test" {
  allow_ssh  = %[1]t
  allow_sudo = %[2]t
  stack_id   = aws_opsworks_stack.test.id
  user_arn   = data.aws_caller_identity.current.arn
}
`, allowSsh, allowSudo)
}
