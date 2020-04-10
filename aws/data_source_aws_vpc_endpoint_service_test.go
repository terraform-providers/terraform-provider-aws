package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsVpcEndpointService_gateway(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	region := testAccGetRegion()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "service_name", fmt.Sprintf("com.amazonaws.%s.s3", region)),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttrPair(datasourceName, "availability_zones.#", "data.aws_availability_zones.available", "names.#"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", ""),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Gateway"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_interface(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	region := testAccGetRegion()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointServiceInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "service_name", fmt.Sprintf("com.amazonaws.%s.ec2", region)),
					resource.TestCheckResourceAttr(datasourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(datasourceName, "base_endpoint_dns_names.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "manages_vpc_endpoints", "false"),
					resource.TestCheckResourceAttr(datasourceName, "owner", "amazon"),
					resource.TestCheckResourceAttr(datasourceName, "private_dns_name", fmt.Sprintf("ec2.%s.%s", region, testAccGetPartitionDNSSuffix())),
					resource.TestCheckResourceAttr(datasourceName, "service_type", "Interface"),
					resource.TestCheckResourceAttr(datasourceName, "vpc_endpoint_policy_supported", "true"),
					resource.TestCheckResourceAttr(datasourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

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
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom_filter(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

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
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpointService_custom_filter_tags(t *testing.T) {
	datasourceName := "data.aws_vpc_endpoint_service.test"
	rName := fmt.Sprintf("tf-testacc-vpcesvc-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))

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
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcEndpointServiceGatewayConfig = `
data "aws_availability_zones" "available" {}

data "aws_vpc_endpoint_service" "test" {
  service = "s3"
}
`

const testAccDataSourceAwsVpcEndpointServiceInterfaceConfig = `
data "aws_vpc_endpoint_service" "test" {
  service = "ec2"
}
`

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
    "${aws_subnet.test1.id}",
    "${aws_subnet.test2.id}",
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
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    "${aws_lb.test.id}",
  ]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccDataSourceAwsVpcEndpointServiceCustomConfig(rName string) string {
	return testAccDataSourceAwsVpcEndpointServiceCustomConfigBase(rName) + fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service_name = "${aws_vpc_endpoint_service.test.service_name}"
}
`)
}

func testAccDataSourceAwsVpcEndpointServiceCustomConfigFilter(rName string) string {
	return testAccDataSourceAwsVpcEndpointServiceCustomConfigBase(rName) + fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  filter {
    name   = "service-name"
    values = ["${aws_vpc_endpoint_service.test.service_name}"]
  }
}
`)
}

func testAccDataSourceAwsVpcEndpointServiceCustomConfigFilterTags(rName string) string {
	return testAccDataSourceAwsVpcEndpointServiceCustomConfigBase(rName) + fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  tags = {
    Name  = "${aws_vpc_endpoint_service.test.tags["Name"]}"
  }
}
`)
}
