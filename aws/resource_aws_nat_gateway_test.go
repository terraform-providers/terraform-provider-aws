package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

	req := &ec2.DescribeNatGatewaysInput{}
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
	resourceName := "aws_nat_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckNatGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNatGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
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
	resourceName := "aws_nat_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNatGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNatGatewayConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNatGatewayConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNatGatewayConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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
				ec2.NatGatewayStateDeleted:  true,
				ec2.NatGatewayStateDeleting: true,
				ec2.NatGatewayStateFailed:   true,
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

const testAccNatGatewayConfigBase = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-nat-gw-basic"
  }
}

resource "aws_subnet" "private" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = false

  tags = {
    Name = "tf-acc-nat-gw-basic-private"
  }
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.2.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-nat-gw-basic-public"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_eip" "test" {
  vpc = true
}
`

const testAccNatGatewayConfig = testAccNatGatewayConfigBase + `
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  depends_on = [aws_internet_gateway.test]
}
`

func testAccNatGatewayConfigTags1(tagKey1, tagValue1 string) string {
	return testAccNatGatewayConfigBase + fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    %[1]q = %[2]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, tagKey1, tagValue1)
}

func testAccNatGatewayConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccNatGatewayConfigBase + fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.public.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }

  depends_on = [aws_internet_gateway.test]
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
