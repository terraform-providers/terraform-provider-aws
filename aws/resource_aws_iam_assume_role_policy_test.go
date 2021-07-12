package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIamAssumeRolePolicy_basic(t *testing.T) {
	rName := acctest.RandString(10)
	roleResourceName := "aws_iam_role.role"
	resourceName := "aws_iam_assume_role_policy.policy"
	oldService := "ec2.amazonaws.com"
	newService := "s3.amazonaws.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIamAssumeRolePolicy_roleConfig(rName, oldService),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAssumeRolePolicyServiceMatches(roleResourceName, oldService),
				),
			},
			{
				Config: testAccAWSIamAssumeRolePolicyConfig(rName, oldService, newService),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAssumeRolePolicyServiceMatches(resourceName, newService),
				),
			},
		},
	})
}

func testAccAWSIamAssumeRolePolicy_roleConfig(rName string, service string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "%s"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName, service)
}

func testAccAWSIamAssumeRolePolicyConfig(rName string, oldService string, newService string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name = "%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "%s"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  lifecycle {
    ignore_changes = [
	  # Ignore changes to assume role policy because we are about
	  # to change it via the resource under test
      assume_role_policy,
    ]
  }
}

resource "aws_iam_assume_role_policy" "policy" {
  role_name = aws_iam_role.role.name
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "%s"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName, oldService, newService)
}

func testAccCheckAWSAssumeRolePolicyServiceMatches(resourceName string, service string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Role name is set")
		}

		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		resp, err := iamconn.GetRole(&iam.GetRoleInput{
			RoleName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		actualAssumeRolePolicy := *resp.Role.AssumeRolePolicyDocument

		if !strings.Contains(actualAssumeRolePolicy, service) {
			return fmt.Errorf("AssumeRolePolicy: '%q' did not contain expected service '%q'.", actualAssumeRolePolicy, service)
		}

		return nil
	}
}
