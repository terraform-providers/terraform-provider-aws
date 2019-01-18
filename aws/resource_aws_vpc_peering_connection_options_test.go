package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSVpcPeeringConnectionOptions_importBasic(t *testing.T) {
	resourceName := "aws_vpc_peering_connection_options.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConnectionOptionsConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVpcPeeringConnectionOptions_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSVpcPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConnectionOptionsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_vpc_peering_connection_options.foo",
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"aws_vpc_peering_connection_options.foo",
						"accepter.1102046665.allow_remote_vpc_dns_resolution",
						"true",
					),
					testAccCheckAWSVpcPeeringConnectionOptions(
						"aws_vpc_peering_connection.foo",
						"accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					resource.TestCheckResourceAttr(
						"aws_vpc_peering_connection_options.foo",
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"aws_vpc_peering_connection_options.foo",
						"requester.41753983.allow_classic_link_to_remote_vpc",
						"true",
					),
					resource.TestCheckResourceAttr(
						"aws_vpc_peering_connection_options.foo",
						"requester.41753983.allow_vpc_to_remote_classic_link",
						"true",
					),
					testAccCheckAWSVpcPeeringConnectionOptions(
						"aws_vpc_peering_connection.foo",
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(true),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(true),
						},
					),
				),
			},
		},
	})
}

const testAccVpcPeeringConnectionOptionsConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-options-foo"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.1.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = "terraform-testacc-vpc-peering-conn-options-bar"
  }
}

resource "aws_vpc_peering_connection" "foo" {
  vpc_id      = "${aws_vpc.foo.id}"
  peer_vpc_id = "${aws_vpc.bar.id}"
  auto_accept = true
}

resource "aws_vpc_peering_connection_options" "foo" {
  vpc_peering_connection_id = "${aws_vpc_peering_connection.foo.id}"

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_vpc_to_remote_classic_link = true
    allow_classic_link_to_remote_vpc = true
  }
}
`
