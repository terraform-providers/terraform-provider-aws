package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_rds_cluster_parameter_group", &resource.Sweeper{
		Name: "aws_rds_cluster_parameter_group",
		F:    testSweepRdsClusterParameterGroups,
		Dependencies: []string{
			"aws_rds_cluster",
		},
	})
}

func testSweepRdsClusterParameterGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).rdsconn

	input := &rds.DescribeDBClusterParameterGroupsInput{}

	for {
		output, err := conn.DescribeDBClusterParameterGroups(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Cluster Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving DB Cluster Parameter Groups: %s", err)
		}

		for _, dbcpg := range output.DBClusterParameterGroups {
			if dbcpg == nil {
				continue
			}

			input := &rds.DeleteDBClusterParameterGroupInput{
				DBClusterParameterGroupName: dbcpg.DBClusterParameterGroupName,
			}
			name := aws.StringValue(dbcpg.DBClusterParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping DB Cluster Parameter Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting DB Cluster Parameter Group: %s", name)

			_, err := conn.DeleteDBClusterParameterGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DB Cluster Parameter Group %s: %s", name, err)
				continue
			}
		}

		if aws.StringValue(output.Marker) == "" {
			break
		}

		input.Marker = output.Marker
	}

	return nil
}

func TestAccAWSDBClusterParameterGroup_basic(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBClusterParameterGroupAttributes(&v, parameterGroupName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("cluster-pg:%s", parameterGroupName)),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test cluster parameter group for terraform"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBClusterParameterGroupAddParametersConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test cluster parameter group for terraform"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_connection",
						"value": "utf8_unicode_ci",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "collation_server",
						"value": "utf8_unicode_ci",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
			{
				Config: testAccAWSDBClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBClusterParameterGroupAttributes(&v, parameterGroupName),
					testAccCheckAWSRDSClusterParameterNotUserDefined(resourceName, "collation_connection"),
					testAccCheckAWSRDSClusterParameterNotUserDefined(resourceName, "collation_server"),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "3"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
		},
	})
}

func TestAccAWSDBClusterParameterGroup_withApplyMethod(t *testing.T) {
	var v rds.DBClusterParameterGroup
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupConfigWithApplyMethod(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBClusterParameterGroupAttributes(&v, parameterGroupName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("cluster-pg:%s", parameterGroupName)),
					resource.TestCheckResourceAttr(resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(resourceName, "description", "Test cluster parameter group for terraform"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":         "character_set_server",
						"value":        "utf8",
						"apply_method": "immediate",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":         "character_set_client",
						"value":        "utf8",
						"apply_method": "pending-reboot",
					}),
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

func TestAccAWSDBClusterParameterGroup_namePrefix(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					resource.TestMatchResourceAttr(
						resourceName, "name", regexp.MustCompile("^tf-test-")),
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

func TestAccAWSDBClusterParameterGroup_namePrefix_Parameter(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupConfig_namePrefix_Parameter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					resource.TestMatchResourceAttr(
						resourceName, "name", regexp.MustCompile("^tf-test-")),
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

func TestAccAWSDBClusterParameterGroup_generatedName(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupConfig_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
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

func TestAccAWSDBClusterParameterGroup_generatedName_Parameter(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupConfig_generatedName_Parameter,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
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

func TestAccAWSDBClusterParameterGroup_disappears(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccAWSDBClusterParameterGroupDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDBClusterParameterGroup_only(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	parameterGroupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupOnlyConfig(parameterGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBClusterParameterGroupAttributes(&v, parameterGroupName),
					resource.TestCheckResourceAttr(
						resourceName, "name", parameterGroupName),
					resource.TestCheckResourceAttr(
						resourceName, "family", "aurora5.6"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Managed by Terraform"),
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

func TestAccAWSDBClusterParameterGroup_updateParameters(t *testing.T) {
	var v rds.DBClusterParameterGroup
	resourceName := "aws_rds_cluster_parameter_group.test"
	groupName := fmt.Sprintf("cluster-parameter-group-test-tf-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBClusterParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBClusterParameterGroupUpdateParametersInitialConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBClusterParameterGroupAttributes(&v, groupName),
					resource.TestCheckResourceAttr(resourceName, "name", groupName),
					resource.TestCheckResourceAttr(resourceName, "family", "aurora5.6"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "utf8",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBClusterParameterGroupUpdateParametersUpdatedConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBClusterParameterGroupExists(resourceName, &v),
					testAccCheckAWSDBClusterParameterGroupAttributes(&v, groupName),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_results",
						"value": "ascii",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_server",
						"value": "ascii",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "character_set_client",
						"value": "utf8",
					}),
				),
			},
		},
	})
}

func testAccCheckAWSDBClusterParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rds_cluster_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBClusterParameterGroups(
			&rds.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBClusterParameterGroups) != 0 &&
				*resp.DBClusterParameterGroups[0].DBClusterParameterGroupName == rs.Primary.ID {
				return errors.New("DB Cluster Parameter Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "DBParameterGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSRDSClusterParameterNotUserDefined(n, paramName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBClusterParametersInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		userDefined := false
		out, err := conn.DescribeDBClusterParameters(&opts)
		for _, param := range out.Parameters {
			if *param.ParameterName == paramName && aws.StringValue(param.ParameterValue) != "" {
				// Some of these resets leave the parameter name present but with a nil value
				userDefined = true
			}
		}

		if userDefined {
			return fmt.Errorf("DB Parameter %s is user defined", paramName)
		}
		return err
	}
}

func testAccCheckAWSDBClusterParameterGroupAttributes(v *rds.DBClusterParameterGroup, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *v.DBClusterParameterGroupName != name {
			return fmt.Errorf("bad name: %#v expected: %v", *v.DBClusterParameterGroupName, name)
		}

		if *v.DBParameterGroupFamily != "aurora5.6" {
			return fmt.Errorf("bad family: %#v", *v.DBParameterGroupFamily)
		}

		return nil
	}
}

func testAccAWSDBClusterParameterGroupDisappears(v *rds.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn
		opts := &rds.DeleteDBClusterParameterGroupInput{
			DBClusterParameterGroupName: v.DBClusterParameterGroupName,
		}
		if _, err := conn.DeleteDBClusterParameterGroup(opts); err != nil {
			return err
		}
		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			opts := &rds.DescribeDBClusterParameterGroupsInput{
				DBClusterParameterGroupName: v.DBClusterParameterGroupName,
			}
			_, err := conn.DescribeDBClusterParameterGroups(opts)
			if err != nil {
				dbparamgrouperr, ok := err.(awserr.Error)
				if ok && dbparamgrouperr.Code() == "DBParameterGroupNotFound" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving DB Cluster Parameter Groups: %s", err))
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for cluster parameter group to be deleted: %v", v.DBClusterParameterGroupName))
		})
	}
}

func testAccCheckAWSDBClusterParameterGroupExists(n string, v *rds.DBClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No DB Cluster Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBClusterParameterGroupsInput{
			DBClusterParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBClusterParameterGroups) != 1 ||
			*resp.DBClusterParameterGroups[0].DBClusterParameterGroupName != rs.Primary.ID {
			return errors.New("DB Cluster Parameter Group not found")
		}

		*v = *resp.DBClusterParameterGroups[0]

		return nil
	}
}

func testAccAWSDBClusterParameterGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%s"
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  tags = {
    foo = "bar"
  }
}
`, name)
}

func testAccAWSDBClusterParameterGroupConfigWithApplyMethod(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%s"
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name         = "character_set_client"
    value        = "utf8"
    apply_method = "pending-reboot"
  }

  tags = {
    foo = "bar"
  }
}
`, name)
}

func testAccAWSDBClusterParameterGroupAddParametersConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name        = "%s"
  family      = "aurora5.6"
  description = "Test cluster parameter group for terraform"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }

  parameter {
    name  = "collation_server"
    value = "utf8_unicode_ci"
  }

  parameter {
    name  = "collation_connection"
    value = "utf8_unicode_ci"
  }

  tags = {
    foo = "bar"
    baz = "foo"
  }
}
`, name)
}

func testAccAWSDBClusterParameterGroupOnlyConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = "%s"
  family = "aurora5.6"
}
`, name)
}

func testAccAWSDBClusterParameterGroupUpdateParametersInitialConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = "%s"
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "utf8"
  }
}
`, name)
}

func testAccAWSDBClusterParameterGroupUpdateParametersUpdatedConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_rds_cluster_parameter_group" "test" {
  name   = "%s"
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "ascii"
  }

  parameter {
    name  = "character_set_client"
    value = "utf8"
  }

  parameter {
    name  = "character_set_results"
    value = "ascii"
  }
}
`, name)
}

const testAccAWSDBClusterParameterGroupConfig_namePrefix = `
resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = "tf-test-"
  family      = "aurora5.6"
}
`

const testAccAWSDBClusterParameterGroupConfig_namePrefix_Parameter = `
resource "aws_rds_cluster_parameter_group" "test" {
  name_prefix = "tf-test-"
  family      = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }
}
`

const testAccAWSDBClusterParameterGroupConfig_generatedName = `
resource "aws_rds_cluster_parameter_group" "test" {
  family = "aurora5.6"
}
`

const testAccAWSDBClusterParameterGroupConfig_generatedName_Parameter = `
resource "aws_rds_cluster_parameter_group" "test" {
  family = "aurora5.6"

  parameter {
    name  = "character_set_server"
    value = "utf8"
  }
}
`
