package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSRedshiftParameterGroup_basic(t *testing.T) {
	var v redshift.ClusterParameterGroup
	resourceName := "aws_redshift_parameter_group.test"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftParameterGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftParameterGroupExists(resourceName, &v),
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

func TestAccAWSRedshiftParameterGroup_withParameters(t *testing.T) {
	var v redshift.ClusterParameterGroup
	rInt := acctest.RandInt()
	resourceName := "aws_redshift_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftParameterGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("test-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "family", "redshift-1.0"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Managed by Terraform"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "require_ssl",
						"value": "true",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "query_group",
						"value": "example",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "enable_user_activity_logging",
						"value": "true",
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

func TestAccAWSRedshiftParameterGroup_withoutParameters(t *testing.T) {
	var v redshift.ClusterParameterGroup
	rInt := acctest.RandInt()
	resourceName := "aws_redshift_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftParameterGroupOnlyConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "name", fmt.Sprintf("test-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "family", "redshift-1.0"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Test parameter group for terraform"),
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

func TestAccAWSRedshiftParameterGroup_withTags(t *testing.T) {
	var v redshift.ClusterParameterGroup
	rInt := acctest.RandInt()
	resourceName := "aws_redshift_parameter_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftParameterGroupConfigWithTags(rInt, "aaa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", fmt.Sprintf("test-terraform-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.description", fmt.Sprintf("Test parameter group for terraform %s", "aaa")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRedshiftParameterGroupConfigWithTags(rInt, "bbb"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.description", fmt.Sprintf("Test parameter group for terraform %s", "bbb")),
				),
			},
			{
				Config: testAccAWSRedshiftParameterGroupConfigWithTagsUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", fmt.Sprintf("test-terraform-%d", rInt)),
				),
			},
		},
	})
}

func testAccCheckAWSRedshiftParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).redshiftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeClusterParameterGroups(
			&redshift.DescribeClusterParameterGroupsInput{
				ParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.ParameterGroups) != 0 &&
				*resp.ParameterGroups[0].ParameterGroupName == rs.Primary.ID {
				return fmt.Errorf("Redshift Parameter Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "ClusterParameterGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSRedshiftParameterGroupExists(n string, v *redshift.ClusterParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn

		opts := redshift.DescribeClusterParameterGroupsInput{
			ParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeClusterParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.ParameterGroups) != 1 ||
			*resp.ParameterGroups[0].ParameterGroupName != rs.Primary.ID {
			return fmt.Errorf("Redshift Parameter Group not found")
		}

		*v = *resp.ParameterGroups[0]

		return nil
	}
}

func testAccAWSRedshiftParameterGroupOnlyConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name        = "test-terraform-%d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"
}
`, rInt)
}

func testAccAWSRedshiftParameterGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name   = "test-terraform-%d"
  family = "redshift-1.0"

  parameter {
    name  = "require_ssl"
    value = "true"
  }

  parameter {
    name  = "query_group"
    value = "example"
  }

  parameter {
    name  = "enable_user_activity_logging"
    value = "true"
  }
}
`, rInt)
}

func testAccAWSRedshiftParameterGroupConfigWithTags(rInt int, rString string) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name        = "test-terraform-%[1]d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"

  tags = {
    environment = "Production"
    name        = "test-terraform-%[1]d"
    description = "Test parameter group for terraform %[2]s"
  }
}
`, rInt, rString)
}

func testAccAWSRedshiftParameterGroupConfigWithTagsUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_redshift_parameter_group" "test" {
  name        = "test-terraform-%[1]d"
  family      = "redshift-1.0"
  description = "Test parameter group for terraform"

  tags = {
    name = "test-terraform-%[1]d"
  }
}
`, rInt)
}
