package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_nat_gateway", &resource.Sweeper{
		Name: "aws_nat_gateway",
		F:    testSweepNatGateways,
	})
}

func testSweepNatGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	req := &ec2.DescribeNatGatewaysInput{
		Filter: []*ec2.Filter{
			{
				Name: aws.String("tag-value"),
				Values: []*string{
					aws.String("terraform-testacc-*"),
					aws.String("tf-acc-test-*"),
				},
			},
		},
	}
	resp, err := conn.DescribeNatGateways(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 NAT Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing NAT Gateways: %s", err)
	}

	if len(resp.NatGateways) == 0 {
		log.Print("[DEBUG] No AWS NAT Gateways to sweep")
		return nil
	}

	for _, natGateway := range resp.NatGateways {
		_, err := conn.DeleteNatGateway(&ec2.DeleteNatGatewayInput{
			NatGatewayId: natGateway.NatGatewayId,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting NAT Gateway (%s): %s",
				*natGateway.NatGatewayId, err)
		}
	}

	return nil
}

func TestAccAWSNatGateway_basic(t *testing.T) {
	var natGateway ec2.NatGateway
	resourceName := "aws_nat_gateway.gateway"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_nat_gateway.gateway",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckNatGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNatGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
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

func TestAccAWSNatGateway_tags(t *testing.T) {
	var natGateway ec2.NatGateway
	resourceName := "aws_nat_gateway.gateway"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNatGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNatGatewayConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
					testAccCheckTags(&natGateway.Tags, "Name", "terraform-testacc-nat-gw-tags"),
					testAccCheckTags(&natGateway.Tags, "foo", "bar"),
				),
			},

			{
				Config: testAccNatGatewayConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
					testAccCheckTags(&natGateway.Tags, "Name", "terraform-testacc-nat-gw-tags"),
					testAccCheckTags(&natGateway.Tags, "foo", ""),
					testAccCheckTags(&natGateway.Tags, "bar", "baz"),
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

func testAccCheckNatGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_nat_gateway" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			status := map[string]bool{
				"deleted":  true,
				"deleting": true,
				"failed":   true,
			}
			if _, ok := status[strings.ToLower(*resp.NatGateways[0].State)]; len(resp.NatGateways) > 0 && !ok {
				return fmt.Errorf("still exists")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NatGatewayNotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckNatGatewayExists(n string, ng *ec2.NatGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeNatGateways(&ec2.DescribeNatGatewaysInput{
			NatGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.NatGateways) == 0 {
			return fmt.Errorf("NatGateway not found")
		}

		*ng = *resp.NatGateways[0]

		return nil
	}
}

const testAccNatGatewayConfig = `
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-nat-gw-basic"
  }
}

resource "aws_subnet" "private" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.1.0/24"
  map_public_ip_on_launch = false
  tags = {
    Name = "tf-acc-nat-gw-basic-private"
  }
}

resource "aws_subnet" "public" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.2.0/24"
  map_public_ip_on_launch = true
  tags = {
    Name = "tf-acc-nat-gw-basic-public"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_eip" "nat_gateway" {
  vpc = true
}

// Actual SUT
resource "aws_nat_gateway" "gateway" {
  allocation_id = "${aws_eip.nat_gateway.id}"
  subnet_id = "${aws_subnet.public.id}"

  tags = {
    Name = "terraform-testacc-nat-gw-basic"
  }

  depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_route_table" "private" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = "${aws_nat_gateway.gateway.id}"
  }
}

resource "aws_route_table_association" "private" {
  subnet_id = "${aws_subnet.private.id}"
  route_table_id = "${aws_route_table.private.id}"
}

resource "aws_route_table" "public" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id = "${aws_subnet.public.id}"
  route_table_id = "${aws_route_table.public.id}"
}
`

const testAccNatGatewayConfigTags = `
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-nat-gw-tags"
  }
}

resource "aws_subnet" "private" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.1.0/24"
  map_public_ip_on_launch = false
  tags = {
    Name = "tf-acc-nat-gw-tags-private"
  }
}

resource "aws_subnet" "public" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.2.0/24"
  map_public_ip_on_launch = true
  tags = {
    Name = "tf-acc-nat-gw-tags-public"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_eip" "nat_gateway" {
  vpc = true
}

// Actual SUT
resource "aws_nat_gateway" "gateway" {
  allocation_id = "${aws_eip.nat_gateway.id}"
  subnet_id = "${aws_subnet.public.id}"

  tags = {
    Name = "terraform-testacc-nat-gw-tags"
    foo = "bar"
  }

  depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_route_table" "private" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = "${aws_nat_gateway.gateway.id}"
  }
}

resource "aws_route_table_association" "private" {
  subnet_id = "${aws_subnet.private.id}"
  route_table_id = "${aws_route_table.private.id}"
}

resource "aws_route_table" "public" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
  cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id = "${aws_subnet.public.id}"
  route_table_id = "${aws_route_table.public.id}"
}
`

const testAccNatGatewayConfigTagsUpdate = `
resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-nat-gw-tags"
  }
}

resource "aws_subnet" "private" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.1.0/24"
  map_public_ip_on_launch = false
  tags = {
    Name = "tf-acc-nat-gw-tags-private"
  }
}

resource "aws_subnet" "public" {
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "10.0.2.0/24"
  map_public_ip_on_launch = true
  tags = {
    Name = "tf-acc-nat-gw-tags-public"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_eip" "nat_gateway" {
  vpc = true
}

// Actual SUT
resource "aws_nat_gateway" "gateway" {
  allocation_id = "${aws_eip.nat_gateway.id}"
  subnet_id = "${aws_subnet.public.id}"

  tags = {
    Name = "terraform-testacc-nat-gw-tags"
    bar = "baz"
  }

  depends_on = ["aws_internet_gateway.gw"]
}

resource "aws_route_table" "private" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = "${aws_nat_gateway.gateway.id}"
  }
}

resource "aws_route_table_association" "private" {
  subnet_id = "${aws_subnet.private.id}"
  route_table_id = "${aws_route_table.private.id}"
}

resource "aws_route_table" "public" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_route_table_association" "public" {
  subnet_id = "${aws_subnet.public.id}"
  route_table_id = "${aws_route_table.public.id}"
}
`
