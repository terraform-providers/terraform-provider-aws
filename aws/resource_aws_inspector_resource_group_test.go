package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSInspectorResourceGroup_basic(t *testing.T) {
	var v1, v2 inspector.ResourceGroup
	resourceName := "aws_inspector_resource_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInspectorResourceGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorResourceGroupExists(resourceName, &v1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "inspector", regexp.MustCompile(`resourcegroup/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "foo"),
				),
			},
			{
				Config: testAccCheckAWSInspectorResourceGroupModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorResourceGroupExists(resourceName, &v2),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "inspector", regexp.MustCompile(`resourcegroup/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "bar"),
					testAccCheckAWSInspectorResourceGroupRecreated(&v1, &v2),
				),
			},
		},
	})
}

func testAccCheckAWSInspectorResourceGroupExists(name string, rg *inspector.ResourceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).inspectorconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		output, err := conn.DescribeResourceGroups(&inspector.DescribeResourceGroupsInput{
			ResourceGroupArns: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}
		if len(output.ResourceGroups) == 0 {
			return fmt.Errorf("No matching Inspector resource groups")
		}

		*rg = *output.ResourceGroups[0]

		return nil
	}
}

func testAccCheckAWSInspectorResourceGroupRecreated(v1, v2 *inspector.ResourceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v2.CreatedAt.Equal(*v1.CreatedAt) {
			return fmt.Errorf("Inspector resource group not recreated when changing tags")
		}

		return nil
	}
}

var testAccAWSInspectorResourceGroup = `
resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = "foo"
  }
}
`

var testAccCheckAWSInspectorResourceGroupModified = `
resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = "bar"
  }
}
`
