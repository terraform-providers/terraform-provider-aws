package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDaxSubnetGroup_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dax_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDaxSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDaxSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxSubnetGroupExists("aws_dax_subnet_group.test"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "subnet_ids.#", "2"),
					resource.TestCheckResourceAttrSet("aws_dax_subnet_group.test", "vpc_id"),
				),
			},
			{
				Config: testAccDaxSubnetGroupConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDaxSubnetGroupExists("aws_dax_subnet_group.test"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "description", "update"),
					resource.TestCheckResourceAttr("aws_dax_subnet_group.test", "subnet_ids.#", "3"),
					resource.TestCheckResourceAttrSet("aws_dax_subnet_group.test", "vpc_id"),
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

func testAccCheckAwsDaxSubnetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).daxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dax_subnet_group" {
			continue
		}

		_, err := conn.DescribeSubnetGroups(&dax.DescribeSubnetGroupsInput{
			SubnetGroupNames: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			if isAWSErr(err, dax.ErrCodeSubnetGroupNotFoundFault, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsDaxSubnetGroupExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).daxconn

		_, err := conn.DescribeSubnetGroups(&dax.DescribeSubnetGroupsInput{
			SubnetGroupNames: []*string{aws.String(rs.Primary.ID)},
		})

		return err
	}
}

func testAccDaxSubnetGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  cidr_block = "10.0.1.0/24"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test2" {
  cidr_block = "10.0.2.0/24"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_dax_subnet_group" "test" {
  name = "%s"
  subnet_ids = [
    "${aws_subnet.test1.id}",
    "${aws_subnet.test2.id}",
  ]
}
`, rName)
}

func testAccDaxSubnetGroupConfig_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  cidr_block = "10.0.1.0/24"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test2" {
  cidr_block = "10.0.2.0/24"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test3" {
  cidr_block = "10.0.3.0/24"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_dax_subnet_group" "test" {
  name = "%s"
  description = "update"
  subnet_ids = [
    "${aws_subnet.test1.id}",
    "${aws_subnet.test2.id}",
    "${aws_subnet.test3.id}",
  ]
}
`, rName)
}
