package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDocDBClusterParameterGroup_basic(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster_parameter_group.bar", "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:rds:[^:]+:\\d{12}:cluster-pg:%s", parameterGroupName))),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster_parameter_group.bar", "name", parameterGroupName),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster_parameter_group.bar", "family", "docdb3.6"),
					resource.TestCheckResourceAttr(
						"aws_docdb_cluster_parameter_group.bar", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.#", "0"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "tags.%", "0"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDocDBClusterParameterGroup_namePrefix(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_docdb_cluster_parameter_group.test", "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            "aws_docdb_cluster_parameter_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSDocDBClusterParameterGroup_generatedName(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.test", &v),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster_parameter_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDocDBClusterParameterGroup_Description(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_Description(parameterGroupName, "custom description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "description", "custom description"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDocDBClusterParameterGroup_disappears(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDocDBClusterParameterGroup_Parameter(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_Parameter(parameterGroupName, "tls", "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.#", "1"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.3297634353.apply_method", "pending-reboot"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.3297634353.name", "tls"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.3297634353.value", "disabled"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_Parameter(parameterGroupName, "tls", "enabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.#", "1"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.4005179180.apply_method", "pending-reboot"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.4005179180.name", "tls"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "parameter.4005179180.value", "enabled"),
				),
			},
		},
	})
}

func TestAccAWSDocDBClusterParameterGroup_Tags(t *testing.T) {
	var v docdb.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDocDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      "aws_docdb_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "tags.key1", "value2"),
				),
			},
			{
				Config: testAccAWSDocDBClusterParameterGroupConfig_Tags(parameterGroupName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDocDBClusterParameterGroupExists("aws_docdb_cluster_parameter_group.bar", &v),
					testAccCheckAWSDocDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_docdb_cluster_parameter_group.bar", "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSDocDBClusterParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).docdbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_docdb_cluster_parameter_group" {
			continue
		}

		resp, err := conn.DescribeDBClusterParameterGroups(
			&docdb.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusterParameterGroups) != 0 &&
				aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) == rs.Primary.ID {
				return errors.New("DocDB Cluster Parameter Group still exists")
			}
		}

		if err != nil {
			if isAWSErr(err, docdb.ErrCodeDBParameterGroupNotFoundFault, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSDocDBClusterParameterGroupDisappears(group *docdb.DBClusterParameterGroup) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).docdbconn

		params := &docdb.DeleteDBClusterParameterGroupInput{
			DBClusterParameterGroupName: group.DBClusterParameterGroupName,
		}

		_, err := conn.DeleteDBClusterParameterGroup(params)
		if err != nil {
			return err
		}

		return waitForDocDBClusterParameterGroupDeletion(conn, *group.DBClusterParameterGroupName)
	}
}

func testAccCheckAWSDocDBClusterParameterGroupAttributes(v *docdb.DBClusterParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.DBClusterParameterGroupName != name {
			return fmt.Errorf("bad name: %#v expected: %v", *v.DBClusterParameterGroupName, name)
		}

		if *v.DBParameterGroupFamily != "docdb3.6" {
			return fmt.Errorf("bad family: %#v", *v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckAWSDocDBClusterParameterGroupExists(n string, v *docdb.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No DocDB Cluster Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).docdbconn

		opts := docdb.DescribeDBClusterParameterGroupsInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBClusterParameterGroups) != 1 ||
			aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) != rs.Primary.ID {
			return fmt.Errorf("DocDB Cluster Parameter Group not found: %s", rs.Primary.ID)
		}

		*v = *resp.DBClusterParameterGroups[0]

		return nil
	}
}

func testAccAWSDocDBClusterParameterGroupConfig(name string) string {
	return fmt.Sprintf(`resource "aws_docdb_cluster_parameter_group" "bar" {
	family = "docdb3.6"
	name   = "%s"
}`, name)
}

func testAccAWSDocDBClusterParameterGroupConfig_Description(name, description string) string {
	return fmt.Sprintf(`resource "aws_docdb_cluster_parameter_group" "bar" {
  family = "docdb3.6"
  description = "%s"
  name        = "%s"
}`, description, name)
}

func testAccAWSDocDBClusterParameterGroupConfig_Parameter(name, pName, pValue string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_parameter_group" "bar" {
  name   = "%s"
  family = "docdb3.6"

  parameter {
    name  = "%s"
    value = "%s"
  }
}
`, name, pName, pValue)
}

func testAccAWSDocDBClusterParameterGroupConfig_Tags(name, tKey, tValue string) string {
	return fmt.Sprintf(`
resource "aws_docdb_cluster_parameter_group" "bar" {
  name   = "%s"
  family = "docdb3.6"

  tags = {
    %s = "%s"
  }
}
`, name, tKey, tValue)
}

const testAccAWSDocDBClusterParameterGroupConfig_namePrefix = `
resource "aws_docdb_cluster_parameter_group" "test" {
  name_prefix = "tf-test-"
  family = "docdb3.6"
}
`
const testAccAWSDocDBClusterParameterGroupConfig_generatedName = `
resource "aws_docdb_cluster_parameter_group" "test" {
	family = "docdb3.6"
}
`
