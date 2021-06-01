package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSResourceGroup_basic(t *testing.T) {
	var v resourcegroups.Group
	resourceName := "aws_resourcegroups_group.test"
	n := fmt.Sprintf("test-group-%d", acctest.RandInt())

	desc1 := "Hello World"
	desc2 := "Foo Bar"

	query2 := `{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Hello",
      "Values": [
        "World"
      ]
    }
  ]
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSResourceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSResourceGroupConfig_basic(n, desc1, testAccAWSResourceGroupConfigQuery),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", n),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
					resource.TestCheckResourceAttr(resourceName, "resource_query.0.query", testAccAWSResourceGroupConfigQuery+"\n"),
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

func TestAccAWSResourceGroup_tags(t *testing.T) {
	var v resourcegroups.Group
	resourceName := "aws_resourcegroups_group.test"
	n := fmt.Sprintf("test-group-%d", acctest.RandInt())
	desc1 := "Hello World"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSResourceGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSResourceGroupConfigTags1(n, desc1, testAccAWSResourceGroupConfigQuery, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSResourceGroupConfigTags2(n, desc1, testAccAWSResourceGroupConfigQuery, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSResourceGroupConfigTags1(n, desc1, testAccAWSResourceGroupConfigQuery, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSResourceGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSResourceGroupExists(n string, v *resourcegroups.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource group name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).resourcegroupsconn

		resp, err := conn.GetGroup(&resourcegroups.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Group.Name == rs.Primary.ID {
			*v = *resp.Group
			return nil
		}

		return fmt.Errorf("Resource Group (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSResourceGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_resourcegroups_group" {
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

const testAccAWSResourceGroupConfigQuery = `{
  "ResourceTypeFilters": [
    "AWS::EC2::Instance"
  ],
  "TagFilters": [
    {
      "Key": "Stage",
      "Values": [
        "Test"
      ]
    }
  ]
}`

func testAccAWSResourceGroupConfig_basic(rName, desc, query string) string {
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

func testAccAWSResourceGroupConfigTags1(rName, desc, query, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name        = "%s"
  description = "%s"

  resource_query {
    query = <<JSON
%s
JSON

  }

  tags = {
    %q = %q
  }
}
`, rName, desc, query, tag1Key, tag1Value)
}

func testAccAWSResourceGroupConfigTags2(rName, desc, query, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_resourcegroups_group" "test" {
  name        = "%s"
  description = "%s"

  resource_query {
    query = <<JSON
%s
JSON

  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, desc, query, tag1Key, tag1Value, tag2Key, tag2Value)
}
