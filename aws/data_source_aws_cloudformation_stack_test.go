package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudFormationStack_dataSource_basic(t *testing.T) {
	stackName := acctest.RandomWithPrefix("tf-acc-ds-basic")
	resourceName := "data.aws_cloudformation_stack.network"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudFormationStackDataSourceConfig_basic(stackName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "outputs.%", "1"),
					resource.TestMatchResourceAttr(resourceName, "outputs.VPCId", regexp.MustCompile("^vpc-[a-z0-9]+")),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.CIDR", "10.10.10.0/24"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", "6"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Form the Cloud"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "meh"),
				),
			},
		},
	})
}

func testAccCheckAwsCloudFormationStackDataSourceConfig_basic(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "cfs" {
  name = "%s"

  parameters = {
    CIDR = "10.10.10.0/24"
  }

  timeout_in_minutes = 6

  template_body = <<STACK
{
  "Parameters": {
    "CIDR": {
      "Type": "String"
    }
  },
  "Resources": {
    "myvpc": {
      "Type": "AWS::EC2::VPC",
      "Properties": {
        "CidrBlock": {
          "Ref": "CIDR"
        },
        "Tags": [
          {
            "Key": "Name",
            "Value": "Primary_CF_VPC"
          }
        ]
      }
    }
  },
  "Outputs": {
    "VPCId": {
      "Value": {
        "Ref": "myvpc"
      },
      "Description": "VPC ID"
    }
  }
}
STACK

  tags = {
    Name   = "Form the Cloud"
    Second = "meh"
  }
}

data "aws_cloudformation_stack" "network" {
  name = aws_cloudformation_stack.cfs.name
}
`, stackName)
}

func TestAccAWSCloudFormationStack_dataSource_yaml(t *testing.T) {
	stackName := acctest.RandomWithPrefix("tf-acc-ds-yaml")
	resourceName := "data.aws_cloudformation_stack.yaml"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsCloudFormationStackDataSourceConfig_yaml(stackName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "outputs.%", "1"),
					resource.TestMatchResourceAttr(resourceName, "outputs.VPCId", regexp.MustCompile("^vpc-[a-z0-9]+")),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.CIDR", "10.10.10.0/24"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", "6"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Form the Cloud"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "meh"),
				),
			},
		},
	})
}

func testAccCheckAwsCloudFormationStackDataSourceConfig_yaml(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "yaml" {
  name = "%s"

  parameters = {
    CIDR = "10.10.10.0/24"
  }

  timeout_in_minutes = 6

  template_body = <<STACK
Parameters:
  CIDR:
    Type: String

Resources:
  myvpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: !Ref CIDR
      Tags:
        -
          Key: Name
          Value: Primary_CF_VPC

Outputs:
  VPCId:
    Value: !Ref myvpc
    Description: VPC ID
STACK

  tags = {
    Name   = "Form the Cloud"
    Second = "meh"
  }
}

data "aws_cloudformation_stack" "yaml" {
  name = aws_cloudformation_stack.yaml.name
}
`, stackName)
}
