package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSElasticacheParameterGroup_basic(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.bar"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &v),
					testAccCheckAWSElasticacheParameterGroupAttributes(&v, rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "family", "redis2.8"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSElasticacheParameterGroup_addParameter(t *testing.T) {
	var v elasticache.CacheParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis2.8", "appendonly", "yes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					resource.TestCheckResourceAttr("aws_elasticache_parameter_group.bar", "parameter.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.name", "appendonly"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.value", "yes"),
				),
			},
			{
				ResourceName:      "aws_elasticache_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter2(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					resource.TestCheckResourceAttr("aws_elasticache_parameter_group.bar", "parameter.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.name", "appendonly"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.283487565.value", "yes"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.2196914567.name", "appendfsync"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "parameter.2196914567.value", "always"),
				),
			},
		},
	})
}

// Regression for https://github.com/terraform-providers/terraform-provider-aws/issues/116
func TestAccAWSElasticacheParameterGroup_removeAllParameters(t *testing.T) {
	var v elasticache.CacheParameterGroup
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter2(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					resource.TestCheckResourceAttr("aws_elasticache_parameter_group.bar", "parameter.#", "2"),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					resource.TestCheckResourceAttr("aws_elasticache_parameter_group.bar", "parameter.#", "0"),
				),
			},
		},
	})
}

// The API throws 500 errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccAWSElasticacheParameterGroup_removeReservedMemoryParameter(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.bar"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis3.2", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
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

// The API throws 500 errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccAWSElasticacheParameterGroup_switchReservedMemoryParameter(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.bar"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis3.2", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis3.2", "reserved-memory-percent", "25"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
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

// The API throws 500 errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccAWSElasticacheParameterGroup_updateReservedMemoryParameter(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.bar"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis2.8", "reserved-memory", "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis2.8", "reserved-memory", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
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

func TestAccAWSElasticacheParameterGroup_UppercaseName(t *testing.T) {
	var v elasticache.CacheParameterGroup
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("TF-ELASTIPG-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis2.8", "appendonly", "yes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists("aws_elasticache_parameter_group.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_elasticache_parameter_group.bar", "name", fmt.Sprintf("tf-elastipg-%d", rInt)),
				),
			},
			{
				ResourceName:      "aws_elasticache_parameter_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticacheParameterGroup_Description(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.bar"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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

func testAccCheckAWSElasticacheParameterGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_parameter_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeCacheParameterGroups(
			&elasticache.DescribeCacheParameterGroupsInput{
				CacheParameterGroupName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.CacheParameterGroups) != 0 &&
				*resp.CacheParameterGroups[0].CacheParameterGroupName == rs.Primary.ID {
				return fmt.Errorf("Cache Parameter Group still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "CacheParameterGroupNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSElasticacheParameterGroupAttributes(v *elasticache.CacheParameterGroup, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if *v.CacheParameterGroupName != rName {
			return fmt.Errorf("bad name: %#v", v.CacheParameterGroupName)
		}

		if *v.CacheParameterGroupFamily != "redis2.8" {
			return fmt.Errorf("bad family: %#v", v.CacheParameterGroupFamily)
		}

		return nil
	}
}

func testAccCheckAWSElasticacheParameterGroupExists(n string, v *elasticache.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cache Parameter Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

		opts := elasticache.DescribeCacheParameterGroupsInput{
			CacheParameterGroupName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeCacheParameterGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.CacheParameterGroups) != 1 ||
			*resp.CacheParameterGroups[0].CacheParameterGroupName != rs.Primary.ID {
			return fmt.Errorf("Cache Parameter Group not found")
		}

		*v = *resp.CacheParameterGroups[0]

		return nil
	}
}

func testAccAWSElasticacheParameterGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
  family = "redis2.8"
  name   = %q
}`, rName)
}

func testAccAWSElasticacheParameterGroupConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
  description = %q
  family      = "redis2.8"
  name        = %q
}`, description, rName)
}

func testAccAWSElasticacheParameterGroupConfigParameter1(rName, family, parameterName1, parameterValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
  family      = %q
  name        = %q

  parameter {
    name  = %q
    value = %q
  }
}`, family, rName, parameterName1, parameterValue1)
}

func testAccAWSElasticacheParameterGroupConfigParameter2(rName, family, parameterName1, parameterValue1, parameterName2, parameterValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "bar" {
  family      = %q
  name        = %q

  parameter {
    name  = %q
    value = %q
  }

  parameter {
    name  = %q
    value = %q
  }
}`, family, rName, parameterName1, parameterValue1, parameterName2, parameterValue2)
}
