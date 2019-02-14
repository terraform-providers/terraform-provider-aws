package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsVpcEndpointService_gateway(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "service_name", "com.amazonaws.us-west-2.s3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "service_type", "Gateway"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "owner", "amazon"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "acceptance_required", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.#", "3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "availability_zones.2050015877", "us-west-2c"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "private_dns_name", ""),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.s3", "base_endpoint_dns_names.3003388505", "s3.us-west-2.amazonaws.com"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_interface(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "service_name", "com.amazonaws.us-west-2.ec2"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "service_type", "Interface"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "owner", "amazon"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "acceptance_required", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "availability_zones.#", "3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "availability_zones.221770259", "us-west-2b"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "availability_zones.2050015877", "us-west-2c"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "private_dns_name", "ec2.us-west-2.amazonaws.com"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.ec2", "base_endpoint_dns_names.1880016359", "ec2.us-west-2.vpce.amazonaws.com"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceCustomConfig(lbName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.foo", "service_type", "Interface"),
					resource.TestMatchResourceAttr( // AWS account ID
						"data.aws_vpc_endpoint_service.foo", "owner", regexp.MustCompile("^[0-9]{12}$")),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.foo", "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.foo", "acceptance_required", "true"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.foo", "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.foo", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint_service.foo", "availability_zones.221770259", "us-west-2b"),
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcEndpointServiceGatewayConfig = `
provider "aws" {
  region = "us-west-2"
}

data "aws_vpc_endpoint_service" "s3" {
  service = "s3"
}
`

const testAccDataSourceAwsVpcEndpointServiceInterfaceConfig = `
provider "aws" {
  region = "us-west-2"
}

data "aws_vpc_endpoint_service" "ec2" {
  service = "ec2"
}
`

func testAccDataSourceAwsVpcEndpointServiceCustomConfig(lbName string) string {
	return fmt.Sprintf(
		`
resource "aws_vpc" "nlb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-service-custom"
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
    Name = "tf-acc-vpc-endpoint-service-custom"
  }
}

resource "aws_subnet" "nlb_test_2" {
  vpc_id            = "${aws_vpc.nlb_test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-vpc-endpoint-service-custom"
  }
}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required = true

  network_load_balancer_arns = [
    "${aws_lb.nlb_test_1.id}",
  ]
}

data "aws_vpc_endpoint_service" "foo" {
  service_name = "${aws_vpc_endpoint_service.foo.service_name}"
}
  `, lbName)
}
