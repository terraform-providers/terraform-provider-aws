package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVPCPeeringConnectionAccepter_sameRegionSameAccount(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVPCPeeringConnectionAccepterConfigSameRegionSameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(resourceNameAccepter, &connection),
					// The aws_vpc_peering_connection documentation says:
					//	vpc_id - The ID of the requester VPC
					//	peer_vpc_id - The ID of the VPC with which you are creating the VPC Peering Connection (accepter)
					//	peer_owner_id -  The AWS account ID of the owner of the peer VPC (accepter)
					//	peer_region -  The region of the accepter VPC of the VPC Peering Connection
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", testAccGetRegion()),
					// The aws_vpc_peering_connection_accepter documentation says:
					//	vpc_id - The ID of the accepter VPC
					//	peer_vpc_id - The ID of the requester VPC
					//	peer_owner_id - The AWS account ID of the owner of the requester VPC
					//	peer_region - The region of the accepter VPC
					// ** TODO
					// ** TODO resourceAwsVPCPeeringRead() is not doing this correctly for same-account peerings
					// ** TODO
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
			{
				Config:                  testAccAwsVPCPeeringConnectionAccepterConfigSameRegionSameAccount(rName),
				ResourceName:            resourceNameAccepter,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccAWSVPCPeeringConnectionAccepter_differentRegionSameAccount(t *testing.T) {
	var connectionMain, connectionPeer ec2.VpcPeeringConnection
	var providers []*schema.Provider
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccAwsVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVPCPeeringConnectionAccepterConfigDifferentRegionSameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(resourceNameConnection, &connectionMain),
					testAccCheckAWSVpcPeeringConnectionExistsWithProvider(resourceNameAccepter, &connectionPeer, testAccAwsRegionProviderFunc(testAccGetAlternateRegion(), &providers)),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", testAccGetAlternateRegion()),
					// ** TODO See TestAccAWSVPCPeeringConnectionAccepter_sameRegion()
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					// resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", testAccGetAlternateRegion()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
			{
				Config:                  testAccAwsVPCPeeringConnectionAccepterConfigDifferentRegionSameAccount(rName),
				ResourceName:            resourceNameAccepter,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auto_accept"},
			},
		},
	})
}

func TestAccAWSVPCPeeringConnectionAccepter_sameRegionDifferentAccount(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	var providers []*schema.Provider
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccAwsVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVPCPeeringConnectionAccepterConfigSameRegionDifferentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(resourceNameConnection, &connection),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
		},
	})
}

func TestAccAWSVPCPeeringConnectionAccepter_differentRegionDifferentAccount(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	var providers []*schema.Provider
	resourceNameMainVpc := "aws_vpc.main"                              // Requester
	resourceNamePeerVpc := "aws_vpc.peer"                              // Accepter
	resourceNameConnection := "aws_vpc_peering_connection.main"        // Requester
	resourceNameAccepter := "aws_vpc_peering_connection_accepter.peer" // Accepter
	rName := fmt.Sprintf("terraform-testacc-pcxaccpt-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccAwsVPCPeeringConnectionAccepterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVPCPeeringConnectionAccepterConfigDifferentRegionDifferentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSVpcPeeringConnectionExists(resourceNameConnection, &connection),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameConnection, "peer_owner_id", resourceNamePeerVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameConnection, "peer_region", testAccGetAlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "vpc_id", resourceNamePeerVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_vpc_id", resourceNameMainVpc, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAccepter, "peer_owner_id", resourceNameMainVpc, "owner_id"),
					resource.TestCheckResourceAttr(resourceNameAccepter, "peer_region", testAccGetAlternateRegion()),
					resource.TestCheckResourceAttr(resourceNameAccepter, "accept_status", "active"),
				),
			},
		},
	})
}

func testAccAwsVPCPeeringConnectionAccepterDestroy(s *terraform.State) error {
	// We don't destroy the underlying VPC Peering Connection.
	return nil
}

func testAccAwsVPCPeeringConnectionAccepterConfigSameRegionSameAccount(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id      = aws_vpc.main.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAwsVPCPeeringConnectionAccepterConfigDifferentRegionSameAccount(rName string) string {
	return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id      = aws_vpc.main.id
  peer_vpc_id = aws_vpc.peer.id
  peer_region = %[2]q
  auto_accept = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName, testAccGetAlternateRegion())
}

func testAccAwsVPCPeeringConnectionAccepterConfigSameRegionDifferentAccount(rName string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = %[2]q
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName, testAccGetRegion())
}

func testAccAwsVPCPeeringConnectionAccepterConfigDifferentRegionDifferentAccount(rName string) string {
	return testAccAlternateAccountAlternateRegionProviderConfig() + fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "peer" {
  provider = "awsalternate"
}

# Requester's side of the connection.
resource "aws_vpc_peering_connection" "main" {
  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = aws_vpc.peer.id
  peer_owner_id = data.aws_caller_identity.peer.account_id
  peer_region   = %[2]q
  auto_accept   = false

  tags = {
    Name = %[1]q
  }
}

# Accepter's side of the connection.
resource "aws_vpc_peering_connection_accepter" "peer" {
  provider = "awsalternate"

  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
  auto_accept               = true

  tags = {
    Name = %[1]q
  }
}
`, rName, testAccGetAlternateRegion())
}
