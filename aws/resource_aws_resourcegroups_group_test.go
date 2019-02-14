package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSResourceGroup_basic(t *testing.T) {
	resourceName := "aws_resourcegroups_group.test"
	n := fmt.Sprintf("test-group-%d", acctest.RandInt())

	desc1 := "Hello World"
	desc2 := "Foo Bar"

	query1 := `{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Stage",
      "Values": ["Test"]
    }
  ]
}
`

	query2 := `{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Hello",
      "Values": ["World"]
    }
  ]
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSResourceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSResourceGroupConfig_basic(n, desc1, query1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSResourceGroupExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", n),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
					resource.TestCheckResourceAttr(resourceName, "resource_query.0.query", query1+"\n"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSResourceGroupConfig_basic(n, desc2, query2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", desc2),
					resource.TestCheckResourceAttr(resourceName, "resource_query.0.query", query2+"\n"),
				),
			},
		},
	})
}

func testAccCheckAWSResourceGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource group name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).resourcegroupsconn

		_, err := conn.GetGroup(&resourcegroups.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSResourceGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_rule_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).resourcegroupsconn
		resp, err := conn.GetGroup(&resourcegroups.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if *resp.Group.Name == rs.Primary.ID {
				return fmt.Errorf("Resource Group %s still exists", rs.Primary.ID)
			}
		}

		if isAWSErr(err, resourcegroups.ErrCodeNotFoundException, "") {

			return nil
		}

		return err
	}

	return nil
}

func testAccAWSResourceGroupConfig_basic(rName string, desc string, query string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name        = "%s"
  description = "%s"

  resource_query {
    query = <<JSON
%s
JSON
  }
}
`, rName, desc, query)
}
