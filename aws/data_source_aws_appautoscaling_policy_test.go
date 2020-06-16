package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSAppautoscalingPolicyDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_appautoscaling_policy.test"
	resourceName := "aws_appautoscaling_policy.test"
	rName := fmt.Sprintf("tf-app-policy-test-%s", acctest.RandString(4))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAppautoscalingPolicyDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_type", dataSourceName, "policy_type"),
					resource.TestCheckResourceAttrPair(resourceName, "resource_id", dataSourceName, "resource_id"),
					resource.TestCheckResourceAttrPair(resourceName, "scalable_dimension", dataSourceName, "scalable_dimension"),
					resource.TestCheckResourceAttrPair(resourceName, "service_namespace", dataSourceName, "service_namespace"),
				),
			},
		},
	})
}

func testAppautoscalingPolicyDataSourceConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
	name = %[1]q
  }
  
  resource "aws_ecs_task_definition" "test" {
	family = %[1]q
  
	container_definitions = <<EOF
  [
	{
	  "name": "busybox",
	  "image": "busybox:latest",
	  "cpu": 10,
	  "memory": 128,
	  "essential": true
	}
  ]
  EOF
  }
  
  resource "aws_ecs_service" "test" {
	cluster                            = "${aws_ecs_cluster.test.id}"
	deployment_maximum_percent         = 200
	deployment_minimum_healthy_percent = 50
	desired_count                      = 0
	name                               = %[1]q
	task_definition                    = "${aws_ecs_task_definition.test.arn}"
  }
  
  resource "aws_appautoscaling_target" "test" {
	max_capacity       = 4
	min_capacity       = 0
	resource_id        = "service/${aws_ecs_cluster.test.name}/${aws_ecs_service.test.name}"
	scalable_dimension = "ecs:service:DesiredCount"
	service_namespace  = "ecs"
  }
  
  resource "aws_appautoscaling_policy" "test" {
	name               = %[1]q
	resource_id        = "${aws_appautoscaling_target.test.resource_id}"
	scalable_dimension = "${aws_appautoscaling_target.test.scalable_dimension}"
	service_namespace  = "${aws_appautoscaling_target.test.service_namespace}"
  
	step_scaling_policy_configuration {
	  adjustment_type         = "ChangeInCapacity"
	  cooldown                = 60
	  metric_aggregation_type = "Average"
  
	  step_adjustment {
		metric_interval_lower_bound = 0
		scaling_adjustment          = 1
	  }
	}
  }

data "aws_appautoscaling_policy" "test" {
  name              = aws_appautoscaling_policy.test.name
  service_namespace = "ecs"
}  
`, r)
}
