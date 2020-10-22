package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSNeptuneClusterParameterGroup_basic(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	parameterGroupName := acctest.RandomWithPrefix("cluster-parameter-group-test-terraform")

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("cluster-pg:%s", parameterGroupName)),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "neptune1"),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSNeptuneClusterParameterGroup_namePrefix(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_generatedName(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
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

func TestAccAWSNeptuneClusterParameterGroup_Description(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := acctest.RandomWithPrefix("cluster-parameter-group-test-terraform")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Description(parameterGroupName, "custom description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "custom description"),
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

func TestAccAWSNeptuneClusterParameterGroup_Parameter(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := acctest.RandomWithPrefix("cluster-parameter-group-test-tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Parameter(parameterGroupName, "neptune_enable_audit_log", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_enable_audit_log",
						"value":        "1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Parameter(parameterGroupName, "neptune_enable_audit_log", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"apply_method": "pending-reboot",
						"name":         "neptune_enable_audit_log",
						"value":        "0",
					}),
				),
			},
		},
	})
}

func TestAccAWSNeptuneClusterParameterGroup_Tags(t *testing.T) {
	var v neptune.DBClusterParameterGroup

	resourceName := "aws_neptune_cluster_parameter_group.test"

	parameterGroupName := acctest.RandomWithPrefix("cluster-parameter-group-test-tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
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
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Tags(parameterGroupName, "key1", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value2"),
				),
			},
			{
				Config: testAccAWSNeptuneClusterParameterGroupConfig_Tags(parameterGroupName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSNeptuneClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  description = "%s"
  family      = "neptune1"
  name        = "%s"
}
`, description, name)
}

func testAccAWSNeptuneClusterParameterGroupConfig_Parameter(name, pName, pValue string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
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
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"

  tags = {
    %s = "%s"
  }
}
`, name, tKey, tValue)
}

func testAccAWSNeptuneClusterParameterGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_neptune_cluster_parameter_group" "test" {
  family = "neptune1"
  name   = "%s"
}
`, name)
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
