package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEcsServiceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ecs_service.test"
	resourceName := "aws_ecs_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsServiceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "desired_count", dataSourceName, "desired_count"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_type", dataSourceName, "launch_type"),
					resource.TestCheckResourceAttrPair(resourceName, "scheduling_strategy", dataSourceName, "scheduling_strategy"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "service_name"),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", dataSourceName, "task_definition"),
				),
			},
		},
	})
}

var testAccCheckAwsEcsServiceDataSourceConfig = fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = "tf-acc-%d"
}

resource "aws_ecs_task_definition" "test" {
  family = "mongodb"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "memoryReservation": 64,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name = "mongodb"
  cluster = "${aws_ecs_cluster.test.id}"
  task_definition = "${aws_ecs_task_definition.test.arn}"
  desired_count = 1
}

data "aws_ecs_service" "test" {
  service_name = "${aws_ecs_service.test.name}"
  cluster_arn = "${aws_ecs_cluster.test.arn}"
}
`, acctest.RandInt())
