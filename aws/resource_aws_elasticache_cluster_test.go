package aws

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
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
	resource.AddTestSweepers("aws_elasticache_cluster", &resource.Sweeper{
		Name: "aws_elasticache_cluster",
		F:    testSweepElasticacheClusters,
		Dependencies: []string{
			"aws_elasticache_replication_group",
		},
	})
}

func testSweepElasticacheClusters(region string) error {
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

	err = conn.DescribeCacheClustersPages(&elasticache.DescribeCacheClustersInput{}, func(page *elasticache.DescribeCacheClustersOutput, isLast bool) bool {
		if len(page.CacheClusters) == 0 {
			log.Print("[DEBUG] No Elasticache Replicaton Groups to sweep")
			return false
		}

		for _, cluster := range page.CacheClusters {
			id := aws.StringValue(cluster.CacheClusterId)
			skip := true
			for _, prefix := range prefixes {
				if strings.HasPrefix(id, prefix) {
					skip = false
					break
				}
			}
			if skip {
				log.Printf("[INFO] Skipping Elasticache Cluster: %s", id)
				continue
			}
			log.Printf("[INFO] Deleting Elasticache Cluster: %s", id)
			err := deleteElasticacheCacheCluster(conn, id)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Elasticache Cache Cluster (%s): %s", id, err)
			}
			err = waitForDeleteElasticacheCacheCluster(conn, id, 40*time.Minute)
			if err != nil {
				log.Printf("[ERROR] Failed waiting for Elasticache Cache Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticache Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticache Clusters: %s", err)
	}
	return nil
}

func TestAccAWSElasticacheCluster_Engine_Memcached_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var ec elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_Engine_Memcached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "port", "11211"),
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

func TestAccAWSElasticacheCluster_Engine_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var ec elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_Engine_Redis(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttr(resourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(resourceName, "port", "6379"),
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

func TestAccAWSElasticacheCluster_ParameterGroupName_Default(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ParameterGroupName(rName, "memcached", "1.4.34", "default.memcached1.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "parameter_group_name", "default.memcached1.4"),
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

func TestAccAWSElasticacheCluster_Port_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var ec elasticache.CacheCluster
	port := 11212
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_Port(rName, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, "engine", "memcached"),
					resource.TestCheckResourceAttr(resourceName, "port", strconv.Itoa(port)),
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

func TestAccAWSElasticacheCluster_SecurityGroup(t *testing.T) {
	var ec elasticache.CacheCluster
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_SecurityGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet("aws_elasticache_cluster.bar", "configuration_endpoint"),
					resource.TestCheckResourceAttrSet("aws_elasticache_cluster.bar", "cluster_address"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_snapshotsWithUpdates(t *testing.T) {
	var ec elasticache.CacheCluster

	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccAWSElasticacheClusterConfig_snapshots, ri, ri, acctest.RandString(10))
	postConfig := fmt.Sprintf(testAccAWSElasticacheClusterConfig_snapshotsUpdated, ri, ri, acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_window", "05:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_retention_limit", "3"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSecurityGroupExists("aws_elasticache_security_group.bar"),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_window", "07:00-09:00"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "snapshot_retention_limit", "7"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_Decrease(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_Increase(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_IncreaseWithPreferredAvailabilityZones(t *testing.T) {
	var ec elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "1"),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", "1"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", "3"),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", "3"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_vpc(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSubnetGroupExists("aws_elasticache_subnet_group.bar", &csg),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					testAccCheckAWSElasticacheClusterAttributes(&ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "availability_zone", "us-west-2a"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_multiAZInVpc(t *testing.T) {
	var csg elasticache.CacheSubnetGroup
	var ec elasticache.CacheCluster
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterMultiAZInVPCConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheSubnetGroupExists("aws_elasticache_subnet_group.bar", &csg),
					testAccCheckAWSElasticacheClusterExists("aws_elasticache_cluster.bar", &ec),
					resource.TestCheckResourceAttr(
						"aws_elasticache_cluster.bar", "availability_zone", "Multiple"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_AZMode_Memcached_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var cluster elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_AZMode_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var cluster elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, "unknown"),
				ExpectError: regexp.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, "cross-az"),
				ExpectError: regexp.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_EngineVersion_Memcached_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, mid, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, "1.4.33"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.33"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, "1.4.24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &mid),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &mid),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.24"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, "1.4.34"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&mid, &post),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "1.4.34"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_EngineVersion_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, mid, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, "3.2.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.6"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, "3.2.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &mid),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &mid),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.4"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, "3.2.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&mid, &post),
					resource.TestCheckResourceAttr(resourceName, "engine_version", "3.2.10"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NodeTypeResize_Memcached_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, "cache.m3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.medium"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, "cache.m3.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.large"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NodeTypeResize_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var pre, post elasticache.CacheCluster
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resourceName := "aws_elasticache_cluster.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, "cache.m3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.medium"),
				),
			},
			{
				Config: testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, "cache.m3.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheClusterExists(resourceName, &post),
					testAccCheckAWSElasticacheClusterNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.m3.large"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_NumCacheNodes_Redis_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_NumCacheNodes_Redis_Ec2Classic(rName, 2),
				ExpectError: regexp.MustCompile(`engine "redis" does not support num_cache_nodes > 1`),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ReplicationGroupID_InvalidAttributes(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "availability_zones", "${list(\"us-east-1a\", \"us-east-1c\")}"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with availability_zones`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "az_mode", "single-az"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with az_mode`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "engine_version", "3.2.10"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with engine_version`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "engine", "redis"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with engine`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "maintenance_window", "sun:05:00-sun:09:00"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with maintenance_window`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "node_type", "cache.m3.medium"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with node_type`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "notification_topic_arn", "arn:aws:sns:us-east-1:123456789012:topic/non-existent"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with notification_topic_arn`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "num_cache_nodes", "1"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with num_cache_nodes`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "parameter_group_name", "non-existent"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with parameter_group_name`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "port", "6379"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with port`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "security_group_ids", "${list(\"sg-12345678\", \"sg-87654321\")}"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with security_group_ids`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "security_group_names", "${list(\"group1\", \"group2\")}"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with security_group_names`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "snapshot_arns", "${list(\"arn:aws:s3:::my_bucket/snapshot1.rdb\")}"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with snapshot_arns`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "snapshot_name", "arn:aws:s3:::my_bucket/snapshot1.rdb"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with snapshot_name`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "snapshot_retention_limit", "0"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with snapshot_retention_limit`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "snapshot_window", "05:00-09:00"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with snapshot_window`),
			},
			{
				Config:      testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, "subnet_group_name", "group1"),
				ExpectError: regexp.MustCompile(`"replication_group_id": conflicts with subnet_group_name`),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ReplicationGroupID_AvailabilityZone_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var cluster elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(7))
	clusterResourceName := "aws_elasticache_cluster.replica"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ReplicationGroupID_AvailabilityZone_Ec2Classic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(replicationGroupResourceName, &replicationGroup),
					testAccCheckAWSElasticacheClusterExists(clusterResourceName, &cluster),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ReplicationGroupID_SingleReplica_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var cluster elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(7))
	clusterResourceName := "aws_elasticache_cluster.replica"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ReplicationGroupID_Replica_Ec2Classic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(replicationGroupResourceName, &replicationGroup),
					testAccCheckAWSElasticacheClusterExists(clusterResourceName, &cluster),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName, "node_type", "cache.m3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName, "port", "6379"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheCluster_ReplicationGroupID_MultipleReplica_Ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var cluster1, cluster2 elasticache.CacheCluster
	var replicationGroup elasticache.ReplicationGroup
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(7))
	clusterResourceName1 := "aws_elasticache_cluster.replica.0"
	clusterResourceName2 := "aws_elasticache_cluster.replica.1"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheClusterConfig_ReplicationGroupID_Replica_Ec2Classic(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists(replicationGroupResourceName, &replicationGroup),
					testAccCheckAWSElasticacheClusterExists(clusterResourceName1, &cluster1),
					testAccCheckAWSElasticacheClusterExists(clusterResourceName2, &cluster2),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster1, &replicationGroup),
					testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(&cluster2, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName1, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName1, "node_type", "cache.m3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName1, "port", "6379"),
					resource.TestCheckResourceAttr(clusterResourceName2, "engine", "redis"),
					resource.TestCheckResourceAttr(clusterResourceName2, "node_type", "cache.m3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName2, "port", "6379"),
				),
			},
		},
	})
}

func testAccCheckAWSElasticacheClusterAttributes(v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.NotificationConfiguration == nil {
			return fmt.Errorf("Expected NotificationConfiguration for ElastiCache Cluster (%s)", *v.CacheClusterId)
		}

		if strings.ToLower(*v.NotificationConfiguration.TopicStatus) != "active" {
			return fmt.Errorf("Expected NotificationConfiguration status to be 'active', got (%s)", *v.NotificationConfiguration.TopicStatus)
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterReplicationGroupIDAttribute(cluster *elasticache.CacheCluster, replicationGroup *elasticache.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if cluster.ReplicationGroupId == nil {
			return errors.New("expected cluster ReplicationGroupId to be set")
		}

		if aws.StringValue(cluster.ReplicationGroupId) != aws.StringValue(replicationGroup.ReplicationGroupId) {
			return errors.New("expected cluster ReplicationGroupId to equal replication group ID")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterNotRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CacheClusterCreateTime) != aws.TimeValue(j.CacheClusterCreateTime) {
			return errors.New("Elasticache Cluster was recreated")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterRecreated(i, j *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CacheClusterCreateTime) == aws.TimeValue(j.CacheClusterCreateTime) {
			return errors.New("Elasticache Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckAWSElasticacheClusterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elasticacheconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_cluster" {
			continue
		}
		res, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			// Verify the error is what we want
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "CacheClusterNotFound" {
				continue
			}
			return err
		}
		if len(res.CacheClusters) > 0 {
			return fmt.Errorf("still exist.")
		}
	}
	return nil
}

func testAccCheckAWSElasticacheClusterExists(n string, v *elasticache.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No cache cluster ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elasticacheconn
		resp, err := conn.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
			CacheClusterId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return fmt.Errorf("Elasticache error: %v", err)
		}

		for _, c := range resp.CacheClusters {
			if *c.CacheClusterId == rs.Primary.ID {
				*v = *c
			}
		}

		return nil
	}
}

func testAccAWSElasticacheClusterConfig_Engine_Memcached(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  cluster_id           = "%s"
  engine               = "memcached"
  node_type            = "cache.m1.small"
  num_cache_nodes      = 1
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_Engine_Redis(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  cluster_id           = "%s"
  engine               = "redis"
  node_type            = "cache.m1.small"
  num_cache_nodes      = 1
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_ParameterGroupName(rName, engine, engineVersion, parameterGroupName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id           = %q
  engine               = %q
  engine_version       = %q
  node_type            = "cache.m1.small"
  num_cache_nodes      = 1
  parameter_group_name = %q
}
`, rName, engine, engineVersion, parameterGroupName)
}

func testAccAWSElasticacheClusterConfig_Port(rName string, port int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  cluster_id           = "%s"
  engine               = "memcached"
  node_type            = "cache.m1.small"
  num_cache_nodes      = 1
  port                 = %d
}
`, rName, port)
}

var testAccAWSElasticacheClusterConfig_SecurityGroup = fmt.Sprintf(`
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }

	tags = {
			Name = "TestAccAWSElasticacheCluster_basic"
		}
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "memcached"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 11211
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheClusterConfig_snapshots = `
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "redis"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 6379
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    snapshot_window = "05:00-09:00"
    snapshot_retention_limit = 3
}
`

var testAccAWSElasticacheClusterConfig_snapshotsUpdated = `
provider "aws" {
	region = "us-east-1"
}
resource "aws_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    ingress {
        from_port = -1
        to_port = -1
        protocol = "icmp"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

resource "aws_elasticache_security_group" "bar" {
    name = "tf-test-security-group-%03d"
    description = "tf-test-security-group-descr"
    security_group_names = ["${aws_security_group.bar.name}"]
}

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "redis"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    port = 6379
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    snapshot_window = "07:00-09:00"
    snapshot_retention_limit = 7
    apply_immediately = true
}
`

func testAccAWSElasticacheClusterConfig_NumCacheNodes(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%s"
  engine               = "memcached"
  node_type            = "cache.m1.small"
  num_cache_nodes      = %d
}
`, rName, numCacheNodes)
}

func testAccAWSElasticacheClusterConfig_NumCacheNodesWithPreferredAvailabilityZones(rName string, numCacheNodes int) string {
	preferredAvailabilityZones := make([]string, numCacheNodes)
	for i := range preferredAvailabilityZones {
		preferredAvailabilityZones[i] = `"${data.aws_availability_zones.available.names[0]}"`
	}

	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_elasticache_cluster" "bar" {
  apply_immediately            = true
  cluster_id                   = "%s"
  engine                       = "memcached"
  node_type                    = "cache.m1.small"
  num_cache_nodes              = %d
  preferred_availability_zones = [%s]
}
`, rName, numCacheNodes, strings.Join(preferredAvailabilityZones, ","))
}

var testAccAWSElasticacheClusterInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
  tags = {
        Name = "terraform-testacc-elasticache-cluster-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-elasticache-cluster-in-vpc"
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

resource "aws_elasticache_cluster" "bar" {
    // Including uppercase letters in this name to ensure
    // that we correctly handle the fact that the API
    // normalizes names to lowercase.
    cluster_id = "tf-%s"
    node_type = "cache.m1.small"
    num_cache_nodes = 1
    engine = "redis"
    engine_version = "2.8.19"
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    parameter_group_name = "default.redis2.8"
    notification_topic_arn      = "${aws_sns_topic.topic_example.arn}"
    availability_zone = "us-west-2a"
}

resource "aws_sns_topic" "topic_example" {
  name = "tf-ecache-cluster-test"
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheClusterMultiAZInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
  tags = {
        Name = "terraform-testacc-elasticache-cluster-multi-az-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-elasticache-cluster-multi-az-in-vpc-bar"
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

resource "aws_elasticache_cluster" "bar" {
    cluster_id = "tf-%s"
    engine = "memcached"
    node_type = "cache.m1.small"
    num_cache_nodes = 2
    port = 11211
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    az_mode = "cross-az"
    preferred_availability_zones = [
        "us-west-2a",
        "us-west-2b"
    ]
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

func testAccAWSElasticacheClusterConfig_AZMode_Memcached_Ec2Classic(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  az_mode              = "%[2]s"
  cluster_id           = "%[1]s"
  engine               = "memcached"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  port                 = 11211
}
`, rName, azMode)
}

func testAccAWSElasticacheClusterConfig_AZMode_Redis_Ec2Classic(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  az_mode              = "%[2]s"
  cluster_id           = "%[1]s"
  engine               = "redis"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  port                 = 6379
}
`, rName, azMode)
}

func testAccAWSElasticacheClusterConfig_EngineVersion_Memcached_Ec2Classic(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "memcached"
  engine_version       = "%[2]s"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  port                 = 11211
}
`, rName, engineVersion)
}

func testAccAWSElasticacheClusterConfig_EngineVersion_Redis_Ec2Classic(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "redis"
  engine_version       = "%[2]s"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = 1
  port                 = 6379
}
`, rName, engineVersion)
}

func testAccAWSElasticacheClusterConfig_NodeType_Memcached_Ec2Classic(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "memcached"
  node_type            = "%[2]s"
  num_cache_nodes      = 1
  port                 = 11211
}
`, rName, nodeType)
}

func testAccAWSElasticacheClusterConfig_NodeType_Redis_Ec2Classic(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "redis"
  node_type            = "%[2]s"
  num_cache_nodes      = 1
  port                 = 6379
}
`, rName, nodeType)
}

func testAccAWSElasticacheClusterConfig_NumCacheNodes_Redis_Ec2Classic(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "bar" {
  apply_immediately    = true
  cluster_id           = "%[1]s"
  engine               = "redis"
  node_type            = "cache.m3.medium"
  num_cache_nodes      = %[2]d
  port                 = 6379
}
`, rName, numCacheNodes)
}

func testAccAWSElasticacheClusterConfig_ReplicationGroupID_InvalidAttribute(rName, attrName, attrValue string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "replica" {
  cluster_id           = "%[1]s"
  replication_group_id = "non-existent-id"
  %[2]s                = "%[3]s"
}
`, rName, attrName, attrValue)
}

func testAccAWSElasticacheClusterConfig_ReplicationGroupID_AvailabilityZone_Ec2Classic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_elasticache_replication_group" "test" {
  replication_group_description  = "Terraform Acceptance Testing"
  replication_group_id           = "%[1]s"
  node_type                      = "cache.m3.medium"
  number_cache_clusters          = 1
  port                           = 6379

  lifecycle {
    ignore_changes = ["number_cache_clusters"]
  }
}

resource "aws_elasticache_cluster" "replica" {
  availability_zone    = "${data.aws_availability_zones.available.names[0]}"
  cluster_id           = "%[1]s1"
  replication_group_id = "${aws_elasticache_replication_group.test.id}"
}
`, rName)
}

func testAccAWSElasticacheClusterConfig_ReplicationGroupID_Replica_Ec2Classic(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_description  = "Terraform Acceptance Testing"
  replication_group_id           = "%[1]s"
  node_type                      = "cache.m3.medium"
  number_cache_clusters          = 1
  port                           = 6379

  lifecycle {
    ignore_changes = ["number_cache_clusters"]
  }
}

resource "aws_elasticache_cluster" "replica" {
  count = %[2]d

  cluster_id           = "%[1]s${count.index}"
  replication_group_id = "${aws_elasticache_replication_group.test.id}"
}
`, rName, count)
}
