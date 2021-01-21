package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEcsDataSource_ecsCluster(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsClusterDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "pending_tasks_count", "0"),
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "running_tasks_count", "0"),
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "registered_container_instances_count", "0"),
					resource.TestCheckResourceAttrSet("data.aws_ecs_cluster.default", "arn"),
				),
			},
		},
	})
}

func TestAccAWSEcsDataSource_ecsClusterContainerInsights(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsClusterDataSourceConfigContainerInsights,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "pending_tasks_count", "0"),
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "running_tasks_count", "0"),
					resource.TestCheckResourceAttr("data.aws_ecs_cluster.default", "registered_container_instances_count", "0"),
					resource.TestCheckResourceAttrSet("data.aws_ecs_cluster.default", "arn"),
					resource.TestCheckResourceAttrPair("data.aws_ecs_cluster.default", "setting.#", "aws_ecs_cluster.default", "setting.#"),
				),
			},
		},
	})
}

var testAccCheckAwsEcsClusterDataSourceConfig = fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "default-%d"
}

data "aws_ecs_cluster" "default" {
  cluster_name = aws_ecs_cluster.default.name
}
`, acctest.RandInt())

var testAccCheckAwsEcsClusterDataSourceConfigContainerInsights = fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "default-%d"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

data "aws_ecs_cluster" "default" {
  cluster_name = aws_ecs_cluster.default.name
}
`, acctest.RandInt())
