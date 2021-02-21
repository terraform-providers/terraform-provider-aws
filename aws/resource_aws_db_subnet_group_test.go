package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_db_subnet_group", &resource.Sweeper{
		Name: "aws_db_subnet_group",
		F:    testSweepRdsDbSubnetGroups,
		Dependencies: []string{
			"aws_db_instance",
		},
	})
}

func testSweepRdsDbSubnetGroups(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).rdsconn
	input := &rds.DescribeDBSubnetGroupsInput{}

	err = conn.DescribeDBSubnetGroupsPages(input, func(out *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		for _, dbSubnetGroup := range out.DBSubnetGroups {
			name := aws.StringValue(dbSubnetGroup.DBSubnetGroupName)
			input := &rds.DeleteDBSubnetGroupInput{
				DBSubnetGroupName: dbSubnetGroup.DBSubnetGroupName,
			}

			log.Printf("[INFO] Deleting RDS DB Subnet Group: %s", name)

			_, err := conn.DeleteDBSubnetGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete RDS DB Subnet Group (%s): %s", name, err)
			}
		}
		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Subnet Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving RDS DB Subnet Groups: %s", err)
	}

	return nil
}

func TestAccAWSDBSubnetGroup_basic(t *testing.T) {
	var v rds.DBSubnetGroup

	resourceName := "aws_db_subnet_group.test"
	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("subgrp:%s$", rName))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description"},
			},
		},
	})
}

func TestAccAWSDBSubnetGroup_namePrefix(t *testing.T) {
	var v rds.DBSubnetGroup
	resourceName := "aws_db_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf_test-")),
				),
			},
		},
	})
}

func TestAccAWSDBSubnetGroup_generatedName(t *testing.T) {
	var v rds.DBSubnetGroup
	resourceName := "aws_db_subnet_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
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
	resourceName := "aws_db_subnet_group.underscores"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists("aws_db_subnet_group.underscores", &v),
					testAccCheckDBSubnetGroupExists("aws_db_subnet_group.periods", &v),
					testAccCheckDBSubnetGroupExists("aws_db_subnet_group.spaces", &v),
					testCheck,
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description"},
			},
		},
	})
}

func TestAccAWSDBSubnetGroup_updateDescription(t *testing.T) {
	var v rds.DBSubnetGroup
	resourceName := "aws_db_subnet_group.test"
	rName := fmt.Sprintf("tf-test-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDBSubnetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDBSubnetGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
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
				Config: testAccDBSubnetGroupConfig_updatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDBSubnetGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "test description updated"),
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
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-db-subnet-group"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-2"
  }
}

resource "aws_db_subnet_group" "test" {
  name       = "%s"
  subnet_ids = [aws_subnet.test.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-dbsubnet-group-test"
  }
}
`, rName)
}

func testAccDBSubnetGroupConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-db-subnet-group-updated-description"
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-1"
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-db-subnet-group-2"
  }
}

resource "aws_db_subnet_group" "test" {
  name        = "%s"
  description = "test description updated"
  subnet_ids  = [aws_subnet.test.id, aws_subnet.bar.id]

  tags = {
    Name = "tf-dbsubnet-group-test"
  }
}
`, rName)
}

const testAccDBSubnetGroupConfig_namePrefix = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-db-subnet-group-name-prefix"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-db-subnet-group-name-prefix-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  tags = {
    Name = "tf-acc-db-subnet-group-name-prefix-b"
  }
}

resource "aws_db_subnet_group" "test" {
  name_prefix = "tf_test-"
  subnet_ids  = [aws_subnet.a.id, aws_subnet.b.id]
}`

const testAccDBSubnetGroupConfig_generatedName = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-db-subnet-group-generated-name"
  }
}

resource "aws_subnet" "a" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-db-subnet-group-generated-name-a"
  }
}

resource "aws_subnet" "b" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]
  tags = {
    Name = "tf-acc-db-subnet-group-generated-name-a"
  }
}

resource "aws_db_subnet_group" "test" {
  subnet_ids = [aws_subnet.a.id, aws_subnet.b.id]
}`

const testAccDBSubnetGroupConfig_withUnderscoresAndPeriodsAndSpaces = `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "terraform-testacc-db-subnet-group-w-underscores-etc"
  }
}

resource "aws_subnet" "frontend" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.1.0/24"
  tags = {
    Name = "tf-acc-db-subnet-group-w-underscores-etc-front"
  }
}

resource "aws_subnet" "backend" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.2.0/24"
  tags = {
    Name = "tf-acc-db-subnet-group-w-underscores-etc-back"
  }
}

resource "aws_db_subnet_group" "underscores" {
  name        = "with_underscores"
  description = "Our main group of subnets"
  subnet_ids  = [aws_subnet.frontend.id, aws_subnet.backend.id]
}

resource "aws_db_subnet_group" "periods" {
  name        = "with.periods"
  description = "Our main group of subnets"
  subnet_ids  = [aws_subnet.frontend.id, aws_subnet.backend.id]
}

resource "aws_db_subnet_group" "spaces" {
  name        = "with spaces"
  description = "Our main group of subnets"
  subnet_ids  = [aws_subnet.frontend.id, aws_subnet.backend.id]
}
`
