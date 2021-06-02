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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		ShowMemberInfo: aws.Bool(true),
	}
	err = conn.DescribeGlobalReplicationGroupsPages(input, func(page *elasticache.DescribeGlobalReplicationGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, globalReplicationGroup := range page.GlobalReplicationGroups {
			id := aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId)

			for _, member := range globalReplicationGroup.Members {
				if aws.StringValue(member.Role) == GlobalReplicationGroupMemberRolePrimary {
					continue
				}

				if err := disassociateElasticacheReplicationGroup(conn, id, aws.StringValue(member.ReplicationGroupId), aws.StringValue(member.ReplicationGroupRegion)); err != nil {
					sweeperErr := fmt.Errorf(
						"error disassociating ElastiCache Replication Group (%s) in %s from Global Group (%s): %w",
						aws.StringValue(member.ReplicationGroupId), aws.StringValue(member.ReplicationGroupRegion), id, err,
					)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				}
			}

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
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
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
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version"),
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
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, elasticache.EndpointsID),
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

func TestAccAWSElasticacheGlobalReplicationGroup_MultipleSecondaries(t *testing.T) {
	var providers []*schema.Provider
	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 3)
		},
		ErrorCheck:        testAccErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheGlobalReplicationGroupConfig_MultipleSecondaries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
		},
	})
}

func TestAccAWSElasticacheGlobalReplicationGroup_ReplaceSecondary_DifferentRegion(t *testing.T) {
	var providers []*schema.Provider
	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 3)
		},
		ErrorCheck:        testAccErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Setup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Move(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
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

		if aws.StringValue(grg.Status) == waiter.GlobalReplicationGroupStatusDeleting || aws.StringValue(grg.Status) == waiter.GlobalReplicationGroupStatusDeleted {
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

func testAccAWSElasticacheGlobalReplicationGroupConfig_MultipleSecondaries(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(3),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", ProviderNameAws),
		testAccElasticacheVpcBaseWithProvider(rName, "alternate", ProviderNameAwsAlternate),
		testAccElasticacheVpcBaseWithProvider(rName, "third", ProviderNameAwsThird),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "alternate" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.alternate.name

  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccAWSElasticacheReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Setup(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(3),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", ProviderNameAws),
		testAccElasticacheVpcBaseWithProvider(rName, "secondary", ProviderNameAwsAlternate),
		testAccElasticacheVpcBaseWithProvider(rName, "third", ProviderNameAwsThird),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.secondary.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccAWSElasticacheReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Move(rName string) string {
	return composeConfig(
		testAccMultipleRegionProviderConfig(3),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", ProviderNameAws),
		testAccElasticacheVpcBaseWithProvider(rName, "secondary", ProviderNameAwsAlternate),
		testAccElasticacheVpcBaseWithProvider(rName, "third", ProviderNameAwsThird),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccElasticacheVpcBaseWithProvider(rName, name, provider string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfigWithProvider(name, provider),
		fmt.Sprintf(`
resource "aws_vpc" "%[1]s" {
  provider = %[2]s

  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "%[1]s" {
  provider = %[2]s
	
  vpc_id            = aws_vpc.%[1]s.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.%[1]s.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_elasticache_subnet_group" "%[1]s" {
  provider = %[2]s
	
  name        = %[3]q
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.%[1]s.id,
  ]
}
`, name, provider, rName),
	)
}

func testAccAvailableAZsNoOptInConfigWithProvider(name, provider string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "%[1]s" {
  provider = %[2]s

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`, name, provider)
}
