package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsInternetGateway_typical(t *testing.T) {
	igwResourceName := "aws_internet_gateway.test"
	vpcResourceName := "aws_vpc.test"
	ds1ResourceName := "data.aws_internet_gateway.by_id"
	ds2ResourceName := "data.aws_internet_gateway.by_filter"
	ds3ResourceName := "data.aws_internet_gateway.by_tags"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1ResourceName, "internet_gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "owner_id", igwResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds1ResourceName, "attachments.0.vpc_id", vpcResourceName, "id"),

					resource.TestCheckResourceAttrPair(ds2ResourceName, "internet_gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "owner_id", igwResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "attachments.0.vpc_id", vpcResourceName, "id"),

					resource.TestCheckResourceAttrPair(ds3ResourceName, "internet_gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "owner_id", igwResourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(ds3ResourceName, "attachments.0.vpc_id", vpcResourceName, "id"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

const testAccDataSourceAwsInternetGatewayConfig = `
provider "aws" {
  region = "eu-central-1"
}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-igw-data-source"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "terraform-testacc-data-source-igw"
  }
}

data "aws_internet_gateway" "by_id" {
  internet_gateway_id = "${aws_internet_gateway.test.id}"
}

data "aws_internet_gateway" "by_tags" {
  tags = {
    Name = "${aws_internet_gateway.test.tags["Name"]}"
  }
}

data "aws_internet_gateway" "by_filter" {
  filter {
    name = "attachment.vpc-id"
    values = ["${aws_vpc.test.id}"]
  }

  depends_on = ["aws_internet_gateway.test"]
}
`
