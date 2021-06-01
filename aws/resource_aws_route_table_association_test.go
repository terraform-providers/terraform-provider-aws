package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRouteTableAssociation_Subnet_basic(t *testing.T) {
	var rta ec2.RouteTableAssociation

	resourceNameAssoc := "aws_route_table_association.foo"
	resourceNameRouteTable := "aws_route_table.foo"
	resourceNameSubnet := "aws_subnet.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "subnet_id", resourceNameSubnet, "id"),
				),
			},
			{
				ResourceName:      resourceNameAssoc,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteTabAssocImportStateIdFunc(resourceNameAssoc),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTableAssociation_Subnet_ChangeRouteTable(t *testing.T) {
	var rta1, rta2 ec2.RouteTableAssociation

	resourceNameAssoc := "aws_route_table_association.foo"
	resourceNameRouteTable1 := "aws_route_table.foo"
	resourceNameRouteTable2 := "aws_route_table.bar"
	resourceNameSubnet := "aws_subnet.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta1),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable1, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "subnet_id", resourceNameSubnet, "id"),
				),
			},
			{
				Config: testAccRouteTableAssociationSubnetConfig_ChangeRouteTable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta2),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable2, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "subnet_id", resourceNameSubnet, "id"),
				),
			},
		},
	})
}

func TestAccAWSRouteTableAssociation_Subnet_ChangeSubnet(t *testing.T) {
	var rta1, rta2 ec2.RouteTableAssociation

	resourceNameAssoc := "aws_route_table_association.foo"
	resourceNameRouteTable := "aws_route_table.foo"
	resourceNameSubnet1 := "aws_subnet.foo"
	resourceNameSubnet2 := "aws_subnet.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta1),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "subnet_id", resourceNameSubnet1, "id"),
				),
			},
			{
				Config: testAccRouteTableAssociationSubnetConfig_ChangeSubnet,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta2),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "subnet_id", resourceNameSubnet2, "id"),
				),
			},
		},
	})
}

func TestAccAWSRouteTableAssociation_Gateway_basic(t *testing.T) {
	var rta ec2.RouteTableAssociation

	resourceNameAssoc := "aws_route_table_association.foo"
	resourceNameRouteTable := "aws_route_table.foo"
	resourceNameGateway := "aws_internet_gateway.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "gateway_id", resourceNameGateway, "id"),
				),
			},
			{
				ResourceName:      resourceNameAssoc,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteTabAssocImportStateIdFunc(resourceNameAssoc),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRouteTableAssociation_Gateway_ChangeRouteTable(t *testing.T) {
	var rta1, rta2 ec2.RouteTableAssociation

	resourceNameAssoc := "aws_route_table_association.foo"
	resourceNameRouteTable1 := "aws_route_table.foo"
	resourceNameRouteTable2 := "aws_route_table.bar"
	resourceNameGateway := "aws_internet_gateway.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta1),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable1, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "gateway_id", resourceNameGateway, "id"),
				),
			},
			{
				Config: testAccRouteTableAssociationGatewayConfig_ChangeRouteTable,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta2),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable2, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "gateway_id", resourceNameGateway, "id"),
				),
			},
		},
	})
}
func TestAccAWSRouteTableAssociation_Gateway_ChangeGateway(t *testing.T) {
	var rta1, rta2 ec2.RouteTableAssociation

	resourceNameAssoc := "aws_route_table_association.foo"
	resourceNameRouteTable := "aws_route_table.foo"
	resourceNameGateway1 := "aws_internet_gateway.foo"
	resourceNameGateway2 := "aws_vpn_gateway.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta1),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "gateway_id", resourceNameGateway1, "id"),
				),
			},
			{
				Config: testAccRouteTableAssociationGatewayConfig_ChangeGateway,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta2),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "route_table_id", resourceNameRouteTable, "id"),
					resource.TestCheckResourceAttrPair(resourceNameAssoc, "gateway_id", resourceNameGateway2, "id"),
				),
			},
		},
	})
}

func TestAccAWSRouteTableAssociation_disappears(t *testing.T) {
	var rta ec2.RouteTableAssociation

	resourceNameAssoc := "aws_route_table_association.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteTableAssociationSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableAssociationExists(resourceNameAssoc, &rta),
					testAccCheckRouteTableAssociationDisappears(&rta),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRouteTableAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route_table_association" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.Attributes["route_table_id"])},
		})
		if err != nil {
			// Verify the error is what we want
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() != "InvalidRouteTableID.NotFound" {
				return err
			}
			return nil
		}

		rt := resp.RouteTables[0]
		if len(rt.Associations) > 0 {
			return fmt.Errorf(
				"route table %s has associations", *rt.RouteTableId)
		}
	}

	return nil
}

func testAccCheckRouteTableAssociationExists(n string, rta *ec2.RouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.Attributes["route_table_id"])},
		})
		if err != nil {
			return err
		}
		if len(resp.RouteTables) == 0 {
			return fmt.Errorf("Route Table not found")
		}

		if len(resp.RouteTables[0].Associations) == 0 {
			return fmt.Errorf("no associations found for Route Table %q", rs.Primary.Attributes["route_table_id"])
		}

		found := false
		for _, association := range resp.RouteTables[0].Associations {
			if rs.Primary.ID == *association.RouteTableAssociationId {
				found = true
				*rta = *association
				break
			}
		}
		if !found {
			return fmt.Errorf("Association %q not found on Route Table %q", rs.Primary.ID, rs.Primary.Attributes["route_table_id"])
		}

		return nil
	}
}

func testAccCheckRouteTableAssociationDisappears(rta *ec2.RouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DisassociateRouteTable(&ec2.DisassociateRouteTableInput{
			AssociationId: rta.RouteTableAssociationId,
		})

		return err
	}
}

func testAccAWSRouteTabAssocImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		var target string
		if rs.Primary.Attributes["subnet_id"] != "" {
			target = rs.Primary.Attributes["subnet_id"]
		} else if rs.Primary.Attributes["gateway_id"] != "" {
			target = rs.Primary.Attributes["gateway_id"]
		}
		return fmt.Sprintf("%s/%s", target, rs.Primary.Attributes["route_table_id"]), nil
	}
}

const testAccRouteTableAssociationSubnetConfig = testAccRouteTableAssociatonCommonVpcConfig + `
resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.foo.id
  }

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_route_table_association" "foo" {
  route_table_id = aws_route_table.foo.id
  subnet_id      = aws_subnet.foo.id
}
`

const testAccRouteTableAssociationSubnetConfig_ChangeRouteTable = testAccRouteTableAssociatonCommonVpcConfig + `
resource "aws_route_table" "bar" {
  vpc_id = aws_vpc.foo.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.foo.id
  }

  tags = {
    Name = "tf-acc-route-update-table-assoc"
  }
}

resource "aws_route_table_association" "foo" {
  route_table_id = aws_route_table.bar.id
  subnet_id      = aws_subnet.foo.id
}
`

const testAccRouteTableAssociationSubnetConfig_ChangeSubnet = testAccRouteTableAssociatonCommonVpcConfig + `
resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id

  route {
    cidr_block = "10.0.0.0/8"
    gateway_id = aws_internet_gateway.foo.id
  }

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_subnet" "bar" {
  vpc_id     = aws_vpc.foo.id
  cidr_block = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_route_table_association" "foo" {
  route_table_id = aws_route_table.foo.id
  subnet_id      = aws_subnet.bar.id
}
`

const testAccRouteTableAssociationGatewayConfig = testAccRouteTableAssociatonCommonVpcConfig + `
resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id

  route {
    cidr_block           = aws_subnet.foo.cidr_block
    network_interface_id = aws_network_interface.appliance.id
  }

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_subnet" "appliance" {
  vpc_id     = aws_vpc.foo.id
  cidr_block = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_network_interface" "appliance" {
  subnet_id = aws_subnet.appliance.id
}

resource "aws_route_table_association" "foo" {
  route_table_id = aws_route_table.foo.id
  gateway_id     = aws_internet_gateway.foo.id
}
`

const testAccRouteTableAssociationGatewayConfig_ChangeRouteTable = testAccRouteTableAssociatonCommonVpcConfig + `
resource "aws_route_table" "bar" {
  vpc_id = aws_vpc.foo.id

  route {
    cidr_block           = aws_subnet.foo.cidr_block
    network_interface_id = aws_network_interface.appliance.id
  }

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_subnet" "appliance" {
  vpc_id     = aws_vpc.foo.id
  cidr_block = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_network_interface" "appliance" {
  subnet_id = aws_subnet.appliance.id
}

resource "aws_route_table_association" "foo" {
  route_table_id = aws_route_table.bar.id
  gateway_id     = aws_internet_gateway.foo.id
}
`

const testAccRouteTableAssociationGatewayConfig_ChangeGateway = testAccRouteTableAssociatonCommonVpcConfig + `
resource "aws_route_table" "foo" {
  vpc_id = aws_vpc.foo.id

  route {
    cidr_block           = aws_subnet.foo.cidr_block
    network_interface_id = aws_network_interface.appliance.id
  }

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_subnet" "appliance" {
  vpc_id     = aws_vpc.foo.id
  cidr_block = "10.1.2.0/24"

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_network_interface" "appliance" {
  subnet_id = aws_subnet.appliance.id
}

resource "aws_vpn_gateway" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_route_table_association" "foo" {
  route_table_id = aws_route_table.foo.id
  gateway_id     = aws_vpn_gateway.bar.id
}
`

const testAccRouteTableAssociatonCommonVpcConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_subnet" "foo" {
  vpc_id     = aws_vpc.foo.id
  cidr_block = "10.1.1.0/24"

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}

resource "aws_internet_gateway" "foo" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-route-table-assoc"
  }
}
`
