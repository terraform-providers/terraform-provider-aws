package aws

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/elasticache/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_elasticache_replication_group", &resource.Sweeper{
		Name: "aws_elasticache_replication_group",
		F:    testSweepElasticacheReplicationGroups,
		Dependencies: []string{
			"aws_elasticache_global_replication_group",
		},
	})
}

func testSweepElasticacheReplicationGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).elasticacheconn

	err = conn.DescribeReplicationGroupsPages(&elasticache.DescribeReplicationGroupsInput{}, func(page *elasticache.DescribeReplicationGroupsOutput, isLast bool) bool {
		if len(page.ReplicationGroups) == 0 {
			log.Print("[DEBUG] No ElastiCache Replicaton Groups to sweep")
			return false
		}

		for _, replicationGroup := range page.ReplicationGroups {
			id := aws.StringValue(replicationGroup.ReplicationGroupId)

			log.Printf("[INFO] Deleting ElastiCache Replication Group: %s", id)
			err := deleteElasticacheReplicationGroup(id, conn, "", 40*time.Minute)
			if err != nil {
				log.Printf("[ERROR] Failed to delete ElastiCache Replication Group (%s): %s", id, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ElastiCache Replication Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving ElastiCache Replication Groups: %w", err)
	}
	return nil
}

func TestAccAWSElasticacheReplicationGroup_basic(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "elasticache", fmt.Sprintf("replicationgroup:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"}, //not in the API
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_Uppercase(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_Uppercase(strings.ToUpper(rName)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "replication_group_id", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_updateDescription(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "replication_group_description", "test description"),
					resource.TestCheckResourceAttr(
						resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigUpdatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "replication_group_description", "updated description"),
					resource.TestCheckResourceAttr(
						resourceName, "auto_minor_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_updateMaintenanceWindow(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_window", "tue:06:30-tue:07:30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigUpdatedMaintenanceWindow(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "maintenance_window", "wed:03:00-wed:06:00"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_updateNodeSize(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigUpdatedNodeSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

//This is a test to prove that we panic we get in https://github.com/hashicorp/terraform/issues/9097
func TestAccAWSElasticacheReplicationGroup_updateParameterGroup(t *testing.T) {
	var rg elasticache.ReplicationGroup
	parameterGroupResourceName1 := "aws_elasticache_parameter_group.test.0"
	parameterGroupResourceName2 := "aws_elasticache_parameter_group.test.1"
	resourceName := "aws_elasticache_replication_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfigParameterGroupName(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", parameterGroupResourceName1, "name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigParameterGroupName(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, "parameter_group_name", parameterGroupResourceName2, "name"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_vpc(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_multiAzNotInVpc(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_MultiAZNotInVPC_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_MultiAZNotInVPC_AvailabilityZones(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zones.0", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zones.1", "data.aws_availability_zones.available", "names.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_multiAzInVpc(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupMultiAZInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s\\..+\\.%s", aws.StringValue(rg.ReplicationGroupId), testAccGetPartitionDNSSuffix())))(s)
					},
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "reader_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s-ro\\..+\\.%s", aws.StringValue(rg.ReplicationGroupId), testAccGetPartitionDNSSuffix())))(s)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_multiAz_NoAutomaticFailover(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheReplicationGroupConfig_MultiAZ_NoAutomaticFailover(rName),
				ExpectError: regexp.MustCompile("automatic_failover_enabled must be true if multi_az_enabled is true"),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_AutomaticFailover_OneCacheCluster(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheReplicationGroupConfig_MultiAZOneCacheCluster_SingleNodeGroup(rName),
				ExpectError: regexp.MustCompile(`if automatic_failover_enabled is true, number_cache_clusters must be greater than 1`),
			},
			{
				Config:      testAccAWSElasticacheReplicationGroupConfig_MultiAZOneCacheCluster_ClusterMode(rName),
				ExpectError: regexp.MustCompile(`if automatic_failover_enabled is true, cluster_mode\[0\].replicas_per_node_group must be greater than 0`),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_redisClusterInVpc2(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupRedisClusterInVPCConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_endpoint_address"),
					resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s\\..+\\.%s", rName, testAccGetPartitionDNSSuffix()))),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint_address"),
					resource.TestMatchResourceAttr(resourceName, "reader_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s-ro\\..+\\.%s", rName, testAccGetPartitionDNSSuffix()))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_Basic(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_NonClusteredParameterGroup(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig_NonClusteredParameterGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexp.MustCompile(fmt.Sprintf("%s\\..+\\.%s", rName, testAccGetPartitionDNSSuffix()))),
					resource.TestCheckNoResourceAttr(resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_UpdateNumNodeGroups_ScaleUp(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 3, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_UpdateNumNodeGroups_ScaleDown(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 3, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_UpdateReplicasPerNodeGroup(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "3"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "8"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "8"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "2"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_UpdateNumNodeGroupsAndReplicasPerNodeGroup_ScaleUp(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "2"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "9"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "9"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_UpdateNumNodeGroupsAndReplicasPerNodeGroup_ScaleDown(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 3, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "2"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "9"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "9"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_clusteringAndCacheNodesCausesError(t *testing.T) {
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheReplicationGroupNativeRedisClusterErrorConfig(rInt, rName),
				ExpectError: regexp.MustCompile("only one of `cluster_mode,number_cache_clusters` can be\\s+specified, but `cluster_mode,number_cache_clusters` were specified."),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_enableSnapshotting(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "snapshot_retention_limit", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigEnableSnapshotting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "snapshot_retention_limit", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_enableAuthTokenTransitEncryption(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroup_EnableAuthTokenTransitEncryptionConfig(acctest.RandInt(), acctest.RandString(10), acctest.RandString(16)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "transit_encryption_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "auth_token", "availability_zones"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_enableAtRestEncryption(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroup_EnableAtRestEncryptionConfig(acctest.RandInt(), acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(
						resourceName, "at_rest_encryption_enabled", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "availability_zones"},
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_useCmkKmsKeyId(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroup_UseCmkKmsKeyId(acctest.RandInt(), acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttrSet("aws_elasticache_replication_group.bar", "kms_key_id"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_Basic(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_Failover_AutoFailoverDisabled(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := false
	multiAZEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_FailoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"},
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					timeout := 40 * time.Minute

					if err := resourceAwsElasticacheReplicationGroupSetPrimaryClusterID(conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_FailoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_Failover_AutoFailoverEnabled(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := true
	multiAZEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_FailoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					timeout := 40 * time.Minute

					// Must disable automatic failover first
					if err := resourceAwsElasticacheReplicationGroupDisableAutomaticFailover(conn, rName, timeout); err != nil {
						t.Fatalf("error disabling automatic failover: %s", err)
					}

					// Set primary
					if err := resourceAwsElasticacheReplicationGroupSetPrimaryClusterID(conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}

					// Re-enable automatic failover like nothing ever happened
					if err := resourceAwsElasticacheReplicationGroupEnableAutomaticFailover(conn, rName, multiAZEnabled, timeout); err != nil {
						t.Fatalf("error re-enabling automatic failover: %s", err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_FailoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_MultiAZEnabled(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := true
	multiAZEnabled := true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_FailoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					timeout := 40 * time.Minute

					// Must disable automatic failover first
					if err := resourceAwsElasticacheReplicationGroupDisableAutomaticFailover(conn, rName, timeout); err != nil {
						t.Fatalf("error disabling automatic failover: %s", err)
					}

					// Set primary
					if err := resourceAwsElasticacheReplicationGroupSetPrimaryClusterID(conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}

					// Re-enable automatic failover like nothing ever happened
					if err := resourceAwsElasticacheReplicationGroupEnableAutomaticFailover(conn, rName, multiAZEnabled, timeout); err != nil {
						t.Fatalf("error re-enabling automatic failover: %s", err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_FailoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_MemberClusterDisappears_NoChange(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := deleteElasticacheCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := waiter.CacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_MemberClusterDisappears_AddMemberCluster(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := deleteElasticacheCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := waiter.CacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_MemberClusterDisappears_RemoveMemberCluster_AtTargetSize(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "3"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := deleteElasticacheCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := waiter.CacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_MemberClusterDisappears_RemoveMemberCluster_ScaleDown(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "4"),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := deleteElasticacheCacheCluster(conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := waiter.CacheClusterDeleted(conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_tags(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately"}, //not in the API
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_FinalSnapshot(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfigFinalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "final_snapshot_identifier", rName),
				),
			},
		},
	})
}

func testAccCheckAWSElasticacheReplicationGroupExists(n string, v *elasticache.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No replication group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
		rg, err := finder.ReplicationGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("ElastiCache error: %w", err)
		}

		*v = *rg

		return nil
	}
}

func testAccCheckAWSElasticacheReplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_replication_group" {
			continue
		}
		_, err := finder.ReplicationGroupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("ElastiCache Replication Group (%s) still exists", rs.Primary.ID)
	}
	return nil
}

func testAccAWSElasticacheReplicationGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfig_Uppercase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_replication_group" "test" {
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = 1
  port                          = 6379
  replication_group_description = "test description"
  replication_group_id          = %[1]q
  subnet_group_name             = aws_elasticache_subnet_group.test.name
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfigEnableSnapshotting(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"
  snapshot_retention_limit      = 2
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfigParameterGroupName(rName string, parameterGroupNameIndex int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  count = 2

  # We do not have a data source for "latest" ElastiCache family
  # so unfortunately we must hardcode this for now
  family = "redis6.x"

  name = "%[1]s-${count.index}"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }
}

resource "aws_elasticache_replication_group" "test" {
  apply_immediately             = true
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  parameter_group_name          = aws_elasticache_parameter_group.test.*.name[%[2]d]
  replication_group_description = "test description"
  replication_group_id          = %[1]q
}
`, rName, parameterGroupNameIndex)
}

func testAccAWSElasticacheReplicationGroupConfigUpdatedDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "updated description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = true
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfigUpdatedMaintenanceWindow(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "updated description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = true
  maintenance_window            = "wed:03:00-wed:06:00"
  snapshot_window               = "01:00-02:00"
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfigUpdatedNodeSize(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "updated description"
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = 2
  port                          = 6379
  apply_immediately             = true
}
`, rName)
}

var testAccAWSElasticacheReplicationGroupInVPCConfig = fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-in-vpc"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids  = [aws_subnet.test.id]
}

resource "aws_security_group" "test" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 1
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  auto_minor_version_upgrade    = false
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

func testAccAWSElasticacheReplicationGroupConfig_MultiAZNotInVPC_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 2
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = true
  multi_az_enabled              = true
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfig_MultiAZNotInVPC_AvailabilityZones(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 2
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = true
  multi_az_enabled              = true
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupMultiAZInVPCConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "terraform-testacc-elasticache-replication-group-multi-az-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-multi-az-in-vpc-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-multi-az-in-vpc-bar"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test2.id,
  ]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled    = true
  multi_az_enabled              = true
  snapshot_window               = "02:00-03:00"
  snapshot_retention_limit      = 7
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfig_MultiAZ_NoAutomaticFailover(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 1
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = false
  multi_az_enabled              = true
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfig_MultiAZOneCacheCluster_SingleNodeGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  number_cache_clusters         = 1
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = true
  multi_az_enabled              = true
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupConfig_MultiAZOneCacheCluster_ClusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  automatic_failover_enabled    = true
  multi_az_enabled              = true

  cluster_mode {
    num_node_groups         = 1
    replicas_per_node_group = 0
  }
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupRedisClusterInVPCConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-redis-cluster-in-vpc"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-redis-cluster-in-vpc-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-redis-cluster-in-vpc-bar"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test2.id,
  ]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = "2"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled    = false
  snapshot_window               = "02:00-03:00"
  snapshot_retention_limit      = 7
  engine_version                = "3.2.4"
  maintenance_window            = "thu:03:00-thu:04:00"
}
`, rName)
}

func testAccAWSElasticacheReplicationGroupNativeRedisClusterErrorConfig(rInt int, rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster-err"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-err-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-err-test"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  automatic_failover_enabled    = true

  cluster_mode {
    replicas_per_node_group = 1
    num_node_groups         = 2
  }

  number_cache_clusters = 3
}
`, rInt, rInt, rName)
}

func testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName string, numNodeGroups, replicasPerNodeGroup int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-test"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  automatic_failover_enabled    = true

  cluster_mode {
    num_node_groups         = %[2]d
    replicas_per_node_group = %[3]d
  }
}
`, rName, numNodeGroups, replicasPerNodeGroup)
}

func testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig_NonClusteredParameterGroup(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t2.medium"
  automatic_failover_enabled    = false

  parameter_group_name = "default.redis6.x"
  cluster_mode {
    num_node_groups         = 1
    replicas_per_node_group = 1
  }
}
`, rName))
}

func testAccAWSElasticacheReplicationGroup_UseCmkKmsKeyId(rInt int, rString string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "foo" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_elasticache_subnet_group" "bar" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.foo.id,
  ]
}

resource "aws_security_group" "bar" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.foo.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_kms_key" "bar" {
  description = "tf-test-cmk-kms-key-id"
}

resource "aws_elasticache_replication_group" "bar" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = "1"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.bar.name
  security_group_ids            = [aws_security_group.bar.id]
  parameter_group_name          = "default.redis3.2"
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  engine_version                = "3.2.6"
  at_rest_encryption_enabled    = true
  kms_key_id                    = aws_kms_key.bar.arn
}
`, rInt, rInt, rString)
}

func testAccAWSElasticacheReplicationGroup_EnableAtRestEncryptionConfig(rInt int, rString string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = "1"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  parameter_group_name          = "default.redis3.2"
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  engine_version                = "3.2.6"
  at_rest_encryption_enabled    = true
}
`, rInt, rInt, rString)
}

func testAccAWSElasticacheReplicationGroup_EnableAuthTokenTransitEncryptionConfig(rInt int, rString10 string, rString16 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-auth-token-transit-encryption"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-auth-token-transit-encryption"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = "tf-%s"
  replication_group_description = "test description"
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = "1"
  port                          = 6379
  subnet_group_name             = aws_elasticache_subnet_group.test.name
  security_group_ids            = [aws_security_group.test.id]
  parameter_group_name          = "default.redis3.2"
  availability_zones            = [data.aws_availability_zones.available.names[0]]
  engine_version                = "3.2.6"
  transit_encryption_enabled    = true
  auth_token                    = "%s"
}
`, rInt, rInt, rString10, rString16)
}

func testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName string, numberCacheClusters int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_replication_group" "test" {
  node_type                     = "cache.t2.micro"
  number_cache_clusters         = %[2]d
  replication_group_id          = %[1]q
  replication_group_description = "Terraform Acceptance Testing - number_cache_clusters"
  subnet_group_name             = aws_elasticache_subnet_group.test.name
}
`, rName, numberCacheClusters)
}

func testAccAWSElasticacheReplicationGroupConfig_FailoverMultiAZ(rName string, numberCacheClusters int, autoFailover, multiAZ bool) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_replication_group" "test" {
  # InvalidParameterCombination: Automatic failover is not supported for T1 and T2 cache node types.
  automatic_failover_enabled    = %[3]t
  multi_az_enabled              = %[4]t
  node_type                     = "cache.t3.medium"
  number_cache_clusters         = %[2]d
  replication_group_id          = "%[1]s"
  replication_group_description = "Terraform Acceptance Testing - number_cache_clusters"
  subnet_group_name             = aws_elasticache_subnet_group.test.name
}
`, rName, numberCacheClusters, autoFailover, multiAZ)
}

func testAccAWSElasticacheReplicationGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSElasticacheReplicationGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  apply_immediately             = true
  auto_minor_version_upgrade    = false
  maintenance_window            = "tue:06:30-tue:07:30"
  snapshot_window               = "01:00-02:00"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSElasticacheReplicationGroupConfigFinalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func resourceAwsElasticacheReplicationGroupDisableAutomaticFailover(conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) error {
	return resourceAwsElasticacheReplicationGroupModify(conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:       aws.String(replicationGroupID),
		ApplyImmediately:         aws.Bool(true),
		AutomaticFailoverEnabled: aws.Bool(false),
		MultiAZEnabled:           aws.Bool(false),
	})
}

func resourceAwsElasticacheReplicationGroupEnableAutomaticFailover(conn *elasticache.ElastiCache, replicationGroupID string, multiAZEnabled bool, timeout time.Duration) error {
	return resourceAwsElasticacheReplicationGroupModify(conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:       aws.String(replicationGroupID),
		ApplyImmediately:         aws.Bool(true),
		AutomaticFailoverEnabled: aws.Bool(true),
		MultiAZEnabled:           aws.Bool(multiAZEnabled),
	})
}

func resourceAwsElasticacheReplicationGroupSetPrimaryClusterID(conn *elasticache.ElastiCache, replicationGroupID, primaryClusterID string, timeout time.Duration) error {
	return resourceAwsElasticacheReplicationGroupModify(conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
		ApplyImmediately:   aws.Bool(true),
		PrimaryClusterId:   aws.String(primaryClusterID),
	})
}

func formatReplicationGroupClusterID(replicationGroupID string, clusterID int) string {
	return fmt.Sprintf("%s-%03d", replicationGroupID, clusterID)
}
