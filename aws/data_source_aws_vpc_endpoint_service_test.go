package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsVpcEndpointService_gateway(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDnsService(datasourceName, "service_name", "dynamodb"),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", "data.aws_availability_zones.available", "names.#"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", ""),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDnsService(datasourceName, "service_name", "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					testAccCheckResourceAttrRegionalHostnameService(datasourceName, "private_dns_name", "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceCustomConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					testAccMatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom_filter(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceCustomConfigFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					testAccMatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom_filter_tags(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceCustomConfigFilterTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "true"),
					resource.TestCheckResourceAttr(datasourceName, "availability_zones.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "false"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tags.Name", rName),
					testAccMatchResourceAttrRegionalARN(datasourceName, "arn", "ec2", regexp.MustCompile(`vpc-endpoint-service/vpce-svc-.+`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_ServiceType_Gateway(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceConfig_ServiceType("s3", "Gateway"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDnsService(datasourceName, "service_name", "s3"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_ServiceType_Interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceConfig_ServiceType("ec2", "Interface"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrRegionalReverseDnsService(datasourceName, "service_name", "ec2"),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcEndpointServiceGatewayConfig = `
data "aws_availability_zones" "available" {}

data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}
`

const testAccDataSourceAwsVpcEndpointServiceInterfaceConfig = `
data "aws_vpc_endpoint_service" "test" {
  service = "ec2"
}
`

func testAccDataSourceAwsVpcEndpointServiceConfig_ServiceType(service string, serviceType string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service      = %[1]q
  service_type = %[2]q
}
`, service, serviceType)
}

func testAccDataSourceAwsVpcEndpointServiceCustomConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    aws_lb.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccDataSourceAwsVpcEndpointServiceCustomConfig(rName string) string {
	return testAccDataSourceAwsVpcEndpointServiceCustomConfigBase(rName) + `
data "aws_vpc_endpoint_service" "test" {
  service_name = aws_vpc_endpoint_service.test.service_name
}
`
}

func testAccDataSourceAwsVpcEndpointServiceCustomConfigFilter(rName string) string {
	return testAccDataSourceAwsVpcEndpointServiceCustomConfigBase(rName) + `
data "aws_vpc_endpoint_service" "test" {
  filter {
    name   = "service-name"
    values = [aws_vpc_endpoint_service.test.service_name]
  }
}
`
}

func testAccDataSourceAwsVpcEndpointServiceCustomConfigFilterTags(rName string) string {
	return testAccDataSourceAwsVpcEndpointServiceCustomConfigBase(rName) + `
data "aws_vpc_endpoint_service" "test" {
  tags = {
    Name = aws_vpc_endpoint_service.test.tags["Name"]
  }
}
`
}
