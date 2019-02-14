package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLBListener_basic(t *testing.T) {
	var conf elbv2.Listener
	lbName := fmt.Sprintf("testlistener-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener.front_end",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_basic(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists("aws_lb_listener.front_end", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "load_balancer_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "port", "80"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.fixed_response.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_BackwardsCompatibility(t *testing.T) {
	var conf elbv2.Listener
	lbName := fmt.Sprintf("testlistener-basic-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_alb_listener.front_end",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfigBackwardsCompatibility(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists("aws_alb_listener.front_end", &conf),
					resource.TestCheckResourceAttrSet("aws_alb_listener.front_end", "load_balancer_arn"),
					resource.TestCheckResourceAttrSet("aws_alb_listener.front_end", "arn"),
					resource.TestCheckResourceAttr("aws_alb_listener.front_end", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_alb_listener.front_end", "port", "80"),
					resource.TestCheckResourceAttr("aws_alb_listener.front_end", "default_action.#", "1"),
					resource.TestCheckResourceAttr("aws_alb_listener.front_end", "default_action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_alb_listener.front_end", "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_alb_listener.front_end", "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttr("aws_alb_listener.front_end", "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("aws_alb_listener.front_end", "default_action.0.fixed_response.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_https(t *testing.T) {
	var conf elbv2.Listener
	lbName := fmt.Sprintf("testlistener-https-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	targetGroupName := fmt.Sprintf("testtargetgroup-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener.front_end",
		Providers:     testAccProvidersWithTLS,
		CheckDestroy:  testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_https(lbName, targetGroupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists("aws_lb_listener.front_end", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "load_balancer_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "default_action.0.target_group_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.fixed_response.#", "0"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "certificate_arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "ssl_policy", "ELBSecurityPolicy-2016-08"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_Protocol_Tls(t *testing.T) {
	var listener1 elbv2.Listener
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_Protocol_Tls(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists(resourceName, &listener1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "TLS"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_redirect(t *testing.T) {
	var conf elbv2.Listener
	lbName := fmt.Sprintf("testlistener-redirect-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener.front_end",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_redirect(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists("aws_lb_listener.front_end", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "load_balancer_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "port", "80"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.type", "redirect"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.0.host", "#{host}"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.0.path", "/#{path}"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.0.port", "443"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.0.protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.0.query", "#{query}"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.0.status_code", "HTTP_301"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.fixed_response.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_fixedResponse(t *testing.T) {
	var conf elbv2.Listener
	lbName := fmt.Sprintf("testlistener-fixedresponse-%s", acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener.front_end",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_fixedResponse(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists("aws_lb_listener.front_end", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "load_balancer_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.front_end", "arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "protocol", "HTTP"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "port", "80"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.type", "fixed-response"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.target_group_arn", ""),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.redirect.#", "0"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.fixed_response.#", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.fixed_response.0.content_type", "text/plain"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.fixed_response.0.message_body", "Fixed response content"),
					resource.TestCheckResourceAttr("aws_lb_listener.front_end", "default_action.0.fixed_response.0.status_code", "200"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_cognito(t *testing.T) {
	var conf elbv2.Listener
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener.test",
		Providers:     testAccProvidersWithTLS,
		CheckDestroy:  testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_cognito(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists("aws_lb_listener.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "load_balancer_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.#", "2"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.type", "authenticate-cognito"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "default_action.0.authenticate_cognito.0.user_pool_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "default_action.0.authenticate_cognito.0.user_pool_client_id"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "default_action.0.authenticate_cognito.0.user_pool_domain"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_cognito.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_cognito.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.1.type", "forward"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.1.order", "2"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "default_action.1.target_group_arn"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_oidc(t *testing.T) {
	var conf elbv2.Listener
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_lb_listener.test",
		Providers:     testAccProvidersWithTLS,
		CheckDestroy:  testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_oidc(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists("aws_lb_listener.test", &conf),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "load_balancer_arn"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "arn"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "protocol", "HTTPS"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "port", "443"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.#", "2"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.order", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.type", "authenticate-oidc"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.authorization_endpoint", "https://example.com/authorization_endpoint"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.client_id", "s6BhdRkqt3"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.client_secret", "7Fjfp0ZBr1KtDRbnfVdmIw"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.issuer", "https://example.com"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.token_endpoint", "https://example.com/token_endpoint"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.user_info_endpoint", "https://example.com/user_info_endpoint"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.authentication_request_extra_params.%", "1"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.0.authenticate_oidc.0.authentication_request_extra_params.param", "test"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.1.order", "2"),
					resource.TestCheckResourceAttr("aws_lb_listener.test", "default_action.1.type", "forward"),
					resource.TestCheckResourceAttrSet("aws_lb_listener.test", "default_action.1.target_group_arn"),
				),
			},
		},
	})
}

func TestAccAWSLBListener_DefaultAction_Order(t *testing.T) {
	var listener elbv2.Listener
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_DefaultAction_Order(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists(resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", "2"),
				),
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6171
func TestAccAWSLBListener_DefaultAction_Order_Recreates(t *testing.T) {
	var listener elbv2.Listener
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lb_listener.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSLBListenerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLBListenerConfig_DefaultAction_Order(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSLBListenerExists(resourceName, &listener),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.1.order", "2"),
					testAccCheckAWSLBListenerDefaultActionOrderDisappears(&listener, 1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSLBListenerDefaultActionOrderDisappears(listener *elbv2.Listener, actionOrderToDelete int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var newDefaultActions []*elbv2.Action

		for i, action := range listener.DefaultActions {
			if int(aws.Int64Value(action.Order)) == actionOrderToDelete {
				newDefaultActions = append(listener.DefaultActions[:i], listener.DefaultActions[i+1:]...)
				break
			}
		}

		if len(newDefaultActions) == 0 {
			return fmt.Errorf("Unable to find default action order %d from default actions: %#v", actionOrderToDelete, listener.DefaultActions)
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		input := &elbv2.ModifyListenerInput{
			DefaultActions: newDefaultActions,
			ListenerArn:    listener.ListenerArn,
		}

		_, err := conn.ModifyListener(input)

		return err
	}
}

func testAccCheckAWSLBListenerExists(n string, res *elbv2.Listener) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Listener ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		describe, err := conn.DescribeListeners(&elbv2.DescribeListenersInput{
			ListenerArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(describe.Listeners) != 1 ||
			*describe.Listeners[0].ListenerArn != rs.Primary.ID {
			return errors.New("Listener not found")
		}

		*res = *describe.Listeners[0]
		return nil
	}
}

func testAccCheckAWSLBListenerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).elbv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lb_listener" && rs.Type != "aws_alb_listener" {
			continue
		}

		describe, err := conn.DescribeListeners(&elbv2.DescribeListenersInput{
			ListenerArns: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.Listeners) != 0 &&
				*describe.Listeners[0].ListenerArn == rs.Primary.ID {
				return fmt.Errorf("Listener %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if isAWSErr(err, elbv2.ErrCodeListenerNotFoundException, "") {
			return nil
		} else {
			return fmt.Errorf("Unexpected error checking LB Listener destroyed: %s", err)
		}
	}

	return nil
}

func testAccAWSLBListenerConfig_basic(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-basic-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
    Name = "TestAccAWSALB_basic"
  }
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerConfigBackwardsCompatibility(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_alb_listener" "front_end" {
   load_balancer_arn = "${aws_alb.alb_test.id}"
   protocol = "HTTP"
   port = "80"

   default_action {
     target_group_arn = "${aws_alb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_alb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_alb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-bc"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-bc-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
    Name = "TestAccAWSALB_basic"
  }
}`, lbName, targetGroupName)
}

func testAccAWSLBListenerConfig_https(lbName, targetGroupName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener" "front_end" {
   load_balancer_arn = "${aws_lb.alb_test.id}"
   protocol = "HTTPS"
   port = "443"
   ssl_policy = "ELBSecurityPolicy-2016-08"
   certificate_arn = "${aws_iam_server_certificate.test_cert.arn}"

   default_action {
     target_group_arn = "${aws_lb_target_group.test.id}"
     type = "forward"
   }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = false
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_basic"
  }

  depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.alb_test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-https"
  }
}

resource "aws_internet_gateway" "gw" {
    vpc_id = "${aws_vpc.alb_test.id}"

  tags = {
        Name = "TestAccAWSALB_basic"
    }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-https-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
    Name = "TestAccAWSALB_basic"
  }
}

resource "aws_iam_server_certificate" "test_cert" {
  name = "terraform-test-cert-%d"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
  private_key      = "${tls_private_key.example.private_key_pem}"
}

resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}
`, lbName, targetGroupName, acctest.RandInt())
}

func testAccAWSLBListenerConfig_Protocol_Tls(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.private_key_pem}"
  validity_period_hours = 12

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }
}

resource "aws_acm_certificate" "test" {
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-lb-listener-protocol-tls"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-lb-listener-protocol-tls"
  }
}

resource "aws_lb" "test" {
  internal           = true
  load_balancer_type = "network"
  name               = %q
  subnets            = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]

  tags = {
    Name = "tf-acc-test-lb-listener-protocol-tls"
  }
}

resource "aws_lb_target_group" "test" {
  name     = %q
  port     = 443
  protocol = "TCP"
  vpc_id   = "${aws_vpc.test.id}"

  health_check {
    interval            = 10
    port                = "traffic-port"
    protocol            = "TCP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
  }

  tags = {
    Name = "tf-acc-test-lb-listener-protocol-tls"
  }
}

resource "aws_lb_listener" "test" {
  certificate_arn   = "${aws_acm_certificate.test.arn}"
  load_balancer_arn = "${aws_lb.test.arn}"
  port              = "443"
  protocol          = "TLS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"

  default_action {
    target_group_arn = "${aws_lb_target_group.test.arn}"
    type             = "forward"
  }
}
`, rName, rName)
}

func testAccAWSLBListenerConfig_redirect(lbName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.alb_test.id}"
  protocol = "HTTP"
  port = "80"

  default_action {
    type = "redirect"
    redirect {
      port = "443"
      protocol = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_redirect"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-redirect"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-redirect-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
    Name = "TestAccAWSALB_redirect"
  }
}`, lbName)
}

func testAccAWSLBListenerConfig_fixedResponse(lbName string) string {
	return fmt.Sprintf(`resource "aws_lb_listener" "front_end" {
  load_balancer_arn = "${aws_lb.alb_test.id}"
  protocol = "HTTP"
  port = "80"

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Fixed response content"
      status_code = "200"
    }
  }
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.*.id[0]}", "${aws_subnet.alb_test.*.id[1]}"]

  idle_timeout = 30
  enable_deletion_protection = false

  tags = {
    Name = "TestAccAWSALB_fixedresponse"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-listener-fixedresponse"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-listener-fixedresponse-${count.index}"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

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
    Name = "TestAccAWSALB_fixedresponse"
  }
}`, lbName)
}

func testAccAWSLBListenerConfig_cognito(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = "%s"
  internal        = false
  security_groups = ["${aws_security_group.test.id}"]
  subnets         = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
  enable_deletion_protection = false
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"
}

resource "aws_security_group" "test" {
  name        = "%s"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.test.id}"

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
}

resource "aws_cognito_user_pool" "test" {
  name = "%s"
}

resource "aws_cognito_user_pool_client" "test" {
  name = "%s"
  user_pool_id = "${aws_cognito_user_pool.test.id}"
  generate_secret = true
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows = ["code", "implicit"]
  allowed_oauth_scopes = ["phone", "email", "openid", "profile", "aws.cognito.signin.user.admin"]
  callback_urls = ["https://www.example.com/callback", "https://www.example.com/redirect"]
  default_redirect_uri = "https://www.example.com/redirect"
  logout_urls = ["https://www.example.com/login"]
}

resource "aws_cognito_user_pool_domain" "test" {
  domain = "%s"
  user_pool_id = "${aws_cognito_user_pool.test.id}"
}

resource "aws_iam_server_certificate" "test" {
  name = "terraform-test-cert-%s"
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.test.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = "${aws_lb.test.id}"
  protocol = "HTTPS"
  port = "443"
  ssl_policy = "ELBSecurityPolicy-2016-08"
  certificate_arn = "${aws_iam_server_certificate.test.arn}"

  default_action {
    type = "authenticate-cognito"
    authenticate_cognito {
      user_pool_arn = "${aws_cognito_user_pool.test.arn}"
      user_pool_client_id = "${aws_cognito_user_pool_client.test.id}"
      user_pool_domain = "${aws_cognito_user_pool_domain.test.domain}"

      authentication_request_extra_params = {
        param  = "test"
      }
    }
  }

  default_action {
    target_group_arn = "${aws_lb_target_group.test.id}"
    type = "forward"
  }
}
`, rName, rName, rName, rName, rName, rName, rName)
}

func testAccAWSLBListenerConfig_oidc(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test" {
  name            = "%s"
  internal        = false
  security_groups = ["${aws_security_group.test.id}"]
  subnets         = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
  enable_deletion_protection = false
}

resource "aws_lb_target_group" "test" {
  name = "%s"
  port = 8080
  protocol = "HTTP"
  vpc_id = "${aws_vpc.test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_subnet" "test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"
}

resource "aws_security_group" "test" {
  name        = "%s"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.test.id}"

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
}

resource "aws_iam_server_certificate" "test" {
  name = "terraform-test-cert-%s"
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.test.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = "${aws_lb.test.id}"
  protocol = "HTTPS"
  port = "443"
  ssl_policy = "ELBSecurityPolicy-2016-08"
  certificate_arn = "${aws_iam_server_certificate.test.arn}"

  default_action {
    type = "authenticate-oidc"
    authenticate_oidc {
      authorization_endpoint =  "https://example.com/authorization_endpoint"
      client_id = "s6BhdRkqt3"
      client_secret = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer = "https://example.com"
      token_endpoint = "https://example.com/token_endpoint"
      user_info_endpoint = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param  = "test"
      }
    }
  }

  default_action {
    target_group_arn = "${aws_lb_target_group.test.id}"
    type = "forward"
  }
}
`, rName, rName, rName, rName)
}

func testAccAWSLBListenerConfig_DefaultAction_Order(rName string) string {
	return fmt.Sprintf(`
variable "rName" {
  default = %q
}

data "aws_availability_zones" "available" {}

resource "aws_lb_listener" "test" {
  load_balancer_arn = "${aws_lb.test.id}"
  protocol          = "HTTPS"
  port              = "443"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "${aws_iam_server_certificate.test.arn}"

  default_action {
    order = 1
    type = "authenticate-oidc"

    authenticate_oidc {
      authorization_endpoint = "https://example.com/authorization_endpoint"
      client_id              = "s6BhdRkqt3"
      client_secret          = "7Fjfp0ZBr1KtDRbnfVdmIw"
      issuer                 = "https://example.com"
      token_endpoint         = "https://example.com/token_endpoint"
      user_info_endpoint     = "https://example.com/user_info_endpoint"

      authentication_request_extra_params = {
        param  = "test"
      }
    }
  }

  default_action {
    order            = 2
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.test.arn}"
  }
}

resource "aws_iam_server_certificate" "test" {
  certificate_body = "${tls_self_signed_cert.test.cert_pem}"
  name             = "${var.rName}"
  private_key      = "${tls_private_key.test.private_key_pem}"
}

resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "test" {
  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.private_key_pem}"
  validity_period_hours = 12

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_lb" "test" {
  internal                   = true
  name                       = "${var.rName}"
  security_groups            = ["${aws_security_group.test.id}"]
  subnets                    = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
}

resource "aws_lb_target_group" "test" {
  name     = "${var.rName}"
  port     = 8080
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.test.id}"

  health_check {
    path = "/health"
    interval = 60
    port = 8081
    protocol = "HTTP"
    timeout = 3
    healthy_threshold = 3
    unhealthy_threshold = 3
    matcher = "200-299"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "${var.rName}"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone       = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block              = "10.0.${count.index}.0/24"
  map_public_ip_on_launch = true
  vpc_id                  = "${aws_vpc.test.id}"

  tags = {
    Name = "${var.rName}"
  }
}

resource "aws_security_group" "test" {
  name        = "${var.rName}"
  vpc_id      = "${aws_vpc.test.id}"

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
    Name = "${var.rName}"
  }
}`, rName)
}
