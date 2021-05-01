package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSLBListener_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lb_listener.test"
	dataSourceName2 := "data.aws_lb_listener.from_lb_and_port"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBListenerConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName, "port", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "load_balancer_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttr(dataSourceName2, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName2, "port", "80"),
					resource.TestCheckResourceAttr(dataSourceName2, "default_action.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName2, "default_action.0.type", "forward"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLBListener_BackwardsCompatibility(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_alb_listener.test"
	dataSourceName2 := "data.aws_alb_listener.from_lb_and_port"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBListenerConfigBackwardsCompatibility(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName, "port", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "load_balancer_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttr(dataSourceName2, "protocol", "HTTP"),
					resource.TestCheckResourceAttr(dataSourceName2, "port", "80"),
					resource.TestCheckResourceAttr(dataSourceName2, "default_action.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName2, "default_action.0.type", "forward"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLBListener_https(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")
	dataSourceName := "data.aws_lb_listener.test"
	dataSourceName2 := "data.aws_lb_listener.from_lb_and_port"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBListenerConfigHTTPS(rName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "load_balancer_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "certificate_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(dataSourceName, "port", "443"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(dataSourceName, "ssl_policy", "ELBSecurityPolicy-2016-08"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "load_balancer_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "certificate_arn"),
					resource.TestCheckResourceAttr(dataSourceName2, "protocol", "HTTPS"),
					resource.TestCheckResourceAttr(dataSourceName2, "port", "443"),
					resource.TestCheckResourceAttr(dataSourceName2, "default_action.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName2, "default_action.0.type", "forward"),
					resource.TestCheckResourceAttr(dataSourceName2, "ssl_policy", "ELBSecurityPolicy-2016-08"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLBListener_DefaultAction_Forward(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lb_listener.test"
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elbv2.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBListenerConfigDefaultActionForward(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "default_action.#", resourceName, "default_action.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_action.0.forward.#", resourceName, "default_action.0.forward.#"),
				),
			},
		},
	})
}

func testAccDataSourceAWSLBListenerConfigBasic(rName string) string {
	return composeConfig(testAccAWSLBListenerConfigBase(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

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
}

data "aws_lb_listener" "test" {
  arn = aws_lb_listener.test.arn
}

data "aws_lb_listener" "from_lb_and_port" {
  load_balancer_arn = aws_lb.test.arn
  port              = aws_lb_listener.test.port
}
`, rName))
}

func testAccDataSourceAWSLBListenerConfigBackwardsCompatibility(rName string) string {
	return composeConfig(testAccAWSLBListenerConfigBase(rName), fmt.Sprintf(`
resource "aws_alb_listener" "test" {
  load_balancer_arn = aws_alb.test.id
  protocol          = "HTTP"
  port              = "80"

  default_action {
    target_group_arn = aws_alb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_alb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccAWSALB_basic"
  }
}

resource "aws_alb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

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
}

data "aws_alb_listener" "test" {
  arn = aws_alb_listener.test.arn
}

data "aws_alb_listener" "from_lb_and_port" {
  load_balancer_arn = aws_alb.test.arn
  port              = aws_alb_listener.test.port
}
`, rName))
}

func testAccDataSourceAWSLBListenerConfigHTTPS(rName, certificate, key string) string {
	return composeConfig(testAccAWSLBListenerConfigBase(rName), fmt.Sprintf(`
resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = aws_iam_server_certificate.test.arn

  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_lb" "test" {
  name            = %[1]q
  internal        = false
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccAWSALB_basic"
  }

  depends_on = [aws_internet_gateway.gw]
}

resource "aws_lb_target_group" "test" {
  name     = %[1]q
  port     = 8080
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id

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
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name     = %[1]q
    TestName = "TestAccAWSALB_basic"
  }
}

resource "aws_iam_server_certificate" "test" {
  name             = %[1]q
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_lb_listener" "test" {
  arn = aws_lb_listener.test.arn
}

data "aws_lb_listener" "from_lb_and_port" {
  load_balancer_arn = aws_lb.test.arn
  port              = aws_lb_listener.test.port
}
`, rName, certificate, key))
}

func testAccDataSourceAWSLBListenerConfigDefaultActionForward(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  internal = true
  name     = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test[0].id
  }

  subnet_mapping {
    subnet_id = aws_subnet.test[1].id
  }
}

resource "aws_lb_target_group" "test" {
  count = 2

  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "forward"

    forward {
      target_group {
        arn    = aws_lb_target_group.test[0].arn
        weight = 1
      }

      target_group {
        arn    = aws_lb_target_group.test[1].arn
        weight = 2
      }
    }
  }
}

data "aws_lb_listener" "test" {
  arn = aws_lb_listener.test.arn
}
`, rName))
}
