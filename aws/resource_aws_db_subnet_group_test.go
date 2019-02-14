package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
)

func TestAccAWSDBSubnetGroup_importBasic(t *testing.T) {
	resourceName := "aws_db_subnet_group.foo"

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
		},
	})
}

func TestAccAWSDBSubnetGroup_basic(t *testing.T) {
	var v rds.DBSubnetGroup

	testCheck := func(*terraform.State) error {
		return nil
	}

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_db_subnet_group.foo", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_db_subnet_group.foo", "description", "Managed by Terraform"),
					resource.TestMatchResourceAttr(
						"aws_db_subnet_group.foo", "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:rds:[^:]+:\\d{12}:subgrp:%s", rName))),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSDBSubnetGroup_namePrefix(t *testing.T) {
	var v rds.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_db_subnet_group.test", "name", regexp.MustCompile("^tf_test-")),
				),
			},
		},
	})
}

func TestAccAWSDBSubnetGroup_generatedName(t *testing.T) {
	var v rds.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.test", &v),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/2603 and
// https://github.com/hashicorp/terraform/issues/2664
func TestAccAWSDBSubnetGroup_withUndocumentedCharacters(t *testing.T) {
	var v rds.DBSubnetGroup

	testCheck := func(*terraform.State) error {
		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.underscores", &v),
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.periods", &v),
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.spaces", &v),
					testCheck,
				),
			},
		},
	})
}

func TestAccAWSDBSubnetGroup_updateDescription(t *testing.T) {
	var v rds.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_db_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},

			{
				Config: testAccDBSubnetGroupConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(
						"aws_db_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_db_subnet_group.foo", "description", "foo description updated"),
				),
			},
		},
	})
}

func testAccCheckDBSubnetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_subnet_group" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeDBSubnetGroups(
			&rds.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
		if err == nil {
			if len(resp.DBSubnetGroups) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		rdserr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if rdserr.Code() != "DBSubnetGroupNotFoundFault" {
			return err
		}
	}

	return nil
}

func testAccCheckDBSubnetGroupExists(n string, v *rds.DBSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn
		resp, err := conn.DescribeDBSubnetGroups(
			&rds.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
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

func testAccDBSubnetGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-db-subnet-group"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-db-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-db-subnet-group-2"
	}
}

resource "aws_db_subnet_group" "foo" {
	name = "%s"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags = {
		Name = "tf-dbsubnet-group-test"
	}
}`, rName)
}

func testAccDBSubnetGroupConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-db-subnet-group-updated-description"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-db-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-db-subnet-group-2"
	}
}

resource "aws_db_subnet_group" "foo" {
	name = "%s"
	description = "foo description updated"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags = {
		Name = "tf-dbsubnet-group-test"
	}
}`, rName)
}

const testAccDBSubnetGroupConfig_namePrefix = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-db-subnet-group-name-prefix"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags = {
		Name = "tf-acc-db-subnet-group-name-prefix-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags = {
		Name = "tf-acc-db-subnet-group-name-prefix-b"
	}
}

resource "aws_db_subnet_group" "test" {
	name_prefix = "tf_test-"
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`

const testAccDBSubnetGroupConfig_generatedName = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-db-subnet-group-generated-name"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags = {
		Name = "tf-acc-db-subnet-group-generated-name-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags = {
		Name = "tf-acc-db-subnet-group-generated-name-a"
	}
}

resource "aws_db_subnet_group" "test" {
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`

const testAccDBSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces = `
resource "aws_vpc" "main" {
    cidr_block = "192.168.0.0/16"
	tags = {
			Name = "terraform-testacc-db-subnet-group-w-underscores-etc"
		}
}

resource "aws_subnet" "frontend" {
    vpc_id = "${aws_vpc.main.id}"
    availability_zone = "us-west-2b"
    cidr_block = "192.168.1.0/24"
  tags = {
        Name = "tf-acc-db-subnet-group-w-underscores-etc-front"
    }
}

resource "aws_subnet" "backend" {
    vpc_id = "${aws_vpc.main.id}"
    availability_zone = "us-west-2c"
    cidr_block = "192.168.2.0/24"
  tags = {
        Name = "tf-acc-db-subnet-group-w-underscores-etc-back"
    }
}

resource "aws_db_subnet_group" "underscores" {
    name = "with_underscores"
    description = "Our main group of subnets"
    subnet_ids = ["${aws_subnet.frontend.id}", "${aws_subnet.backend.id}"]
}

resource "aws_db_subnet_group" "periods" {
    name = "with.periods"
    description = "Our main group of subnets"
    subnet_ids = ["${aws_subnet.frontend.id}", "${aws_subnet.backend.id}"]
}

resource "aws_db_subnet_group" "spaces" {
    name = "with spaces"
    description = "Our main group of subnets"
    subnet_ids = ["${aws_subnet.frontend.id}", "${aws_subnet.backend.id}"]
}
`
