package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSCloudformationExportDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:                    testAccCheckAwsCloudformationExportConfig(rName),
				PreventPostDestroyRefresh: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cloudformation_export.waiter", "value", "waiter"),
					resource.TestMatchResourceAttr("data.aws_cloudformation_export.vpc", "value",
						regexp.MustCompile("^vpc-[a-z0-9]{8,}$")),
					resource.TestMatchResourceAttr("data.aws_cloudformation_export.vpc", "exporting_stack_id",
						regexp.MustCompile("^arn:aws:cloudformation")),
				),
			},
		},
	})
}

func testAccCheckAwsCloudformationExportConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "cfs" {
  name = "%s1"
  timeout_in_minutes = 6
  template_body = <<STACK
{
  "Resources": {
    "waiter": {
      "Type": "AWS::CloudFormation::WaitConditionHandle",
      "Properties": { }
    }
  },
  "Outputs": {
    "waiter": {
      "Value": "waiter" ,
      "Description": "VPC ID",
      "Export": {
        "Name": "waiter" 
      }
    }
  }
}
STACK
  tags = {
    TestExport = "waiter"
    Second = "meh"
  }
}
resource "aws_cloudformation_stack" "yaml" {
  name = "%s2"
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
  MyVpcId:
    Value: !Ref myvpc
    Description: VPC ID
    Export:
      Name: MyVpcId
STACK
  tags = {
    TestExport = "MyVpcId"
    Second = "meh"
  }
}
data "aws_cloudformation_export" "vpc" {
	name = "${aws_cloudformation_stack.yaml.tags["TestExport"]}"
}
data "aws_cloudformation_export" "waiter" {
	name = "${aws_cloudformation_stack.cfs.tags["TestExport"]}"
}
`, rName, rName)
}
