package aws

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_elasticache_parameter_group", &resource.Sweeper{
		Name: "aws_elasticache_parameter_group",
		F:    testSweepElasticacheParameterGroups,
		Dependencies: []string{
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
		},
	})
}

func testSweepElasticacheParameterGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).elasticacheconn

	err = conn.DescribeCacheParameterGroupsPages(&elasticache.DescribeCacheParameterGroupsInput{}, func(page *elasticache.DescribeCacheParameterGroupsOutput, isLast bool) bool {
		if len(page.CacheParameterGroups) == 0 {
			log.Print("[DEBUG] No Elasticache Parameter Groups to sweep")
			return false
		}

		for _, parameterGroup := range page.CacheParameterGroups {
			name := aws.StringValue(parameterGroup.CacheParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping Elasticache Cache Parameter Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Elasticache Parameter Group: %s", name)
			_, err := conn.DeleteCacheParameterGroup(&elasticache.DeleteCacheParameterGroupInput{
				CacheParameterGroupName: aws.String(name),
			})
			if err != nil {
				log.Printf("[ERROR] Failed to delete Elasticache Parameter Group (%s): %s", name, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticache Parameter Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticache Parameter Group: %w", err)
	}
	return nil
}

func TestAccAWSElasticacheParameterGroup_basic(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
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
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis2.8", "appendonly", "yes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendonly",
						"value": "yes",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter2(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendonly",
						"value": "yes",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendfsync",
						"value": "always",
					}),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform-provider-aws/issues/116
func TestAccAWSElasticacheParameterGroup_removeAllParameters(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter2(rName, "redis2.8", "appendonly", "yes", "appendfsync", "always"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendonly",
						"value": "yes",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "appendfsync",
						"value": "always",
					}),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "0"),
				),
			},
		},
	})
}

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccAWSElasticacheParameterGroup_removeReservedMemoryParameter_AllParameters(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccAWSElasticacheParameterGroup_removeReservedMemoryParameter_RemainingParameters(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
	rName := fmt.Sprintf("parameter-group-test-terraform-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheParameterGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter2(rName, "redis3.2", "reserved-memory", "0", "tcp-keepalive", "360"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tcp-keepalive",
						"value": "360",
					}),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis3.2", "tcp-keepalive", "360"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "tcp-keepalive",
						"value": "360",
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccAWSElasticacheParameterGroup_switchReservedMemoryParameter(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis3.2", "reserved-memory-percent", "25"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory-percent",
						"value": "25",
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

// The API returns errors when attempting to reset the reserved-memory parameter.
// This covers our custom logic handling for this situation.
func TestAccAWSElasticacheParameterGroup_updateReservedMemoryParameter(t *testing.T) {
	var cacheParameterGroup1 elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "0",
					}),
				),
			},
			{
				Config: testAccAWSElasticacheParameterGroupConfigParameter1(rName, "redis2.8", "reserved-memory", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &cacheParameterGroup1),
					resource.TestCheckResourceAttr(resourceName, "parameter.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "parameter.*", map[string]string{
						"name":  "reserved-memory",
						"value": "1",
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

func TestAccAWSElasticacheParameterGroup_UppercaseName(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
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
					testAccCheckAWSElasticacheParameterGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-elastipg-%d", rInt)),
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

func TestAccAWSElasticacheParameterGroup_Description(t *testing.T) {
	var v elasticache.CacheParameterGroup
	resourceName := "aws_elasticache_parameter_group.test"
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

		if tfawserr.ErrCodeEquals(err, elasticache.ErrCodeCacheParameterGroupNotFoundFault) {
			return nil
		}
		if err != nil {
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
resource "aws_elasticache_parameter_group" "test" {
  family = "redis2.8"
  name   = %q
}
`, rName)
}

func testAccAWSElasticacheParameterGroupConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  description = %q
  family      = "redis2.8"
  name        = %q
}
`, description, rName)
}

func testAccAWSElasticacheParameterGroupConfigParameter1(rName, family, parameterName1, parameterValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %q
  name   = %q

  parameter {
    name  = %q
    value = %q
  }
}
`, family, rName, parameterName1, parameterValue1)
}

func testAccAWSElasticacheParameterGroupConfigParameter2(rName, family, parameterName1, parameterValue1, parameterName2, parameterValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  family = %q
  name   = %q

  parameter {
    name  = %q
    value = %q
  }

  parameter {
    name  = %q
    value = %q
  }
}
`, family, rName, parameterName1, parameterValue1, parameterName2, parameterValue2)
}

func TestFlattenElasticacheParameters(t *testing.T) {
	cases := []struct {
		Input  []*elasticache.Parameter
		Output []map[string]interface{}
	}{
		{
			Input: []*elasticache.Parameter{
				{
					ParameterName:  aws.String("activerehashing"),
					ParameterValue: aws.String("yes"),
				},
			},
			Output: []map[string]interface{}{
				{
					"name":  "activerehashing",
					"value": "yes",
				},
			},
		},
	}

	for _, tc := range cases {
		output := flattenElastiCacheParameters(tc.Input)
		if !reflect.DeepEqual(output, tc.Output) {
			t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v", output, tc.Output)
		}
	}
}

func TestExpandElasticacheParameters(t *testing.T) {
	expanded := []interface{}{
		map[string]interface{}{
			"name":         "activerehashing",
			"value":        "yes",
			"apply_method": "immediate",
		},
	}
	parameters := expandElastiCacheParameters(expanded)

	expected := &elasticache.ParameterNameValue{
		ParameterName:  aws.String("activerehashing"),
		ParameterValue: aws.String("yes"),
	}

	if !reflect.DeepEqual(parameters[0], expected) {
		t.Fatalf(
			"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
			parameters[0],
			expected)
	}
}

func TestElastiCacheParameterChanges(t *testing.T) {
	cases := []struct {
		Name                string
		Old                 *schema.Set
		New                 *schema.Set
		ExpectedRemove      []*elasticache.ParameterNameValue
		ExpectedAddOrUpdate []*elasticache.ParameterNameValue
	}{
		{
			Name:                "Empty",
			Old:                 new(schema.Set),
			New:                 new(schema.Set),
			ExpectedRemove:      []*elasticache.ParameterNameValue{},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "Remove all",
			Old: schema.NewSet(resourceAwsElasticacheParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			New: new(schema.Set),
			ExpectedRemove: []*elasticache.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "No change",
			Old: schema.NewSet(resourceAwsElasticacheParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			New: schema.NewSet(resourceAwsElasticacheParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
			}),
			ExpectedRemove:      []*elasticache.ParameterNameValue{},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "Remove partial",
			Old: schema.NewSet(resourceAwsElasticacheParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "reserved-memory",
					"value": "0",
				},
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			New: schema.NewSet(resourceAwsElasticacheParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			ExpectedRemove: []*elasticache.ParameterNameValue{
				{
					ParameterName:  aws.String("reserved-memory"),
					ParameterValue: aws.String("0"),
				},
			},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{},
		},
		{
			Name: "Add to existing",
			Old: schema.NewSet(resourceAwsElasticacheParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
			}),
			New: schema.NewSet(resourceAwsElasticacheParameterHash, []interface{}{
				map[string]interface{}{
					"name":  "appendonly",
					"value": "yes",
				},
				map[string]interface{}{
					"name":  "appendfsync",
					"value": "always",
				},
			}),
			ExpectedRemove: []*elasticache.ParameterNameValue{},
			ExpectedAddOrUpdate: []*elasticache.ParameterNameValue{
				{
					ParameterName:  aws.String("appendfsync"),
					ParameterValue: aws.String("always"),
				},
			},
		},
	}

	for _, tc := range cases {
		remove, addOrUpdate := elastiCacheParameterChanges(tc.Old, tc.New)
		if !reflect.DeepEqual(remove, tc.ExpectedRemove) {
			t.Errorf("Case %q: Remove did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedRemove, remove)
		}
		if !reflect.DeepEqual(addOrUpdate, tc.ExpectedAddOrUpdate) {
			t.Errorf("Case %q: AddOrUpdate did not match\n%#v\n\nGot:\n%#v", tc.Name, tc.ExpectedAddOrUpdate, addOrUpdate)
		}
	}
}
