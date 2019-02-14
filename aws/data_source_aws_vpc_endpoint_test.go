package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsVpcEndpoint_gatewayBasic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_gatewayBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcEndpointCheckExists("data.aws_vpc_endpoint.s3", "aws_vpc_endpoint.s3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint.s3", "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttrSet("data.aws_vpc_endpoint.s3", "prefix_list_id"),
					resource.TestCheckResourceAttrSet("data.aws_vpc_endpoint.s3", "cidr_blocks.#"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.s3", "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.s3", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.s3", "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.s3", "security_group_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_byId(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_byId,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcEndpointCheckExists("data.aws_vpc_endpoint.by_id", "aws_vpc_endpoint.s3"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_gatewayWithRouteTable(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_gatewayWithRouteTable,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcEndpointCheckExists("data.aws_vpc_endpoint.s3", "aws_vpc_endpoint.s3"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint.s3", "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint.s3", "route_table_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcEndpoint_interface(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcEndpointConfig_interface,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsVpcEndpointCheckExists("data.aws_vpc_endpoint.ec2", "aws_vpc_endpoint.ec2"),
					resource.TestCheckResourceAttr(
						"data.aws_vpc_endpoint.ec2", "vpc_endpoint_type", "Interface"),
					resource.TestCheckNoResourceAttr("data.aws_vpc_endpoint.ec2", "prefix_list_id"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.ec2", "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.ec2", "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.ec2", "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.ec2", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("data.aws_vpc_endpoint.ec2", "private_dns_enabled", "false"),
				),
			},
		},
	})
}

func testAccDataSourceAwsVpcEndpointCheckExists(dsName, rsName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", dsName)
		}

		vpceRs, ok := s.RootModule().Resources[rsName]
		if !ok {
			return fmt.Errorf("can't find %s in state", rsName)
		}

		attr := rs.Primary.Attributes

		if attr["id"] != vpceRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"id is %s; want %s",
				attr["id"],
				vpceRs.Primary.Attributes["id"],
			)
		}

		return nil
	}
}

const testAccDataSourceAwsVpcEndpointConfig_gatewayBasic = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-gw-basic"
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "${aws_vpc_endpoint.s3.service_name}"
  state = "available"
}
`

const testAccDataSourceAwsVpcEndpointConfig_byId = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-by-id"
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
}

data "aws_vpc_endpoint" "by_id" {
  id = "${aws_vpc_endpoint.s3.id}"
}
`

const testAccDataSourceAwsVpcEndpointConfig_gatewayWithRouteTable = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-with-route-table"
  }
}

resource "aws_route_table" "rt" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.us-west-2.s3"
  route_table_ids = ["${aws_route_table.rt.id}"]
}

data "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "${aws_vpc_endpoint.s3.service_name}"
  state = "available"
}
`

const testAccDataSourceAwsVpcEndpointConfig_interface = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-data-source-interface"
  }
}

resource "aws_subnet" "sn" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${aws_vpc.foo.cidr_block}"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-vpc-endpoint-data-source-interface"
  }
}

resource "aws_security_group" "sg" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "ec2" {
  vpc_id = "${aws_vpc.foo.id}"
  vpc_endpoint_type = "Interface"
  service_name = "com.amazonaws.us-west-2.ec2"
  subnet_ids = ["${aws_subnet.sn.id}"]
  security_group_ids = ["${aws_security_group.sg.id}"]
  private_dns_enabled = false
}

data "aws_vpc_endpoint" "ec2" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "${aws_vpc_endpoint.ec2.service_name}"
  state = "available"
}
`
