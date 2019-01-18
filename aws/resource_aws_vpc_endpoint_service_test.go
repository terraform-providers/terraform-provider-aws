package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSVpcEndpointService_importBasic(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	resourceName := "aws_vpc_endpoint_service.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceBasicConfig(lbName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVpcEndpointService_basic(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	lb1Name := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	lb2Name := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpc_endpoint_service.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceBasicConfig(lb1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists("aws_vpc_endpoint_service.foo", &svcCfg),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_service.foo", "acceptance_required", "false"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_service.foo", "network_load_balancer_arns.#", "1"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_service.foo", "allowed_principals.#", "1"),
				),
			},
			{
				Config: testAccVpcEndpointServiceModifiedConfig(lb1Name, lb2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists("aws_vpc_endpoint_service.foo", &svcCfg),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_service.foo", "acceptance_required", "true"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_service.foo", "network_load_balancer_arns.#", "2"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint_service.foo", "allowed_principals.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointService_removed(t *testing.T) {
	var svcCfg ec2.ServiceConfiguration
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	testDestroy := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DeleteVpcEndpointServiceConfigurations(&ec2.DeleteVpcEndpointServiceConfigurationsInput{
			ServiceIds: []*string{svcCfg.ServiceId},
		})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceBasicConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointServiceExists("aws_vpc_endpoint_service.foo", &svcCfg),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVpcEndpointServiceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_service" {
			continue
		}

		resp, err := conn.DescribeVpcEndpointServiceConfigurations(&ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidVpcEndpointServiceId.NotFound" {
				continue
			}
			return err
		}
		if len(resp.ServiceConfigurations) > 0 {
			return fmt.Errorf("VPC Endpoint Services still exist.")
		}

		return err
	}

	return nil
}

func testAccCheckVpcEndpointServiceExists(n string, svcCfg *ec2.ServiceConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint Service ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeVpcEndpointServiceConfigurations(&ec2.DescribeVpcEndpointServiceConfigurationsInput{
			ServiceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.ServiceConfigurations) == 0 {
			return fmt.Errorf("VPC Endpoint Service not found")
		}

		*svcCfg = *resp.ServiceConfigurations[0]

		return nil
	}
}

func testAccVpcEndpointServiceBasicConfig(lb1Name string) string {
	return fmt.Sprintf(
		`
resource "aws_vpc" "nlb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-service"
  }
}

resource "aws_lb" "nlb_test_1" {
  name = "%s"

  subnets = [
    "${aws_subnet.nlb_test_1.id}",
    "${aws_subnet.nlb_test_2.id}",
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = "testAccVpcEndpointServiceBasicConfig_nlb1"
  }
}

resource "aws_subnet" "nlb_test_1" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-vpc-endpoint-service-1"
  }
}

resource "aws_subnet" "nlb_test_2" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-vpc-endpoint-service-2"
  }
}

data "aws_caller_identity" "current" {}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required = false

  network_load_balancer_arns = [
    "${aws_lb.nlb_test_1.id}",
  ]

  allowed_principals = [
    "${data.aws_caller_identity.current.arn}"
  ]
}
	`, lb1Name)
}

func testAccVpcEndpointServiceModifiedConfig(lb1Name, lb2Name string) string {
	return fmt.Sprintf(
		`
resource "aws_vpc" "nlb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-service"
  }
}

resource "aws_lb" "nlb_test_1" {
  name = "%s"

  subnets = [
    "${aws_subnet.nlb_test_1.id}",
    "${aws_subnet.nlb_test_2.id}",
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = "testAccVpcEndpointServiceBasicConfig_nlb1"
  }
}

resource "aws_lb" "nlb_test_2" {
	name = "%s"

	subnets = [
	  "${aws_subnet.nlb_test_1.id}",
	  "${aws_subnet.nlb_test_2.id}",
	]

	load_balancer_type         = "network"
	internal                   = true
	idle_timeout               = 60
	enable_deletion_protection = false

	tags = {
	  Name = "testAccVpcEndpointServiceBasicConfig_nlb2"
	}
  }

resource "aws_subnet" "nlb_test_1" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-vpc-endpoint-service-1"
  }
}

resource "aws_subnet" "nlb_test_2" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-vpc-endpoint-service-2"
  }
}

data "aws_caller_identity" "current" {}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required = true

  network_load_balancer_arns = [
	"${aws_lb.nlb_test_1.id}",
	"${aws_lb.nlb_test_2.id}",
  ]

  allowed_principals = []
}
	`, lb1Name, lb2Name)
}
