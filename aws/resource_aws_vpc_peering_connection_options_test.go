package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func TestAccAWSVpcPeeringConnectionOptions_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-testacc-pcxopts-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	resourceName := "aws_vpc_peering_connection_options.test"
	pcxResourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:            func() { testAccPreCheck(t) },
		Providers:           testAccProviders,
		CheckDestroy:        testAccCheckAWSVpcPeeringConnectionDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConnectionOptionsConfig_sameRegion_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.41753983.allow_remote_vpc_dns_resolution",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.41753983.allow_classic_link_to_remote_vpc",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.41753983.allow_vpc_to_remote_classic_link",
						"true",
					),
					testAccCheckAWSVpcPeeringConnectionOptions(
						pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(true),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(true),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.1102046665.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.1102046665.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.1102046665.allow_vpc_to_remote_classic_link",
						"false",
					),
					testAccCheckAWSVpcPeeringConnectionOptions(
						pcxResourceName,
						"accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVpcPeeringConnectionOptions_differentRegionSameAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := fmt.Sprintf("tf-testacc-pcxopts-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	resourceName := "aws_vpc_peering_connection_options.test"         // Requester
	resourceNamePeer := "aws_vpc_peering_connection_options.peer"     // Accepter
	pcxResourceName := "aws_vpc_peering_connection.test"              // Requester
	pcxResourceNamePeer := "aws_vpc_peering_connection_accepter.peer" // Accepter

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories:   testAccProviderFactories(&providers),
		CheckDestroy:        testAccCheckAWSVpcPeeringConnectionDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConnectionOptionsConfig_differentRegion_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.1102046665.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.1102046665.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.1102046665.allow_vpc_to_remote_classic_link",
						"false",
					),
					testAccCheckAWSVpcPeeringConnectionOptions(
						pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.1102046665.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.1102046665.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.1102046665.allow_vpc_to_remote_classic_link",
						"false",
					),
					testAccCheckAWSVpcPeeringConnectionOptionsWithProvider(
						pcxResourceNamePeer,
						"accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
						testAccAwsRegionProviderFunc(testAccGetAlternateRegion(), &providers),
					),
				),
			},
			{
				Config:            testAccVpcPeeringConnectionOptionsConfig_differentRegion_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVpcPeeringConnectionOptions_sameRegionDifferentAccount(t *testing.T) {
	var providers []*schema.Provider
	rName := fmt.Sprintf("tf-testacc-pcxopts-%s", acctest.RandStringFromCharSet(13, acctest.CharSetAlphaNum))
	resourceName := "aws_vpc_peering_connection_options.test"     // Requester
	resourceNamePeer := "aws_vpc_peering_connection_options.peer" // Accepter
	pcxResourceName := "aws_vpc_peering_connection.test"          // Requester

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories:   testAccProviderFactories(&providers),
		CheckDestroy:        testAccCheckAWSVpcPeeringConnectionDestroy,
		DisableBinaryDriver: true,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConnectionOptionsConfig_sameRegion_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.1102046665.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.1102046665.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.1102046665.allow_vpc_to_remote_classic_link",
						"false",
					),
					testAccCheckAWSVpcPeeringConnectionOptions(
						pcxResourceName,
						"requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.1102046665.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.1102046665.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceNamePeer,
						"accepter.1102046665.allow_vpc_to_remote_classic_link",
						"false",
					),
				),
			},
			{
				Config:            testAccVpcPeeringConnectionOptionsConfig_sameRegion_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVpcPeeringConnectionOptionsConfig_sameRegion_sameAccount(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.peer.id}"
  auto_accept = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection_options" "test" {
  vpc_peering_connection_id = "${aws_vpc_peering_connection.test.id}"

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_vpc_to_remote_classic_link = true
    allow_classic_link_to_remote_vpc = true
  }
}
`, rName)
}

func testAccVpcPeeringConnectionOptionsConfig_differentRegion_sameAccount(rName string) string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "aws.alternate"

  cidr_block = "10.1.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

// Requester's side of the connection.
resource "aws_vpc_peering_connection" "test" {
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.peer.id}"
  auto_accept = false
  peer_region = %[2]q
  tags = {
    Name = %[1]q
  }
}

// Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "aws.alternate"

  vpc_peering_connection_id = "${aws_vpc_peering_connection.test.id}"
  auto_accept = true
  tags = {
    Name = %[1]q
  }
}

// Requester's side of the connection.
resource "aws_vpc_peering_connection_options" "test" {
  # As options can't be set until the connection has been accepted
  # create an explicit dependency on the accepter.
  vpc_peering_connection_id = "${aws_vpc_peering_connection_accepter.peer.id}"

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

// Accepter's side of the connection.
resource "aws_vpc_peering_connection_options" "peer" {
  provider = "aws.alternate"

  vpc_peering_connection_id = "${aws_vpc_peering_connection_accepter.peer.id}"

  accepter {
    allow_remote_vpc_dns_resolution = true
  }
}
`, rName, testAccGetAlternateRegion())
}

func testAccVpcPeeringConnectionOptionsConfig_sameRegion_differentAccount(rName string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "aws.alternate"

  cidr_block = "10.1.0.0/16"
  enable_dns_hostnames = true
  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "aws.alternate"
}

// Requester's side of the connection.
resource "aws_vpc_peering_connection" "test" {
  vpc_id = "${aws_vpc.test.id}"
  peer_vpc_id = "${aws_vpc.peer.id}"
  peer_owner_id = "${data.aws_caller_identity.peer.account_id}"
  auto_accept = false
  tags = {
    Name = %[1]q
  }
}

 // Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "aws.alternate"

  vpc_peering_connection_id = "${aws_vpc_peering_connection.test.id}"
  auto_accept = true
  tags = {
    Name = %[1]q
  }
}

// Requester's side of the connection.
resource "aws_vpc_peering_connection_options" "test" {
  # As options can't be set until the connection has been accepted
  # create an explicit dependency on the accepter.
  vpc_peering_connection_id = "${aws_vpc_peering_connection_accepter.peer.id}"

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

// Accepter's side of the connection.
resource "aws_vpc_peering_connection_options" "peer" {
  provider = "aws.alternate"

  vpc_peering_connection_id = "${aws_vpc_peering_connection_accepter.peer.id}"

  accepter {
    allow_remote_vpc_dns_resolution = true
  }
}
`, rName)
}
