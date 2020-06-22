package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSRedshiftSubnetGroup_basic(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rInt := acctest.RandInt()
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRedshiftSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRedshiftSubnetGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRedshiftSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "test description"),
				),
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

func TestAccAWSRedshiftSubnetGroup_updateDescription(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rInt := acctest.RandInt()
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRedshiftSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRedshiftSubnetGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRedshiftSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "description", "test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccRedshiftSubnetGroup_updateDescription(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRedshiftSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "description", "test description updated"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftSubnetGroup_updateSubnetIds(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rInt := acctest.RandInt()
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRedshiftSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRedshiftSubnetGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRedshiftSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "subnet_ids.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccRedshiftSubnetGroupConfig_updateSubnetIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRedshiftSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "subnet_ids.#", "3"),
				),
			},
		},
	})
}

func TestAccAWSRedshiftSubnetGroup_tags(t *testing.T) {
	var v redshift.ClusterSubnetGroup
	rInt := acctest.RandInt()
	resourceName := "aws_redshift_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRedshiftSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRedshiftSubnetGroupConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRedshiftSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-redshift-subnetgroup"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"description"},
			},
			{
				Config: testAccRedshiftSubnetGroupConfigWithTagsUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRedshiftSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-redshift-subnetgroup"),
					resource.TestCheckResourceAttr(resourceName, "tags.test", "test2"),
				),
			},
		},
	})
}

func testAccCheckRedshiftSubnetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).redshiftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_subnet_group" {
			continue
		}

		resp, err := conn.DescribeClusterSubnetGroups(
			&redshift.DescribeClusterSubnetGroupsInput{
				ClusterSubnetGroupName: aws.String(rs.Primary.ID)})
		if err == nil {
			if len(resp.ClusterSubnetGroups) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		redshiftErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if redshiftErr.Code() != "ClusterSubnetGroupNotFoundFault" {
			return err
		}
	}

	return nil
}

func testAccCheckRedshiftSubnetGroupExists(n string, v *redshift.ClusterSubnetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		resp, err := conn.DescribeClusterSubnetGroups(
			&redshift.DescribeClusterSubnetGroupsInput{ClusterSubnetGroupName: aws.String(rs.Primary.ID)})
		if err != nil {
			return err
		}
		if len(resp.ClusterSubnetGroups) == 0 {
			return fmt.Errorf("ClusterSubnetGroup not found")
		}

		*v = *resp.ClusterSubnetGroups[0]

		return nil
	}
}

func testAccRedshiftSubnetGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-subnet-group"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-test"
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2b"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-test2"
  }
}

resource "aws_redshift_subnet_group" "test" {
  name        = "test-%d"
  description = "test description"
  subnet_ids  = ["${aws_subnet.test.id}", "${aws_subnet.test2.id}"]
}
`, rInt)
}

func testAccRedshiftSubnetGroup_updateDescription(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-subnet-group-upd-description"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-upd-description-test"
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2b"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-upd-description-test2"
  }
}

resource "aws_redshift_subnet_group" "test" {
  name        = "test-%d"
  description = "test description updated"
  subnet_ids  = ["${aws_subnet.test.id}", "${aws_subnet.test2.id}"]
}
`, rInt)
}

func testAccRedshiftSubnetGroupConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-subnet-group-with-tags"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-with-tags-test"
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2b"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-with-tags-test2"
  }
}

resource "aws_redshift_subnet_group" "test" {
  name       = "test-%d"
  subnet_ids = ["${aws_subnet.test.id}", "${aws_subnet.test2.id}"]

  tags = {
    Name = "tf-redshift-subnetgroup"
  }
}
`, rInt)
}

func testAccRedshiftSubnetGroupConfigWithTagsUpdated(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-subnet-group-with-tags"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-with-tags-test"
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2b"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-with-tags-test2"
  }
}

resource "aws_redshift_subnet_group" "test" {
  name       = "test-%d"
  subnet_ids = ["${aws_subnet.test.id}", "${aws_subnet.test2.id}"]

  tags = {
    Name        = "tf-redshift-subnetgroup"
    environment = "production"
    test         = "test2"
  }
}
`, rInt)
}

func testAccRedshiftSubnetGroupConfig_updateSubnetIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-redshift-subnet-group-upd-subnet-ids"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-west-2a"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-upd-subnet-ids-test"
  }
}

resource "aws_subnet" "test2" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = "us-west-2b"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-upd-subnet-ids-test2"
  }
}

resource "aws_subnet" "testtest2" {
  cidr_block        = "10.1.3.0/24"
  availability_zone = "us-west-2c"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-redshift-subnet-group-upd-subnet-ids-testtest2"
  }
}

resource "aws_redshift_subnet_group" "test" {
  name       = "test-%d"
  subnet_ids = ["${aws_subnet.test.id}", "${aws_subnet.test2.id}", "${aws_subnet.testtest2.id}"]
}
`, rInt)
}
