package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDefaultVpc_basic(t *testing.T) {
	var vpc ec2.Vpc

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultVpcConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_default_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "172.31.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "cidr_block", "172.31.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "tags.Name", "Default VPC"),
					resource.TestCheckResourceAttrSet(
						"aws_default_vpc.foo", "arn"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "ipv6_association_id", ""),
					resource.TestCheckResourceAttr(
						"aws_default_vpc.foo", "ipv6_cidr_block", ""),
					testAccCheckResourceAttrAccountID("aws_default_vpc.foo", "owner_id"),
				),
			},
		},
	})
}

func testAccCheckAWSDefaultVpcDestroy(s *terraform.State) error {
	// We expect VPC to still exist
	return nil
}

const testAccAWSDefaultVpcConfigBasic = `
provider "aws" {
    region = "us-west-2"
}

resource "aws_default_vpc" "foo" {
	tags = {
		Name = "Default VPC"
	}
}
`
