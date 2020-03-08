package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsWorkspacesIpGroup_basic(t *testing.T) {
	var v workspaces.IpGroup
	ipGroupName := fmt.Sprintf("terraform-acctest-%s", acctest.RandString(10))
	ipGroupNewName := fmt.Sprintf("terraform-acctest-new-%s", acctest.RandString(10))
	ipGroupDescription := fmt.Sprintf("Terraform Acceptance Test %s", strings.Title(acctest.RandString(20)))
	resourceName := "aws_workspaces_ip_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesIpGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWorkspacesIpGroupConfigA(ipGroupName, ipGroupDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", ipGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ipGroupDescription),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Terraform", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.IPGroup", "Home"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsWorkspacesIpGroupConfigB(ipGroupNewName, ipGroupDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesIpGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", ipGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ipGroupDescription),
					resource.TestCheckResourceAttr(resourceName, "rules.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.IPGroup", "Home"),
					resource.TestCheckResourceAttr(resourceName, "tags.Purpose", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsWorkspacesIpGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_workspaces_ip_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn
		resp, err := conn.DescribeIpGroups(&workspaces.DescribeIpGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error Describing Workspaces IP Group: %s", err)
		}

		// Return nil if the IP Group is already destroyed (does not exist)
		if len(resp.Result) == 0 {
			return nil
		}

		if *resp.Result[0].GroupId == rs.Primary.ID {
			return fmt.Errorf("Workspaces IP Group %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWorkspacesIpGroupExists(n string, v *workspaces.IpGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Workpsaces IP Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn
		resp, err := conn.DescribeIpGroups(&workspaces.DescribeIpGroupsInput{
			GroupIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if *resp.Result[0].GroupId == rs.Primary.ID {
			*v = *resp.Result[0]
			return nil
		}

		return fmt.Errorf("Workspaces IP Group (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWorkspacesIpGroupConfigA(name, description string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test" {
  name        = "%s"
  description = "%s"

  rules {
    source = "10.0.0.0/16"
  }

  rules {
    source      = "10.0.0.1/16"
    description = "Home"
  }

  tags = {
    Name = "test"
    Terraform = true
    IPGroup = "Home"
  }
}
`, name, description)
}

func testAccAwsWorkspacesIpGroupConfigB(name, description string) string {
	return fmt.Sprintf(`
resource "aws_workspaces_ip_group" "test" {
  name        = "%s"
  description = "%s"

  rules {
    source      = "10.0.0.1/16"
    description = "Home"
  }

  tags = {
    Purpose   = "test"
    IPGroup = "Home"
  }
}
`, name, description)
}
