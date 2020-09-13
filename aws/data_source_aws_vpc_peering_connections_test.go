package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsVpcPeeringConnections_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcPeeringConnectionsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_vpc_peering_connections.test_by_filters", "ids.#", "2"),
				),
			},
		},
	})
}

const testAccDataSourceAwsVpcPeeringConnectionsConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source-foo"
    Type = "primary"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source-bar"
    Type = "secondary"
  }
}

resource "aws_vpc" "baz" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source-baz"
    Type = "secondary"
  }
}

resource "aws_vpc_peering_connection" "conn1" {
  vpc_id      = aws_vpc.foo.id
  peer_vpc_id = aws_vpc.bar.id
  auto_accept = true

  tags = {
    Name        = "terraform-testacc-vpc-peering-connection-data-source-foo-to-bar"
    Environment = "test"
  }
}

resource "aws_vpc_peering_connection" "conn2" {
  vpc_id      = aws_vpc.foo.id
  peer_vpc_id = aws_vpc.baz.id
  auto_accept = true

  tags = {
    Name        = "terraform-testacc-vpc-peering-connection-data-source-foo-to-baz"
    Environment = "test"
  }
}

data "aws_vpc_peering_connections" "test_by_filters" {
  filter {
    name   = "vpc-peering-connection-id"
    values = [aws_vpc_peering_connection.conn1.id, aws_vpc_peering_connection.conn2.id]
  }
}
`
