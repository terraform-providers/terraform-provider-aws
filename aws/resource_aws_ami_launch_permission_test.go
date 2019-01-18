package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAMILaunchPermission_Basic(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMILaunchPermissionExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSAMILaunchPermission_Disappears_LaunchPermission(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMILaunchPermissionExists(resourceName),
					testAccCheckAWSAMILaunchPermissionDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Bug reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6222
// Images with <group>all</group> will not have <userId> and can cause a panic
func TestAccAWSAMILaunchPermission_Disappears_LaunchPermission_Public(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMILaunchPermissionExists(resourceName),
					testAccCheckAWSAMILaunchPermissionAddPublic(resourceName),
					testAccCheckAWSAMILaunchPermissionDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAMILaunchPermission_Disappears_AMI(t *testing.T) {
	imageID := ""
	resourceName := "aws_ami_launch_permission.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMILaunchPermissionExists(resourceName),
				),
			},
			// Here we delete the AMI to verify the follow-on refresh after this step
			// should not error.
			{
				Config: testAccAWSAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceGetAttr("aws_ami_copy.test", "id", &imageID),
					testAccAWSAMIDisappears(&imageID),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testCheckResourceGetAttr(name, key string, value *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s", name)
		}

		*value = is.Attributes[key]
		return nil
	}
}

func testAccCheckAWSAMILaunchPermissionExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		accountID := rs.Primary.Attributes["account_id"]
		imageID := rs.Primary.Attributes["image_id"]

		if has, err := hasLaunchPermission(conn, imageID, accountID); err != nil {
			return err
		} else if !has {
			return fmt.Errorf("launch permission does not exist for '%s' on '%s'", accountID, imageID)
		}
		return nil
	}
}

func testAccCheckAWSAMILaunchPermissionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami_launch_permission" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		accountID := rs.Primary.Attributes["account_id"]
		imageID := rs.Primary.Attributes["image_id"]

		if has, err := hasLaunchPermission(conn, imageID, accountID); err != nil {
			return err
		} else if has {
			return fmt.Errorf("launch permission still exists for '%s' on '%s'", accountID, imageID)
		}
	}

	return nil
}

func testAccCheckAWSAMILaunchPermissionAddPublic(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		imageID := rs.Primary.Attributes["image_id"]

		input := &ec2.ModifyImageAttributeInput{
			ImageId:   aws.String(imageID),
			Attribute: aws.String("launchPermission"),
			LaunchPermission: &ec2.LaunchPermissionModifications{
				Add: []*ec2.LaunchPermission{
					{Group: aws.String("all")},
				},
			},
		}

		_, err := conn.ModifyImageAttribute(input)

		return err
	}
}

func testAccCheckAWSAMILaunchPermissionDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		accountID := rs.Primary.Attributes["account_id"]
		imageID := rs.Primary.Attributes["image_id"]

		input := &ec2.ModifyImageAttributeInput{
			ImageId:   aws.String(imageID),
			Attribute: aws.String("launchPermission"),
			LaunchPermission: &ec2.LaunchPermissionModifications{
				Remove: []*ec2.LaunchPermission{
					{UserId: aws.String(accountID)},
				},
			},
		}

		_, err := conn.ModifyImageAttribute(input)

		return err
	}
}

// testAccAWSAMIDisappears is technically a "test check function" but really it
// exists to perform a side effect of deleting an AMI out from under a resource
// so we can test that Terraform will react properly
func testAccAWSAMIDisappears(imageID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		req := &ec2.DeregisterImageInput{
			ImageId: aws.String(*imageID),
		}

		_, err := conn.DeregisterImage(req)
		if err != nil {
			return err
		}

		if err := resourceAwsAmiWaitForDestroy(AWSAMIDeleteRetryTimeout, *imageID, conn); err != nil {
			return err
		}
		return nil
	}
}

func testAccAWSAMILaunchPermissionConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %q
  name              = %q
  source_ami_id     = "${data.aws_ami.amzn-ami-minimal-hvm.id}"
  source_ami_region = "${data.aws_region.current.name}"
}

resource "aws_ami_launch_permission" "test" {
  account_id = "${data.aws_caller_identity.current.account_id}"
  image_id   = "${aws_ami_copy.test.id}"
}
`, rName, rName)
}
