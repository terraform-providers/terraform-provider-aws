package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
)

func TestAccAWSDocDBSubnetGroup_basic(t *testing.T) {
	var v docdb.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      "aws_docdb_subnet_group.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDocDBSubnetGroup_disappears(t *testing.T) {
	var v docdb.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					testAccCheckAWSDocDBSubnetGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDocDBSubnetGroup_namePrefix(t *testing.T) {
	var v docdb.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBSubnetGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBSubnetGroupExists(
						"aws_docdb_subnet_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_subnet_group.test", "name", regexp.MustCompile("^tf_test-")),
				),
			},
			{
				ResourceName:            "aws_docdb_subnet_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSDocDBSubnetGroup_generatedName(t *testing.T) {
	var v docdb.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBSubnetGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBSubnetGroupExists(
						"aws_docdb_subnet_group.test", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_subnet_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDocDBSubnetGroup_updateDescription(t *testing.T) {
	var v docdb.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDocDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDocDBSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},

			{
				Config: testAccDocDBSubnetGroupConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDocDBSubnetGroupExists(
						"aws_docdb_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_docdb_subnet_group.foo", "description", "foo description updated"),
				),
			},
			{
				ResourceName:      "aws_docdb_subnet_group.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDocDBSubnetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).docdbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_subnet_group" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeDBSubnetGroups(
			&docdb.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})

		if err == nil {
			if len(resp.DBSubnetGroups) != 0 &&
				aws.StringValue(resp.DBSubnetGroups[0].DBSubnetGroupName) == rs.Primary.ID {
				return fmt.Errorf("DocDB Subnet Group %s still exists", rs.Primary.ID)
			}
		}

		if err != nil {
			if isAWSErr(err, docdb.ErrCodeDBSubnetGroupNotFoundFault, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSDocDBSubnetGroupDisappears(group *docdb.DBSubnetGroup) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).docdbconn

		params := &docdb.DeleteDBSubnetGroupInput{
			DBSubnetGroupName: group.DBSubnetGroupName,
		}

		_, err := conn.DeleteDBSubnetGroup(params)
		if err != nil {
			return err
		}

		return waitForDocDBSubnetGroupDeletion(conn, *group.DBSubnetGroupName)
	}
}

func testAccCheckDocDBSubnetGroupExists(n string, v *docdb.DBSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).docdbconn
		resp, err := conn.DescribeDBSubnetGroups(
			&docdb.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
		if err != nil {
			return err
		}
		if len(resp.DBSubnetGroups) == 0 {
			return fmt.Errorf("DbSubnetGroup not found")
		}

		*v = *resp.DBSubnetGroups[0]

		return nil
	}
}

func testAccDocDBSubnetGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-docdb-subnet-group"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-docdb-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-docdb-subnet-group-2"
	}
}

resource "aws_docdb_subnet_group" "foo" {
	name = "%s"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags = {
		Name = "tf-docdb-subnet-group-test"
	}
}`, rName)
}

func testAccDocDBSubnetGroupConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-docdb-subnet-group"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-docdb-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-docdb-subnet-group-2"
	}
}
	
resource "aws_docdb_subnet_group" "foo" {
	name = "%s"
	description = "foo description updated"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags = {
		Name = "tf-docdb-subnet-group-test"
	}
}`, rName)
}

const testAccDocDBSubnetGroupConfig_namePrefix = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-docdb-subnet-group-name-prefix"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags = {
		Name = "tf-acc-docdb-subnet-group-name-prefix-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags = {
		Name = "tf-acc-docdb-subnet-group-name-prefix-b"
	}
}

resource "aws_docdb_subnet_group" "test" {
	name_prefix = "tf_test-"
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`

const testAccDocDBSubnetGroupConfig_generatedName = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-docdb-subnet-group-generated-name"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags = {
		Name = "tf-acc-docdb-subnet-group-generated-name-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags = {
		Name = "tf-acc-docdb-subnet-group-generated-name-a"
	}
}

resource "aws_docdb_subnet_group" "test" {
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`
