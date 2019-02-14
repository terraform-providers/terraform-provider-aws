package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsElasticacheReplicationGroup_basic(t *testing.T) {
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "replication_group_id", fmt.Sprintf("tf-%s", rName)),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "replication_group_description", "test description"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "auth_token_enabled", "false"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "port", "6379"),
					resource.TestCheckResourceAttrSet("data.aws_elasticache_replication_group.bar", "primary_endpoint_address"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "number_cache_clusters", "2"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "member_clusters.#", "2"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "node_type", "cache.m1.small"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.bar", "snapshot_window", "01:00-02:00"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsElasticacheReplicationGroup_ClusterMode(t *testing.T) {
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsElasticacheReplicationGroupConfig_ClusterMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.cluster", "replication_group_id", fmt.Sprintf("tf-%s", rName)),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.cluster", "replication_group_description", "test description"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.cluster", "auth_token_enabled", "false"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.cluster", "automatic_failover_enabled", "true"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.cluster", "port", "6379"),
					resource.TestCheckResourceAttrSet("data.aws_elasticache_replication_group.cluster", "configuration_endpoint_address"),
					resource.TestCheckResourceAttr("data.aws_elasticache_replication_group.cluster", "node_type", "cache.m1.small"),
				),
			},
		},
	})
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_elasticache_replication_group" "bar" {
	replication_group_id = "tf-%s"
	replication_group_description = "test description"
	node_type = "cache.m1.small"
	number_cache_clusters = 2
	port = 6379
	availability_zones = ["${data.aws_availability_zones.available.names[0]}", "${data.aws_availability_zones.available.names[1]}"]
	automatic_failover_enabled = true
	snapshot_window = "01:00-02:00"
}

data "aws_elasticache_replication_group" "bar" {
	replication_group_id = "${aws_elasticache_replication_group.bar.replication_group_id}"
}`, rName)
}

func testAccDataSourceAwsElasticacheReplicationGroupConfig_ClusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "cluster" {
	replication_group_id = "tf-%s"
	replication_group_description = "test description"
	node_type = "cache.m1.small"
	port = 6379
	automatic_failover_enabled = true
	cluster_mode {
		replicas_per_node_group     = 1
		num_node_groups             = 2
	}
}

data "aws_elasticache_replication_group" "cluster" {
	replication_group_id = "${aws_elasticache_replication_group.cluster.replication_group_id}"
}`, rName)
}
