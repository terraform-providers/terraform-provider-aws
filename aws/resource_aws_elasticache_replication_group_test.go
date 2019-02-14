package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_elasticache_replication_group", &resource.Sweeper{
		Name: "aws_elasticache_replication_group",
		F:    testSweepElasticacheReplicationGroups,
	})
}

func testSweepElasticacheReplicationGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).elasticacheconn

	prefixes := []string{
		"tf-",
		"tf-test-",
		"tf-acc-test-",
	}

	err = conn.DescribeReplicationGroupsPages(&elasticache.DescribeReplicationGroupsInput{}, func(page *elasticache.DescribeReplicationGroupsOutput, isLast bool) bool {
		if len(page.ReplicationGroups) == 0 {
			log.Print("[DEBUG] No Elasticache Replicaton Groups to sweep")
			return false
		}

		for _, replicationGroup := range page.ReplicationGroups {
			id := aws.StringValue(replicationGroup.ReplicationGroupId)
			skip := true
			for _, prefix := range prefixes {
				if strings.HasPrefix(id, prefix) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping Elasticache Replication Group: %s", id)
				continue
			}
			log.Printf("[INFO] Deleting Elasticache Replication Group: %s", id)
			err := deleteElasticacheReplicationGroup(id, conn)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Elasticache Replication Group (%s): %s", id, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticache Replication Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticache Replication Groups: %s", err)
	}
	return nil
}

func TestAccAWSElasticacheReplicationGroup_importBasic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	name := acctest.RandString(10)

	resourceName := "aws_elasticache_replication_group.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(name),
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

func TestAccAWSElasticacheReplicationGroup_basic(t *testing.T) {
	var rg elasticache.ReplicationGroup

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "cluster_mode.#", "0"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "member_clusters.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_Uppercase(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rStr := acctest.RandString(5)
	rgName := fmt.Sprintf("TF-ELASTIRG-%s", rStr)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_Uppercase(rgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "replication_group_id", fmt.Sprintf("tf-elastirg-%s", rStr)),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_updateDescription(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "replication_group_description", "test description"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "auto_minor_version_upgrade", "false"),
				),
			},

			{
				Config: testAccAWSElasticacheReplicationGroupConfigUpdatedDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "replication_group_description", "updated description"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "auto_minor_version_upgrade", "true"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_updateMaintenanceWindow(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "maintenance_window", "tue:06:30-tue:07:30"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfigUpdatedMaintenanceWindow(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "maintenance_window", "wed:03:00-wed:06:00"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_updateNodeSize(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "node_type", "cache.m1.small"),
				),
			},

			{
				Config: testAccAWSElasticacheReplicationGroupConfigUpdatedNodeSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "node_type", "cache.m1.medium"),
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
	rName := acctest.RandString(10)

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
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "1"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "auto_minor_version_upgrade", "false"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_multiAzInVpc(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupMultiAZInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(
						"aws_elasticache_replication_group.bar", "primary_endpoint_address"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_redisClusterInVpc2(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupRedisClusterInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(
						"aws_elasticache_replication_group.bar", "primary_endpoint_address"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_Basic(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandString(10)
	resourceName := "aws_elasticache_replication_group.bar"

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
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint_address"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_NumNodeGroups(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandString(10)
	resourceName := "aws_elasticache_replication_group.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 3, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 1, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.num_node_groups", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode.0.replicas_per_node_group", "1"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_clusteringAndCacheNodesCausesError(t *testing.T) {
	rInt := acctest.RandInt()
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheReplicationGroupNativeRedisClusterErrorConfig(rInt, rName),
				ExpectError: regexp.MustCompile("Either `number_cache_clusters` or `cluster_mode` must be set"),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_enableSnapshotting(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_retention_limit", "0"),
				),
			},

			{
				Config: testAccAWSElasticacheReplicationGroupConfigEnableSnapshotting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_retention_limit", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_enableAuthTokenTransitEncryption(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroup_EnableAuthTokenTransitEncryptionConfig(acctest.RandInt(), acctest.RandString(10), acctest.RandString(16)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "transit_encryption_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_enableAtRestEncryption(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroup_EnableAtRestEncryptionConfig(acctest.RandInt(), acctest.RandString(10)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "at_rest_encryption_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(4))
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 4, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "4"),
				),
			},
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_Failover_AutoFailoverDisabled(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(4))
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 3, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
				),
			},
			{
				PreConfig: func() {
					// Simulate failover so primary is on node we are trying to delete
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
					input := &elasticache.ModifyReplicationGroupInput{
						ApplyImmediately:   aws.Bool(true),
						PrimaryClusterId:   aws.String(fmt.Sprintf("%s-003", rName)),
						ReplicationGroupId: aws.String(rName),
					}
					if _, err := conn.ModifyReplicationGroup(input); err != nil {
						t.Fatalf("error setting new primary cache cluster: %s", err)
					}
					if err := waitForModifyElasticacheReplicationGroup(conn, rName, 40*time.Minute); err != nil {
						t.Fatalf("error waiting for new primary cache cluster: %s", err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_NumberCacheClusters_Failover_AutoFailoverEnabled(t *testing.T) {
	var replicationGroup elasticache.ReplicationGroup
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(4))
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 3, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "3"),
				),
			},
			{
				PreConfig: func() {
					// Simulate failover so primary is on node we are trying to delete
					conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

					// Must disable automatic failover first
					var input *elasticache.ModifyReplicationGroupInput = &elasticache.ModifyReplicationGroupInput{
						ApplyImmediately:         aws.Bool(true),
						AutomaticFailoverEnabled: aws.Bool(false),
						ReplicationGroupId:       aws.String(rName),
					}
					if _, err := conn.ModifyReplicationGroup(input); err != nil {
						t.Fatalf("error disabling automatic failover: %s", err)
					}
					if err := waitForModifyElasticacheReplicationGroup(conn, rName, 40*time.Minute); err != nil {
						t.Fatalf("error waiting for disabling automatic failover: %s", err)
					}

					// Failover
					input = &elasticache.ModifyReplicationGroupInput{
						ApplyImmediately:   aws.Bool(true),
						PrimaryClusterId:   aws.String(fmt.Sprintf("%s-003", rName)),
						ReplicationGroupId: aws.String(rName),
					}
					if _, err := conn.ModifyReplicationGroup(input); err != nil {
						t.Fatalf("error setting new primary cache cluster: %s", err)
					}
					if err := waitForModifyElasticacheReplicationGroup(conn, rName, 40*time.Minute); err != nil {
						t.Fatalf("error waiting for new primary cache cluster: %s", err)
					}

					// Re-enable automatic failover like nothing ever happened
					input = &elasticache.ModifyReplicationGroupInput{
						ApplyImmediately:         aws.Bool(true),
						AutomaticFailoverEnabled: aws.Bool(true),
						ReplicationGroupId:       aws.String(rName),
					}
					if _, err := conn.ModifyReplicationGroup(input); err != nil {
						t.Fatalf("error enabled automatic failover: %s", err)
					}
					if err := waitForModifyElasticacheReplicationGroup(conn, rName, 40*time.Minute); err != nil {
						t.Fatalf("error waiting for enabled automatic failover: %s", err)
					}
				},
				Config: testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName, 2, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "number_cache_clusters", "2"),
				),
			},
		},
	})
}

func TestResourceAWSElastiCacheReplicationGroupIdValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 0,
		},
		{
			Value:    "t.sting",
			ErrCount: 1,
		},
		{
			Value:    "t--sting",
			ErrCount: 1,
		},
		{
			Value:    "1testing",
			ErrCount: 1,
		},
		{
			Value:    "testing-",
			ErrCount: 1,
		},
		{
			Value:    randomString(21),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateAwsElastiCacheReplicationGroupId(tc.Value, "aws_elasticache_replication_group_replication_group_id")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the ElastiCache Replication Group Id to trigger a validation error")
		}
	}
}

func TestResourceAWSElastiCacheReplicationGroupEngineValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "Redis",
			ErrCount: 0,
		},
		{
			Value:    "REDIS",
			ErrCount: 0,
		},
		{
			Value:    "memcached",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateAwsElastiCacheReplicationGroupEngine(tc.Value, "aws_elasticache_replication_group_engine")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the ElastiCache Replication Group Engine to trigger a validation error")
		}
	}
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
		res, err := conn.DescribeReplicationGroups(&elasticache.DescribeReplicationGroupsInput{
			ReplicationGroupId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Elasticache error: %v", err)
		}

		for _, rg := range res.ReplicationGroups {
			if *rg.ReplicationGroupId == rs.Primary.ID {
				*v = *rg
			}
		}

		return nil
	}
}

func testAccCheckAWSElasticacheReplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_replication_group" {
			continue
		}
		res, err := conn.DescribeReplicationGroups(&elasticache.DescribeReplicationGroupsInput{
			ReplicationGroupId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			// Verify the error is what we want
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ReplicationGroupNotFoundFault" {
				continue
			}
			return err
		}
		if len(res.ReplicationGroups) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccAWSElasticacheReplicationGroupConfig(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "test description"
    node_type = "cache.m1.small"
    number_cache_clusters = 2
    port = 6379
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
    auto_minor_version_upgrade = false
    maintenance_window = "tue:06:30-tue:07:30"
    snapshot_window = "01:00-02:00"
}`, rName, rName, rName)
}

func testAccAWSElasticacheReplicationGroupConfig_Uppercase(rgName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "bar" {
  replication_group_id = "%s"
  replication_group_description = "test description"
  node_type = "cache.t2.micro"
  number_cache_clusters = 1
  port = 6379
}`, rgName)
}

func testAccAWSElasticacheReplicationGroupConfigEnableSnapshotting(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "test description"
    node_type = "cache.m1.small"
    number_cache_clusters = 2
    port = 6379
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
    auto_minor_version_upgrade = false
    maintenance_window = "tue:06:30-tue:07:30"
    snapshot_window = "01:00-02:00"
    snapshot_retention_limit = 2
}`, rName, rName, rName)
}

func testAccAWSElasticacheReplicationGroupConfigParameterGroupName(rName string, parameterGroupNameIndex int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  count = 2

  # We do not have a data source for "latest" Elasticache family
  # so unfortunately we must hardcode this for now
  family = "redis5.0"
  name   = "tf-%s-${count.index}"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }
}

resource "aws_elasticache_replication_group" "test" {
  apply_immediately             = true
  node_type                     = "cache.m1.small"
  number_cache_clusters         = 2
  parameter_group_name          = "${aws_elasticache_parameter_group.test.*.name[%d]}"
  replication_group_description = "test description"
  replication_group_id          = "tf-%s"
}
`, rName, parameterGroupNameIndex, rName)
}

func testAccAWSElasticacheReplicationGroupConfigUpdatedDescription(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "updated description"
    node_type = "cache.m1.small"
    number_cache_clusters = 2
    port = 6379
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
    auto_minor_version_upgrade = true
}`, rName, rName, rName)
}

func testAccAWSElasticacheReplicationGroupConfigUpdatedMaintenanceWindow(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "updated description"
    node_type = "cache.m1.small"
    number_cache_clusters = 2
    port = 6379
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
    auto_minor_version_upgrade = true
    maintenance_window = "wed:03:00-wed:06:00"
    snapshot_window = "01:00-02:00"
}`, rName, rName, rName)
}

func testAccAWSElasticacheReplicationGroupConfigUpdatedNodeSize(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%s"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "updated description"
    node_type = "cache.m1.medium"
    number_cache_clusters = 2
    port = 6379
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
}`, rName, rName, rName)
}

var testAccAWSElasticacheReplicationGroupInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
  tags = {
        Name = "terraform-testacc-elasticache-replication-group-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-elasticache-replication-group-in-vpc"
    }
}

resource "aws_elasticache_subnet_group" "bar" {
    name = "tf-test-cache-subnet-%03d"
    description = "tf-test-cache-subnet-group-descr"
    subnet_ids = ["${aws_subnet.foo.id}"]
}

resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    vpc_id = "${aws_vpc.foo.id}"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "test description"
    node_type = "cache.m1.small"
    number_cache_clusters = 1
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    availability_zones = ["us-west-2a"]
    auto_minor_version_upgrade = false
}

`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheReplicationGroupMultiAZInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
  tags = {
        Name = "terraform-testacc-elasticache-replication-group-multi-az-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-elasticache-replication-group-multi-az-in-vpc-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-elasticache-replication-group-multi-az-in-vpc-bar"
    }
}

resource "aws_elasticache_subnet_group" "bar" {
    name = "tf-test-cache-subnet-%03d"
    description = "tf-test-cache-subnet-group-descr"
    subnet_ids = [
        "${aws_subnet.foo.id}",
        "${aws_subnet.bar.id}"
    ]
}

resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    vpc_id = "${aws_vpc.foo.id}"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "test description"
    node_type = "cache.m1.small"
    number_cache_clusters = 2
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    availability_zones = ["us-west-2a","us-west-2b"]
    automatic_failover_enabled = true
    snapshot_window = "02:00-03:00"
    snapshot_retention_limit = 7
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheReplicationGroupRedisClusterInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
  tags = {
        Name = "terraform-testacc-elasticache-replication-group-redis-cluster-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-elasticache-replication-group-redis-cluster-in-vpc-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-elasticache-replication-group-redis-cluster-in-vpc-bar"
    }
}

resource "aws_elasticache_subnet_group" "bar" {
    name = "tf-test-cache-subnet-%03d"
    description = "tf-test-cache-subnet-group-descr"
    subnet_ids = [
        "${aws_subnet.foo.id}",
        "${aws_subnet.bar.id}"
    ]
}

resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    vpc_id = "${aws_vpc.foo.id}"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "test description"
    node_type = "cache.m3.medium"
    number_cache_clusters = "2"
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    availability_zones = ["us-west-2a","us-west-2b"]
    automatic_failover_enabled = false
    snapshot_window = "02:00-03:00"
    snapshot_retention_limit = 7
    engine_version = "3.2.4"
    maintenance_window = "thu:03:00-thu:04:00"
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

func testAccAWSElasticacheReplicationGroupNativeRedisClusterErrorConfig(rInt int, rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
  tags = {
        Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster-err"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-elasticache-replication-group-native-redis-cluster-err-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-elasticache-replication-group-native-redis-cluster-err-bar"
    }
}

resource "aws_elasticache_subnet_group" "bar" {
    name = "tf-test-cache-subnet-%03d"
    description = "tf-test-cache-subnet-group-descr"
    subnet_ids = [
        "${aws_subnet.foo.id}",
        "${aws_subnet.bar.id}"
    ]
}

resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    vpc_id = "${aws_vpc.foo.id}"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "test description"
    node_type = "cache.t2.micro"
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    automatic_failover_enabled = true
    cluster_mode {
      replicas_per_node_group = 1
      num_node_groups = 2
    }
    number_cache_clusters = 3
}`, rInt, rInt, rName)
}

func testAccAWSElasticacheReplicationGroupNativeRedisClusterConfig(rName string, numNodeGroups, replicasPerNodeGroup int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
  tags = {
        Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-elasticache-replication-group-native-redis-cluster-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-elasticache-replication-group-native-redis-cluster-bar"
    }
}

resource "aws_elasticache_subnet_group" "bar" {
    name = "tf-test-%[1]s"
    description = "tf-test-cache-subnet-group-descr"
    subnet_ids = [
        "${aws_subnet.foo.id}",
        "${aws_subnet.bar.id}"
    ]
}

resource "aws_security_group" "bar" {
    name = "tf-test-%[1]s"
    description = "tf-test-security-group-descr"
    vpc_id = "${aws_vpc.foo.id}"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%[1]s"
    replication_group_description = "test description"
    node_type = "cache.t2.micro"
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    automatic_failover_enabled = true
    cluster_mode {
      num_node_groups         = %d
      replicas_per_node_group = %d
    }
}`, rName, numNodeGroups, replicasPerNodeGroup)
}

func testAccAWSElasticacheReplicationGroup_EnableAtRestEncryptionConfig(rInt int, rString string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "terraform-testacc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "192.168.0.0/20"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_elasticache_subnet_group" "bar" {
  name = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    "${aws_subnet.foo.id}",
  ]
}

resource "aws_security_group" "bar" {
  name = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id = "${aws_vpc.foo.id}"
  ingress {
    from_port = -1
    to_port = -1
    protocol = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "bar" {
  replication_group_id = "tf-%s"
  replication_group_description = "test description"
  node_type = "cache.t2.micro"
  number_cache_clusters = "1"
  port = 6379
  subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
  security_group_ids = ["${aws_security_group.bar.id}"]
  parameter_group_name = "default.redis3.2"
  availability_zones = ["us-west-2a"]
  engine_version = "3.2.6"
  at_rest_encryption_enabled = true
}
`, rInt, rInt, rString)
}

func testAccAWSElasticacheReplicationGroup_EnableAuthTokenTransitEncryptionConfig(rInt int, rString10 string, rString16 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "192.168.0.0/16"
  tags = {
    Name = "terraform-testacc-elasticache-replication-group-auth-token-transit-encryption"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "192.168.0.0/20"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-elasticache-replication-group-auth-token-transit-encryption"
  }
}

resource "aws_elasticache_subnet_group" "bar" {
  name = "tf-test-cache-subnet-%03d"
  description = "tf-test-cache-subnet-group-descr"
  subnet_ids = [
    "${aws_subnet.foo.id}",
  ]
}

resource "aws_security_group" "bar" {
  name = "tf-test-security-group-%03d"
  description = "tf-test-security-group-descr"
  vpc_id = "${aws_vpc.foo.id}"
  ingress {
    from_port = -1
    to_port = -1
    protocol = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "bar" {
  replication_group_id = "tf-%s"
  replication_group_description = "test description"
  node_type = "cache.t2.micro"
  number_cache_clusters = "1"
  port = 6379
  subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
  security_group_ids = ["${aws_security_group.bar.id}"]
  parameter_group_name = "default.redis3.2"
  availability_zones = ["us-west-2a"]
  engine_version = "3.2.6"
  transit_encryption_enabled = true
  auth_token = "%s"
}
`, rInt, rInt, rString10, rString16)
}

func testAccAWSElasticacheReplicationGroupConfig_NumberCacheClusters(rName string, numberCacheClusters int, autoFailover bool) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"
  tags = {
      Name = "terraform-testacc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "192.168.${count.index}.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-elasticache-replication-group-number-cache-clusters"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
}

resource "aws_elasticache_replication_group" "test" {
  # InvalidParameterCombination: Automatic failover is not supported for T1 and T2 cache node types.
  automatic_failover_enabled    = %[2]t
  node_type                     = "cache.m3.medium"
  number_cache_clusters         = %[3]d
  replication_group_id          = "%[1]s"
  replication_group_description = "Terraform Acceptance Testing - number_cache_clusters"
  subnet_group_name             = "${aws_elasticache_subnet_group.test.name}"
}`, rName, autoFailover, numberCacheClusters)
}
