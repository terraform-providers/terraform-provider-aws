package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsElasticacheReplicationGroup_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "member_clusters.#", resourceName, "member_clusters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "number_cache_clusters", resourceName, "number_cache_clusters"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "primary_endpoint_address", resourceName, "primary_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "reader_endpoint_address", resourceName, "reader_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_description", resourceName, "replication_group_description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_id", resourceName, "replication_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_window", resourceName, "snapshot_window"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_ClusterMode(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_ClusterMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration_endpoint_address", resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_description", resourceName, "replication_group_description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_id", resourceName, "replication_group_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_MultiAZ(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_MultiAZ(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_NonExistent(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsElasticacheReplicationGroupConfig_NonExistent,
				ExpectError: regexp.MustCompile(`couldn't find resource`),
			},
		},
	})
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_basic(rName string) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  port                          = 6379
  availability_zones            = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled    = true
  snapshot_window               = "01:00-02:00"
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_ClusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  port                          = 6379
  automatic_failover_enabled    = true

  cluster_mode {
    replicas_per_node_group = 1
    num_node_groups         = 2
  }
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_MultiAZ(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test description"
  node_type                     = "cache.t3.small"
  number_cache_clusters         = 2
  automatic_failover_enabled    = true
  multi_az_enabled              = true
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

const testAccDataSourceAwsElasticacheReplicationGroupConfig_NonExistent = `
data "aws_elasticache_replication_group" "test" {
  replication_group_id = "tf-acc-test-nonexistent"
}
`
