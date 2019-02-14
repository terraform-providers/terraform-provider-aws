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
	"github.com/aws/aws-sdk-go/service/neptune"
)

func TestAccAWSNeptuneSubnetGroup_basic(t *testing.T) {
	var v neptune.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNeptuneSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},
			{
				ResourceName:      "aws_neptune_subnet_group.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneSubnetGroup_namePrefix(t *testing.T) {
	var v neptune.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNeptuneSubnetGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_subnet_group.test", "name", regexp.MustCompile("^tf_test-")),
				),
			},
			{
				ResourceName:            "aws_neptune_subnet_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSNeptuneSubnetGroup_generatedName(t *testing.T) {
	var v neptune.DBSubnetGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNeptuneSubnetGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.test", &v),
				),
			},
			{
				ResourceName:      "aws_neptune_subnet_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneSubnetGroup_updateDescription(t *testing.T) {
	var v neptune.DBSubnetGroup

	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNeptuneSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNeptuneSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "description", "Managed by Terraform"),
				),
			},

			{
				Config: testAccNeptuneSubnetGroupConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNeptuneSubnetGroupExists(
						"aws_neptune_subnet_group.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_neptune_subnet_group.foo", "description", "foo description updated"),
				),
			},
			{
				ResourceName:      "aws_neptune_subnet_group.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckNeptuneSubnetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_subnet_group" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeDBSubnetGroups(
			&neptune.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
		if err == nil {
			if len(resp.DBSubnetGroups) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		neptuneerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if neptuneerr.Code() != "DBSubnetGroupNotFoundFault" {
			return err
		}
	}

	return nil
}

func testAccCheckNeptuneSubnetGroupExists(n string, v *neptune.DBSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn
		resp, err := conn.DescribeDBSubnetGroups(
			&neptune.DescribeDBSubnetGroupsInput{DBSubnetGroupName: aws.String(rs.Primary.ID)})
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

func testAccNeptuneSubnetGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-neptune-subnet-group"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-neptune-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-neptune-subnet-group-2"
	}
}

resource "aws_neptune_subnet_group" "foo" {
	name = "%s"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags = {
		Name = "tf-neptunesubnet-group-test"
	}
}`, rName)
}

func testAccNeptuneSubnetGroupConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-neptune-subnet-group-updated-description"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-neptune-subnet-group-1"
	}
}

resource "aws_subnet" "bar" {
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-neptune-subnet-group-2"
	}
}

resource "aws_neptune_subnet_group" "foo" {
	name = "%s"
	description = "foo description updated"
	subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
	tags = {
		Name = "tf-neptunesubnet-group-test"
	}
}`, rName)
}

const testAccNeptuneSubnetGroupConfig_namePrefix = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-neptune-subnet-group-name-prefix"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags = {
		Name = "tf-acc-neptune-subnet-group-name-prefix-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags = {
		Name = "tf-acc-neptune-subnet-group-name-prefix-b"
	}
}

resource "aws_neptune_subnet_group" "test" {
	name_prefix = "tf_test-"
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`

const testAccNeptuneSubnetGroupConfig_generatedName = `
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-neptune-subnet-group-generated-name"
	}
}

resource "aws_subnet" "a" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.1.0/24"
	availability_zone = "us-west-2a"
	tags = {
		Name = "tf-acc-neptune-subnet-group-generated-name-a"
	}
}

resource "aws_subnet" "b" {
	vpc_id = "${aws_vpc.test.id}"
	cidr_block = "10.1.2.0/24"
	availability_zone = "us-west-2b"
	tags = {
		Name = "tf-acc-neptune-subnet-group-generated-name-a"
	}
}

resource "aws_neptune_subnet_group" "test" {
	subnet_ids = ["${aws_subnet.a.id}", "${aws_subnet.b.id}"]
}`
