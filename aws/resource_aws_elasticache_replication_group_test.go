package aws

import (
	"fmt"
	"log"
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
			err := deleteElasticacheReplicationGroup(id, 40*time.Minute, conn)
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

func TestAccAWSElasticacheReplicationGroup_basic(t *testing.T) {
	var rg elasticache.ReplicationGroup

	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSElasticacheReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElasticacheReplicationGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "parameter_group_name", "default.redis3.2"),
				),
			},

			{
				Config: testAccAWSElasticacheReplicationGroupConfigUpdatedParameterGroup(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSElasticacheReplicationGroupExists("aws_elasticache_replication_group.bar", &rg),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "parameter_group_name", fmt.Sprintf("allkeys-lru-%d", rInt)),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_vpc(t *testing.T) {
	var rg elasticache.ReplicationGroup
	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
						"aws_elasticache_replication_group.bar", "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(
						"aws_elasticache_replication_group.bar", "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(
						"aws_elasticache_replication_group.bar", "configuration_endpoint_address"),
				),
			},
		},
	})
}

func TestAccAWSElasticacheReplicationGroup_ClusterMode_Basic(t *testing.T) {
	var rg elasticache.ReplicationGroup
	rName := acctest.RandString(10)
	resourceName := "aws_elasticache_replication_group.bar"

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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

	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
	resource.Test(t, resource.TestCase{
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
    parameter_group_name = "default.redis3.2"
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
    parameter_group_name = "default.redis3.2"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
    auto_minor_version_upgrade = false
    maintenance_window = "tue:06:30-tue:07:30"
    snapshot_window = "01:00-02:00"
    snapshot_retention_limit = 2
}`, rName, rName, rName)
}

func testAccAWSElasticacheReplicationGroupConfigUpdatedParameterGroup(rName string, rInt int) string {
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

resource "aws_elasticache_parameter_group" "bar" {
    name = "allkeys-lru-%d"
    family = "redis3.2"

    parameter {
        name = "maxmemory-policy"
        value = "allkeys-lru"
    }
}

resource "aws_elasticache_replication_group" "bar" {
    replication_group_id = "tf-%s"
    replication_group_description = "test description"
    node_type = "cache.m1.small"
    number_cache_clusters = 2
    port = 6379
    parameter_group_name = "${aws_elasticache_parameter_group.bar.name}"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
}`, rName, rName, rInt, rName)
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
    parameter_group_name = "default.redis3.2"
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
    parameter_group_name = "default.redis3.2"
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
    parameter_group_name = "default.redis3.2"
    security_group_names = ["${aws_elasticache_security_group.bar.name}"]
    apply_immediately = true
}`, rName, rName, rName)
}

var testAccAWSElasticacheReplicationGroupInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
    tags {
        Name = "terraform-testacc-elasticache-replication-group-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
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
    parameter_group_name = "default.redis3.2"
    availability_zones = ["us-west-2a"]
    auto_minor_version_upgrade = false
}

`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheReplicationGroupMultiAZInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
    tags {
        Name = "terraform-testacc-elasticache-replication-group-multi-az-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "tf-acc-elasticache-replication-group-multi-az-in-vpc-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
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
    parameter_group_name = "default.redis3.2"
    availability_zones = ["us-west-2a","us-west-2b"]
    automatic_failover_enabled = true
    snapshot_window = "02:00-03:00"
    snapshot_retention_limit = 7
}
`, acctest.RandInt(), acctest.RandInt(), acctest.RandString(10))

var testAccAWSElasticacheReplicationGroupRedisClusterInVPCConfig = fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "192.168.0.0/16"
    tags {
        Name = "terraform-testacc-elasticache-replication-group-redis-cluster-in-vpc"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "tf-acc-elasticache-replication-group-redis-cluster-in-vpc-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
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
    node_type = "cache.t2.micro"
    number_cache_clusters = "2"
    port = 6379
    subnet_group_name = "${aws_elasticache_subnet_group.bar.name}"
    security_group_ids = ["${aws_security_group.bar.id}"]
    parameter_group_name = "default.redis3.2.cluster.on"
    availability_zones = ["us-west-2a","us-west-2b"]
    automatic_failover_enabled = true
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
    tags {
        Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster-err"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "tf-acc-elasticache-replication-group-native-redis-cluster-err-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
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
    parameter_group_name = "default.redis3.2.cluster.on"
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
    tags {
        Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster"
    }
}

resource "aws_subnet" "foo" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.0.0/20"
    availability_zone = "us-west-2a"
    tags {
        Name = "tf-acc-elasticache-replication-group-native-redis-cluster-foo"
    }
}

resource "aws_subnet" "bar" {
    vpc_id = "${aws_vpc.foo.id}"
    cidr_block = "192.168.16.0/20"
    availability_zone = "us-west-2b"
    tags {
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
    parameter_group_name = "default.redis3.2.cluster.on"
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
  tags {
    Name = "terraform-testacc-elasticache-replication-group-at-rest-encryption"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "192.168.0.0/20"
  availability_zone = "us-west-2a"
  tags {
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
  tags {
    Name = "terraform-testacc-elasticache-replication-group-auth-token-transit-encryption"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "192.168.0.0/20"
  availability_zone = "us-west-2a"
  tags {
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
