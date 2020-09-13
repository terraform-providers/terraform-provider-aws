package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSALBTargetGroup_basic(t *testing.T) {
	lbName := fmt.Sprintf("testlb-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceNameArn := "data.aws_lb_target_group.alb_tg_test_with_arn"
	resourceName := "data.aws_lb_target_group.alb_tg_test_with_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBTargetGroupConfigBasic(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameArn, "name", targetGroupName),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceNameArn, "port", "8080"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "vpc_id"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttr(resourceNameArn, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceNameArn, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.TestName", "TestAccDataSourceAWSALBTargetGroup_basic"),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.matcher", "200-299"),

					resource.TestCheckResourceAttr(resourceName, "name", targetGroupName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet(resourceName, "load_balancing_algorithm_type"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", "TestAccDataSourceAWSALBTargetGroup_basic"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLBTargetGroup_BackwardsCompatibility(t *testing.T) {
	lbName := fmt.Sprintf("testlb-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceNameArn := "data.aws_alb_target_group.alb_tg_test_with_arn"
	resourceName := "data.aws_alb_target_group.alb_tg_test_with_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBTargetGroupConfigBackwardsCompatibility(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceNameArn, "name", targetGroupName),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceNameArn, "port", "8080"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet(resourceNameArn, "vpc_id"),
					resource.TestCheckResourceAttr(resourceNameArn, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceNameArn, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "tags.TestName", "TestAccDataSourceAWSALBTargetGroup_basic"),
					resource.TestCheckResourceAttr(resourceNameArn, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceNameArn, "health_check.0.matcher", "200-299"),

					resource.TestCheckResourceAttr(resourceName, "name", targetGroupName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "arn_suffix"),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
					resource.TestCheckResourceAttr(resourceName, "deregistration_delay", "300"),
					resource.TestCheckResourceAttr(resourceName, "slow_start", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", "TestAccDataSourceAWSALBTargetGroup_basic"),
					resource.TestCheckResourceAttr(resourceName, "stickiness.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.path", "/health"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.port", "8081"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.timeout", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.healthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.unhealthy_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check.0.matcher", "200-299"),
				),
			},
		},
	})
}

func testAccDataSourceAWSLBTargetGroupConfigBasic(lbName string, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_lb_listener" "front_end" {
  load_balancer_arn = aws_lb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccDataSourceAWSALBTargetGroup_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    TestName = "TestAccDataSourceAWSALBTargetGroup_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-data-source-target-group-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-data-source-target-group-basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    TestName = "TestAccDataSourceAWSALBTargetGroup_basic"
  }
}

data "aws_lb_target_group" "alb_tg_test_with_arn" {
  arn = aws_lb_target_group.test.arn
}

data "aws_lb_target_group" "alb_tg_test_with_name" {
  name = aws_lb_target_group.test.name
}
`, lbName, targetGroupName)
}

func testAccDataSourceAWSLBTargetGroupConfigBackwardsCompatibility(lbName string, targetGroupName string) string {
	return fmt.Sprintf(`
resource "aws_alb_listener" "front_end" {
  load_balancer_arn = aws_alb.alb_test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_alb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_alb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccDataSourceAWSALBTargetGroup_basic"
  }
}

resource "aws_alb_target_group" "test" {
  name     = "%s"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.alb_test.id

  health_check {
    path                = "/health"
    interval            = 60
    port                = 8081
    protocol            = "HTTP"
    timeout             = 3
    healthy_threshold   = 3
    unhealthy_threshold = 3
    matcher             = "200-299"
  }

  tags = {
    TestName = "TestAccDataSourceAWSALBTargetGroup_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-data-source-target-group-bc"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-data-source-target-group-bc"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    TestName = "TestAccDataSourceAWSALBTargetGroup_basic"
  }
}

data "aws_alb_target_group" "alb_tg_test_with_arn" {
  arn = aws_alb_target_group.test.arn
}

data "aws_alb_target_group" "alb_tg_test_with_name" {
  name = aws_alb_target_group.test.name
}
`, lbName, targetGroupName)
}
