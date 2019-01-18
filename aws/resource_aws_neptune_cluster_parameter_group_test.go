package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSNeptuneClusterParameterGroup_basic(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:rds:[^:]+:\\d{12}:cluster-pg:%s", parameterGroupName))),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "name", parameterGroupName),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "family", "neptune1"),
					resource.TestCheckResourceAttr(
						"aws_neptune_cluster_parameter_group.bar", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.#", "0"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "tags.%", "0"),
				),
			},
			{
				ResourceName:      "aws_neptune_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_namePrefix(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.test", &v),
					resource.TestMatchResourceAttr(
						"aws_neptune_cluster_parameter_group.test", "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            "aws_neptune_cluster_parameter_group.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_generatedName(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.test", &v),
				),
			},
			{
				ResourceName:      "aws_neptune_cluster_parameter_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_Description(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Description(parameterGroupName, "custom description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "description", "custom description"),
				),
			},
			{
				ResourceName:      "aws_neptune_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_Parameter(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Parameter(parameterGroupName, "neptune_enable_audit_log", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.#", "1"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.709171678.apply_method", "pending-reboot"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.709171678.name", "neptune_enable_audit_log"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.709171678.value", "1"),
				),
			},
			{
				ResourceName:      "aws_neptune_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Parameter(parameterGroupName, "neptune_enable_audit_log", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.#", "1"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.861808799.apply_method", "pending-reboot"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.861808799.name", "neptune_enable_audit_log"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "parameter.861808799.value", "0"),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_Tags(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      "aws_neptune_cluster_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "tags.key1", "value2"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Tags(parameterGroupName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists("aws_neptune_cluster_parameter_group.bar", &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_neptune_cluster_parameter_group.bar", "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSNeptuneClusterParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_cluster_parameter_group" {
			continue
		}

		resp, err := conn.DescribeDBClusterParameterGroups(
			&neptune.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusterParameterGroups) != 0 &&
				aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) == rs.Primary.ID {
				return errors.New("Neptune Cluster Parameter Group still exists")
			}
		}

		if err != nil {
			if isAWSErr(err, neptune.ErrCodeDBParameterGroupNotFoundFault, "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func testAccCheckAWSNeptuneClusterParameterGroupAttributes(v *neptune.DBClusterParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.DBClusterParameterGroupName != name {
			return fmt.Errorf("bad name: %#v expected: %v", *v.DBClusterParameterGroupName, name)
		}

		if *v.DBParameterGroupFamily != "neptune1" {
			return fmt.Errorf("bad family: %#v", *v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckAWSNeptuneClusterParameterGroupExists(n string, v *neptune.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Neptune Cluster Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn

		opts := neptune.DescribeDBClusterParameterGroupsInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBClusterParameterGroups) != 1 ||
			aws.StringValue(resp.DBClusterParameterGroups[0].DBClusterParameterGroupName) != rs.Primary.ID {
			return errors.New("Neptune Cluster Parameter Group not found")
		}

		*v = *resp.DBClusterParameterGroups[0]

		return nil
	}
}

func testAccAWSNeptuneClusterParameterGroupConfig_Description(name, description string) string {
	return fmt.Sprintf(`resource "aws_neptune_cluster_parameter_group" "bar" {
  description = "%s"
  family      = "neptune1"
  name        = "%s"
}`, description, name)
}

func testAccAWSNeptuneClusterParameterGroupConfig_Parameter(name, pName, pValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "bar" {
  family = "neptune1"
  name   = "%s"

  parameter {
    name  = "%s"
    value = "%s"
  }
}
`, name, pName, pValue)
}

func testAccAWSNeptuneClusterParameterGroupConfig_Tags(name, tKey, tValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "bar" {
  family = "neptune1"
  name   = "%s"

  tags = {
    %s = "%s"
  }
}
`, name, tKey, tValue)
}

func testAccAWSNeptuneClusterParameterGroupConfig(name string) string {
	return fmt.Sprintf(`resource "aws_neptune_cluster_parameter_group" "bar" {
  family = "neptune1"
  name   = "%s"
}`, name)
}

const testAccAWSNeptuneClusterParameterGroupConfig_namePrefix = `
resource "aws_neptune_cluster_parameter_group" "test" {
  family      = "neptune1"
  name_prefix = "tf-test-"
}
`
const testAccAWSNeptuneClusterParameterGroupConfig_generatedName = `
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
}
`
