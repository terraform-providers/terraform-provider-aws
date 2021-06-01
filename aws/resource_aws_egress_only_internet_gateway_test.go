package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_egress_only_internet_gateway", &resource.Sweeper{
		Name: "aws_egress_only_internet_gateway",
		F:    testSweepEc2EgressOnlyInternetGateways,
	})
}

func testSweepEc2EgressOnlyInternetGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	input := &ec2.DescribeEgressOnlyInternetGatewaysInput{}
	err = conn.DescribeEgressOnlyInternetGatewaysPages(input, func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gateway := range page.EgressOnlyInternetGateways {
			id := aws.StringValue(gateway.EgressOnlyInternetGatewayId)
			input := &ec2.DeleteEgressOnlyInternetGatewayInput{
				EgressOnlyInternetGatewayId: gateway.EgressOnlyInternetGatewayId,
			}

			log.Printf("[INFO] Deleting EC2 Egress Only Internet Gateway: %s", id)

			_, err := conn.DeleteEgressOnlyInternetGateway(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 Egress Only Internet Gateway (%s): %s", id, err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Egress Only Internet Gateway sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing EC2 Egress Only Internet Gateways: %s", err)
	}

	return nil
}

func TestAccAWSEgressOnlyInternetGateway_basic(t *testing.T) {
	var igw ec2.EgressOnlyInternetGateway
	resourceName := "aws_egress_only_internet_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEgressOnlyInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEgressOnlyInternetGatewayConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSEgressOnlyInternetGatewayExists(resourceName, &igw),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSEgressOnlyInternetGateway_Tags(t *testing.T) {
	var v ec2.EgressOnlyInternetGateway
	resourceName := "aws_egress_only_internet_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEgressOnlyInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEgressOnlyInternetGatewayConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEgressOnlyInternetGatewayExists(resourceName, &v),
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
				Config: testAccAWSEgressOnlyInternetGatewayConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEgressOnlyInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEgressOnlyInternetGatewayConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEgressOnlyInternetGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSEgressOnlyInternetGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_egress_only_internet_gateway" {
			continue
		}

		describe, err := conn.DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{
			EgressOnlyInternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(describe.EgressOnlyInternetGateways) != 0 &&
				*describe.EgressOnlyInternetGateways[0].EgressOnlyInternetGatewayId == rs.Primary.ID {
				return fmt.Errorf("Egress Only Internet Gateway %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}

	return nil
}

func testAccCheckAWSEgressOnlyInternetGatewayExists(n string, igw *ec2.EgressOnlyInternetGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Egress Only IGW ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeEgressOnlyInternetGateways(&ec2.DescribeEgressOnlyInternetGatewaysInput{
			EgressOnlyInternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.EgressOnlyInternetGateways) == 0 {
			return fmt.Errorf("Egress Only IGW not found")
		}

		*igw = *resp.EgressOnlyInternetGateways[0]

		return nil
	}
}

const testAccAWSEgressOnlyInternetGatewayConfig_basic = `
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-egress-only-igw-basic"
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}
`

func testAccAWSEgressOnlyInternetGatewayConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-egress-only-igw-tags"
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSEgressOnlyInternetGatewayConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-egress-only-igw-tags"
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
