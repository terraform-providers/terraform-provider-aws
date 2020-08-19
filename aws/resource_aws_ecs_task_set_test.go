package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsEcsTaskSet_basic(t *testing.T) {
	var taskSet ecs.TaskSet

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resourceName := "aws_ecs_task_set.mongo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSet(clusterName, tdName, svcName, 0.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					testAccCheckAwsEcsTaskSetArn(resourceName, clusterName, svcName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_withARN(t *testing.T) {
	var taskSet ecs.TaskSet

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resourceName := "aws_ecs_task_set.mongo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSet(clusterName, tdName, svcName, 0.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
				),
			},

			{
				Config: testAccAWSEcsTaskSetModified(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "external_id", "TEST_ID"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_disappears(t *testing.T) {
	var taskSet ecs.TaskSet

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resourceName := "aws_ecs_task_set.mongo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSet(clusterName, tdName, svcName, 0.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcsTaskSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEcsTaskSet_scale(t *testing.T) {
	var taskSet ecs.TaskSet

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resourceName := "aws_ecs_task_set.mongo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSet(clusterName, tdName, svcName, 0.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "scale.0.unit", "PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "scale.0.value", "0"),
				),
			},
			{
				Config: testAccAWSEcsTaskSet(clusterName, tdName, svcName, 100.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "scale.0.unit", "PERCENT"),
					resource.TestCheckResourceAttr(resourceName, "scale.0.value", "100"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_withCapacityProviderStrategy(t *testing.T) {
	var taskSet ecs.TaskSet

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	providerName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_task_set.mongo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSetWithCapacityProviderStrategy(providerName, clusterName, tdName, svcName, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
				),
			},
			{
				Config: testAccAWSEcsTaskSetWithCapacityProviderStrategy(providerName, clusterName, tdName, svcName, 10, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_withMultipleCapacityProviderStrategies(t *testing.T) {
	var taskSet ecs.TaskSet

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	sgName := acctest.RandomWithPrefix("tf-acc-sg")
	resourceName := "aws_ecs_task_set.mongo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSetWithMultipleCapacityProviderStrategies(clusterName, tdName, svcName, sgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_withAlb(t *testing.T) {
	var taskSet ecs.TaskSet

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	lbName := acctest.RandomWithPrefix("tf-acc-lb")
	resourceName := "aws_ecs_task_set.with_alb"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSetWithAlb(clusterName, tdName, lbName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_withLaunchTypeFargate(t *testing.T) {
	var taskSet ecs.TaskSet

	sg1Name := acctest.RandomWithPrefix("tf-acc-sg-1")
	sg2Name := acctest.RandomWithPrefix("tf-acc-sg-2")
	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resourceName := "aws_ecs_task_set.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSetWithLaunchTypeFargate(sg1Name, sg2Name, clusterName, tdName, svcName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.3.0"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_withLaunchTypeFargateAndPlatformVersion(t *testing.T) {
	var taskSet ecs.TaskSet

	sg1Name := acctest.RandomWithPrefix("tf-acc-sg-1")
	sg2Name := acctest.RandomWithPrefix("tf-acc-sg-2")
	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resourceName := "aws_ecs_task_set.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSetWithLaunchTypeFargateAndPlatformVersion(sg1Name, sg2Name, clusterName, tdName, svcName, "1.2.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.2.0"),
				),
			},
			{
				Config: testAccAWSEcsTaskSetWithLaunchTypeFargateAndPlatformVersion(sg1Name, sg2Name, clusterName, tdName, svcName, "1.3.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.3.0"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_withServiceRegistries(t *testing.T) {
	var taskSet ecs.TaskSet
	rString := acctest.RandString(8)

	clusterName := acctest.RandomWithPrefix("tf-acc-cluster")
	tdName := acctest.RandomWithPrefix("tf-acc-td")
	svcName := acctest.RandomWithPrefix("tf-acc-svc")
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSet_withServiceRegistries(rString, clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSEcsTaskSet_Tags(t *testing.T) {
	var taskSet ecs.TaskSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsTaskSetConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSEcsTaskSetConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEcsTaskSetConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsTaskSetExists(resourceName, &taskSet),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

//////////////
// Fixtures //
//////////////

func testAccAWSEcsTaskSet(clusterName, tdName, svcName string, scale float64) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}
resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}
resource "aws_ecs_service" "mongo" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.default.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "mongo" {
  service         = "${aws_ecs_service.mongo.id}"
  cluster         = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  scale {
    value = %f
  }
}
`, clusterName, tdName, svcName, scale)
}

func testAccAWSEcsTaskSetModified(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}
resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}
resource "aws_ecs_service" "mongo" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.default.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "mongo" {
  service         = "${aws_ecs_service.mongo.id}"
  cluster         = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  external_id     = "TEST_ID"
}
`, clusterName, tdName, svcName)
}

func testAccAWSEcsTaskSetWithCapacityProviderStrategy(providerName, clusterName, tdName, svcName string, weight, base int) string {
	return testAccAWSEcsCapacityProviderConfigBase(providerName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q
  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
resource "aws_ecs_cluster" "default" {
  name = "%s"
}
resource "aws_ecs_task_definition" "mongo" {
  family = "%s"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}
resource "aws_ecs_service" "mongo" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.default.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "mongo" {
  service         = "${aws_ecs_service.mongo.id}"
  cluster         = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.test.name
    weight = %d
    base   = %d
  }
}
`, providerName, clusterName, tdName, svcName, weight, base)
}

func testAccAWSEcsTaskSetWithMultipleCapacityProviderStrategies(clusterName, tdName, svcName, sgName string) string {
	return testAccAWSEcsClusterCapacityProviders(clusterName) + fmt.Sprintf(`
resource "aws_ecs_service" "mongo" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.test.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "mongo" {
  service         = "${aws_ecs_service.mongo.id}"
  cluster         = "${aws_ecs_cluster.test.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  network_configuration {
    security_groups  = [aws_security_group.allow_all.id]
    subnets          = [aws_subnet.main.id]
    assign_public_ip = false
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}
resource "aws_ecs_task_definition" "mongo" {
  family                   = "%s"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}
resource "aws_security_group" "allow_all" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"
  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}
resource "aws_subnet" "main" {
  cidr_block        = "${cidrsubnet(aws_vpc.main.cidr_block, 8, 1)}"
  vpc_id            = "${aws_vpc.main.id}"
  tags = {
    Name = "tf-acc-ecs-service-with-multiple-capacity-providers"
  }
}
resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "tf-acc-ecs-service-with-multiple-capacity-providers"
  }
}
`, tdName, svcName, sgName)
}

func testAccAWSEcsTaskSetWithAlb(clusterName, tdName, lbName, svcName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "terraform-testacc-ecs-service-with-alb"
  }
}
resource "aws_subnet" "main" {
  count             = 2
  cidr_block        = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id            = "${aws_vpc.main.id}"
  tags = {
    Name = "tf-acc-ecs-service-with-alb"
  }
}
resource "aws_ecs_cluster" "main" {
  name = "%s"
}
resource "aws_ecs_task_definition" "with_lb_changes" {
  family = "%s"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}
resource "aws_lb_target_group" "test" {
  name     = "${aws_lb.main.name}"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.main.id}"
}
resource "aws_lb" "main" {
  name     = "%s"
  internal = true
  subnets  = ["${aws_subnet.main.*.id[0]}", "${aws_subnet.main.*.id[1]}"]
}
resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.main.id}"
  port              = "80"
  protocol          = "HTTP"
  default_action {
    target_group_arn = "${aws_lb_target_group.test.id}"
    type             = "forward"
  }
}
resource "aws_ecs_service" "with_alb" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.main.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "with_alb" {
  service         = "${aws_ecs_service.with_alb.id}"
  cluster         = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.with_lb_changes.arn}"
  load_balancers {
    target_group_arn = "${aws_lb_target_group.test.id}"
    container_name   = "ghost"
    container_port   = "2368"
  }
}
`, clusterName, tdName, lbName, svcName)
}

func testAccAWSEcsTaskSetConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
}
resource "aws_ecs_task_definition" "test" {
  family = %q
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}
resource "aws_ecs_service" "test" {
  cluster               = "${aws_ecs_cluster.test.id}"
  desired_count         = 0
  name                  = %q
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "test" {
  service         = "${aws_ecs_service.test.id}"
  cluster         = "${aws_ecs_cluster.test.id}"
  task_definition = "${aws_ecs_task_definition.test.arn}"
  tags = {
    %q = %q
  }
}
`, rName, rName, rName, tag1Key, tag1Value)
}

func testAccAWSEcsTaskSetConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %q
}
resource "aws_ecs_task_definition" "test" {
  family = %q
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}
resource "aws_ecs_service" "test" {
  cluster               = "${aws_ecs_cluster.test.id}"
  desired_count         = 0
  name                  = %q
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "test" {
  service         = "${aws_ecs_service.test.id}"
  cluster         = "${aws_ecs_cluster.test.id}"
  task_definition = "${aws_ecs_task_definition.test.arn}"
  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, rName, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAWSEcsTaskSet_withServiceRegistries(rName, clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "tf-acc-with-svc-reg"
  }
}
resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = "${cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.test.names[count.index]}"
  vpc_id            = "${aws_vpc.test.id}"
  tags = {
    Name = "tf-acc-with-svc-reg"
  }
}
resource "aws_security_group" "test" {
  name   = "tf-acc-sg-%s"
  vpc_id = "${aws_vpc.test.id}"
  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["${aws_vpc.test.cidr_block}"]
  }
}
resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "tf-acc-sd-%s.terraform.local"
  description = "test"
  vpc         = "${aws_vpc.test.id}"
}
resource "aws_service_discovery_service" "test" {
  name = "tf-acc-sd-%s"
  dns_config {
    namespace_id = "${aws_service_discovery_private_dns_namespace.test.id}"
    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}
resource "aws_ecs_cluster" "test" {
  name = "%s"
}
resource "aws_ecs_task_definition" "test" {
  family       = "%s"
  network_mode = "awsvpc"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}
resource "aws_ecs_service" "test" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.test.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "test" {
  service         = "${aws_ecs_service.test.id}"
  cluster         = "${aws_ecs_cluster.test.id}"
  task_definition = "${aws_ecs_task_definition.test.arn}"
  service_registries {
    port         = 34567
    registry_arn = "${aws_service_discovery_service.test.arn}"
  }
  network_configuration {
    security_groups = ["${aws_security_group.test.id}"]
    subnets         = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
  }
}
`, rName, rName, rName, clusterName, tdName, svcName)
}

func testAccAWSEcsTaskSetWithLaunchTypeFargate(sg1Name, sg2Name, clusterName, tdName, svcName, assignPublicIP string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "terraform-testacc-ecs-service-with-launch-type-fargate"
  }
}
resource "aws_subnet" "main" {
  count             = 2
  cidr_block        = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id            = "${aws_vpc.main.id}"
  tags = {
    Name = "tf-acc-ecs-service-with-launch-type-fargate"
  }
}
resource "aws_security_group" "allow_all_a" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}
resource "aws_security_group" "allow_all_b" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}
resource "aws_ecs_cluster" "main" {
  name = "%s"
}
resource "aws_ecs_task_definition" "mongo" {
  family                   = "%s"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}
resource "aws_ecs_service" "main" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.main.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "main" {
  service          = "${aws_ecs_service.main.id}"
  cluster          = "${aws_ecs_cluster.main.id}"
  task_definition  = "${aws_ecs_task_definition.mongo.arn}"
  launch_type      = "FARGATE"
  platform_version = "1.3.0"
  network_configuration {
    security_groups  = ["${aws_security_group.allow_all_a.id}", "${aws_security_group.allow_all_b.id}"]
    subnets          = ["${aws_subnet.main.*.id[0]}", "${aws_subnet.main.*.id[1]}"]
    assign_public_ip = %s
  }
}
`, sg1Name, sg2Name, clusterName, tdName, svcName, assignPublicIP)
}

func testAccAWSEcsTaskSetWithLaunchTypeFargateAndPlatformVersion(sg1Name, sg2Name, clusterName, tdName, svcName, platformVersion string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
resource "aws_vpc" "main" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "terraform-testacc-ecs-service-with-launch-type-fargate-and-platform-version"
  }
}
resource "aws_subnet" "main" {
  count             = 2
  cidr_block        = "${cidrsubnet(aws_vpc.main.cidr_block, 8, count.index)}"
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  vpc_id            = "${aws_vpc.main.id}"
  tags = {
    Name = "tf-acc-ecs-service-with-launch-type-fargate-and-platform-version"
  }
}
resource "aws_security_group" "allow_all_a" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}
resource "aws_security_group" "allow_all_b" {
  name        = "%s"
  description = "Allow all inbound traffic"
  vpc_id      = "${aws_vpc.main.id}"
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = ["${aws_vpc.main.cidr_block}"]
  }
}
resource "aws_ecs_cluster" "main" {
  name = "%s"
}
resource "aws_ecs_task_definition" "mongo" {
  family                   = "%s"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}
resource "aws_ecs_service" "main" {
  name                  = "%s"
  cluster               = "${aws_ecs_cluster.main.id}"
  desired_count         = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
resource "aws_ecs_task_set" "main" {
  service         = "${aws_ecs_service.main.id}"
  cluster         = "${aws_ecs_cluster.main.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  launch_type      = "FARGATE"
  platform_version = %q
  network_configuration {
    security_groups  = ["${aws_security_group.allow_all_a.id}", "${aws_security_group.allow_all_b.id}"]
    subnets          = ["${aws_subnet.main.*.id[0]}", "${aws_subnet.main.*.id[1]}"]
    assign_public_ip = false
  }
}
`, sg1Name, sg2Name, clusterName, tdName, svcName, platformVersion)
}

////////////
// Utils //
///////////

func testAccCheckAwsEcsTaskSetArn(resourceName, clusterName, svcName string, m *ecs.TaskSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := aws.StringValue(m.Id)
		taskSetArnPrefix := fmt.Sprintf("task-set/%s/%s/%s", clusterName, svcName, id)
		return testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ecs", taskSetArnPrefix)(s)
	}
}

func testAccCheckAWSEcsTaskSetExists(name string, taskSet *ecs.TaskSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		input := &ecs.DescribeTaskSetsInput{
			TaskSets: []*string{aws.String(rs.Primary.ID)},
			Cluster:  aws.String(rs.Primary.Attributes["cluster"]),
			Service:  aws.String(rs.Primary.Attributes["service"]),
		}
		var output *ecs.DescribeTaskSetsOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			output, err = conn.DescribeTaskSets(input)

			if err != nil {
				if isAWSErr(err, ecs.ErrCodeClusterNotFoundException, "") ||
					isAWSErr(err, ecs.ErrCodeServiceNotFoundException, "") ||
					isAWSErr(err, ecs.ErrCodeTaskSetNotFoundException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			if len(output.TaskSets) == 0 {
				return resource.RetryableError(fmt.Errorf("task set not found: %s", rs.Primary.ID))
			}

			return nil
		})

		if err != nil {
			return err
		}

		*taskSet = *output.TaskSets[0]

		return nil
	}
}

func testAccCheckAWSEcsTaskSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_task_set" {
			continue
		}

		out, err := conn.DescribeTaskSets(&ecs.DescribeTaskSetsInput{
			TaskSets: []*string{aws.String(rs.Primary.ID)},
			Cluster:  aws.String(rs.Primary.Attributes["cluster"]),
			Service:  aws.String(rs.Primary.Attributes["service"]),
		})

		if err == nil {
			if len(out.TaskSets) > 0 {
				var activeTaskSets []*ecs.TaskSet
				for _, ts := range out.TaskSets {
					if *ts.Status != "INACTIVE" {
						activeTaskSets = append(activeTaskSets, ts)
					}
				}
				if len(activeTaskSets) == 0 {
					return nil
				}

				return fmt.Errorf("ECS task set still exists:\n%#v", activeTaskSets)
			}
			return nil
		}

		return err
	}

	return nil
}
