package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_elasticache_global_replication_group", &resource.Sweeper{
		Name: "aws_elasticache_global_replication_group",
		F:    testSweepElasticacheGlobalReplicationGroups,
	})
}

func testSweepElasticacheGlobalReplicationGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).elasticacheconn

	var sweeperErrs *multierror.Error

	input := &elasticache.DescribeGlobalReplicationGroupsInput{}
	err = conn.DescribeGlobalReplicationGroupsPages(input, func(page *elasticache.DescribeGlobalReplicationGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, globalReplicationGroup := range page.GlobalReplicationGroups {
			id := aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId)

			log.Printf("[INFO] Deleting ElastiCache Global Replication Group: %s", id)
			err := deleteElasticacheGlobalReplicationGroup(conn, id)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting ElastiCache Global Replication Group (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping ElastiCache Global Replication Group sweep for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing ElastiCache Global Replication Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSElasticacheGlobalReplicationGroup_basic(t *testing.T) {
	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := acctest.RandomWithPrefix("tf-acc-test")
	primaryReplicationGroupId := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticacheGlobalReplicationGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					testAccCheckAWSElasticacheReplicationGroupExists(primaryReplicationGroupResourceName, &primaryReplicationGroup),
					testAccMatchResourceAttrGlobalARN(resourceName, "arn", "elasticache", regexp.MustCompile(`globalreplicationgroup:`+elasticacheGlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryReplicationGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttr(resourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_enabled", primaryReplicationGroupResourceName, "cluster_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", primaryReplicationGroupResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "actual_engine_version", primaryReplicationGroupResourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_id_suffix", rName),
					resource.TestMatchResourceAttr(resourceName, "global_replication_group_id", regexp.MustCompile(elasticacheGlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", elasticacheEmptyDescription),
					resource.TestCheckResourceAttr(resourceName, "primary_replication_group_id", primaryReplicationGroupId),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", "false"),
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

func TestAccAWSElasticacheGlobalReplicationGroup_Description(t *testing.T) {
	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	primaryReplicationGroupId := acctest.RandomWithPrefix("tf-acc-test")
	description1 := acctest.RandString(10)
	description2 := acctest.RandString(10)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticacheGlobalReplicationGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSElasticacheGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description2),
				),
			},
		},
	})
}

func TestAccAWSElasticacheGlobalReplicationGroup_disappears(t *testing.T) {
	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	primaryReplicationGroupId := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSElasticacheGlobalReplicationGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsElasticacheGlobalReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName string, v *elasticache.GlobalReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Global Replication Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
		grg, err := finder.GlobalReplicationGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error retrieving ElastiCache Global Replication Group (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(grg.Status) != waiter.GlobalReplicationGroupStatusAvailable && aws.StringValue(grg.Status) != waiter.GlobalReplicationGroupStatusPrimaryOnly {
			return fmt.Errorf("ElastiCache Global Replication Group (%s) exists, but is in a non-available state: %s", rs.Primary.ID, aws.StringValue(grg.Status))
		}

		*v = *grg

		return nil
	}
}

func testAccCheckAWSElasticacheGlobalReplicationGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_global_replication_group" {
			continue
		}

		_, err := finder.GlobalReplicationGroupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("ElastiCache Global Replication Group (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPreCheckAWSElasticacheGlobalReplicationGroup(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	input := &elasticache.DescribeGlobalReplicationGroupsInput{}
	_, err := conn.DescribeGlobalReplicationGroups(input)

	if testAccPreCheckSkipError(err) ||
		tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "Access Denied to API Version: APIGlobalDatastore") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSElasticacheGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, primaryReplicationGroupId)
}

func testAccAWSElasticacheGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  global_replication_group_description = %[3]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, description)
}
