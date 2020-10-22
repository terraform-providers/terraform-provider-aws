package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// add sweeper to delete known test VPN Gateways
func init() {
	resource.AddTestSweepers("aws_vpn_gateway", &resource.Sweeper{
		Name: "aws_vpn_gateway",
		F:    testSweepVPNGateways,
		Dependencies: []string{
			"aws_dx_gateway_association",
		},
	})
}

func testSweepVPNGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	var sweeperErrs *multierror.Error

	req := &ec2.DescribeVpnGatewaysInput{}
	resp, err := conn.DescribeVpnGateways(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPN Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing VPN Gateways: %s", err)
	}

	if len(resp.VpnGateways) == 0 {
		log.Print("[DEBUG] No VPN Gateways to sweep")
		return nil
	}

	for _, vpng := range resp.VpnGateways {
		if aws.StringValue(vpng.State) == ec2.VpnStateDeleted {
			continue
		}

		for _, vpcAttachment := range vpng.VpcAttachments {
			if aws.StringValue(vpcAttachment.State) == ec2.AttachmentStatusDetached {
				continue
			}

			r := resourceAwsVpnGatewayAttachment()
			d := r.Data(nil)
			d.Set("vpc_id", vpcAttachment.VpcId)
			d.Set("vpn_gateway_id", vpng.VpnGatewayId)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		r := resourceAwsVpnGateway()
		d := r.Data(nil)
		d.SetId(aws.StringValue(vpng.VpnGatewayId))
		err := r.Delete(d, client)

		if err != nil {
			log.Printf("[ERROR] %s", err)
			sweeperErrs = multierror.Append(sweeperErrs, err)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSVpnGateway_basic(t *testing.T) {
	var v, v2 ec2.VpnGateway
	resourceName := "aws_vpn_gateway.test"

	testNotEqual := func(*terraform.State) error {
		if len(v.VpcAttachments) == 0 {
			return fmt.Errorf("VPN Gateway A is not attached")
		}
		if len(v2.VpcAttachments) == 0 {
			return fmt.Errorf("VPN Gateway B is not attached")
		}

		id1 := v.VpcAttachments[0].VpcId
		id2 := v2.VpcAttachments[0].VpcId
		if id1 == id2 {
			return fmt.Errorf("Both attachment IDs are the same")
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpn-gateway/vgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpnGatewayConfigChangeVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v2),
					testNotEqual,
				),
			},
		},
	})
}

func TestAccAWSVpnGateway_withAvailabilityZoneSetToState(t *testing.T) {
	var v ec2.VpnGateway
	resourceName := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfigWithAZ,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "availability_zone", "us-west-2a"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"availability_zone"},
			},
		},
	})
}
func TestAccAWSVpnGateway_withAmazonSideAsnSetToState(t *testing.T) {
	var v ec2.VpnGateway
	resourceName := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfigWithASN,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "amazon_side_asn", "4294967294"),
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

func TestAccAWSVpnGateway_disappears(t *testing.T) {
	var v ec2.VpnGateway
	resourceName := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpnGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpnGateway_reattach(t *testing.T) {
	var vpc1, vpc2 ec2.Vpc
	var vgw1, vgw2 ec2.VpnGateway
	resourceName := "aws_vpn_gateway.test"
	resourceName2 := "aws_vpn_gateway.test2"

	testAttachmentFunc := func(vgw *ec2.VpnGateway, vpc *ec2.Vpc) func(*terraform.State) error {
		return func(*terraform.State) error {
			if len(vgw.VpcAttachments) == 0 {
				return fmt.Errorf("VPN Gateway %q has no VPC attachments.",
					*vgw.VpnGatewayId)
			}

			if len(vgw.VpcAttachments) > 1 {
				count := 0
				for _, v := range vgw.VpcAttachments {
					if *v.State == "attached" {
						count += 1
					}
				}
				if count > 1 {
					return fmt.Errorf(
						"VPN Gateway %q has an unexpected number of VPC attachments (more than 1): %#v",
						*vgw.VpnGatewayId, vgw.VpcAttachments)
				}
			}

			if *vgw.VpcAttachments[0].State != "attached" {
				return fmt.Errorf("Expected VPN Gateway %q to be attached.",
					*vgw.VpnGatewayId)
			}

			if *vgw.VpcAttachments[0].VpcId != *vpc.VpcId {
				return fmt.Errorf("Expected VPN Gateway %q to be attached to VPC %q, but got: %q",
					*vgw.VpnGatewayId, *vpc.VpcId, *vgw.VpcAttachments[0].VpcId)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVpnGatewayConfigReattach,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.test", &vpc1),
					testAccCheckVpcExists("aws_vpc.test2", &vpc2),
					testAccCheckVpnGatewayExists(
						resourceName, &vgw1),
					testAccCheckVpnGatewayExists(
						resourceName2, &vgw2),
					testAttachmentFunc(&vgw1, &vpc1),
					testAttachmentFunc(&vgw2, &vpc2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckVpnGatewayConfigReattachChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(
						resourceName, &vgw1),
					testAccCheckVpnGatewayExists(
						resourceName2, &vgw2),
					testAttachmentFunc(&vgw2, &vpc1),
					testAttachmentFunc(&vgw1, &vpc2),
				),
			},
			{
				Config: testAccCheckVpnGatewayConfigReattach,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(
						resourceName, &vgw1),
					testAccCheckVpnGatewayExists(
						resourceName2, &vgw2),
					testAttachmentFunc(&vgw1, &vpc1),
					testAttachmentFunc(&vgw2, &vpc2),
				),
			},
		},
	})
}

func TestAccAWSVpnGateway_delete(t *testing.T) {
	var vpnGateway ec2.VpnGateway
	resourceName := "aws_vpn_gateway.test"

	testDeleted := func(r string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			_, ok := s.RootModule().Resources[r]
			if ok {
				return fmt.Errorf("VPN Gateway %q should have been deleted.", r)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &vpnGateway)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccNoVpnGatewayConfig,
				Check:  resource.ComposeTestCheckFunc(testDeleted(resourceName)),
			},
		},
	})
}

func TestAccAWSVpnGateway_tags(t *testing.T) {
	var v ec2.VpnGateway
	resourceName := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVpnGatewayConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v),
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
				Config: testAccCheckVpnGatewayConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCheckVpnGatewayConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckVpnGatewayDestroy(s *terraform.State) error {
	ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_gateway" {
			continue
		}

		// Try to find the resource
		resp, err := ec2conn.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{
			VpnGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			var v *ec2.VpnGateway
			for _, g := range resp.VpnGateways {
				if *g.VpnGatewayId == rs.Primary.ID {
					v = g
				}
			}

			if v == nil {
				// wasn't found
				return nil
			}

			if *v.State != "deleted" {
				return fmt.Errorf("Expected VPN Gateway to be in deleted state, but was not: %s", v)
			}
			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpnGatewayID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckVpnGatewayExists(n string, ig *ec2.VpnGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := ec2conn.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{
			VpnGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.VpnGateways) == 0 {
			return fmt.Errorf("VPN Gateway not found")
		}

		*ig = *resp.VpnGateways[0]

		return nil
	}
}

const testAccNoVpnGatewayConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-removed"
  }
}
`

const testAccVpnGatewayConfig = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id
}
`

const testAccVpnGatewayConfigChangeVPC = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway"
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-change-vpc"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test2.id

  tags = {
    Name = "terraform-testacc-vpn-gateway-basic"
  }
}
`

func testAccCheckVpnGatewayConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-tags"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccCheckVpnGatewayConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-tags"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccCheckVpnGatewayConfigReattach = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-test"
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-test2"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}

resource "aws_vpn_gateway" "test2" {
  vpc_id = aws_vpc.test2.id

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}
`

const testAccCheckVpnGatewayConfigReattachChange = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-test"
  }
}

resource "aws_vpc" "test2" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-test2"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test2.id

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}

resource "aws_vpn_gateway" "test2" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}
`

const testAccVpnGatewayConfigWithAZ = `
data "aws_availability_zones" "azs" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-with-az"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.azs.names[0]

  tags = {
    Name = "terraform-testacc-vpn-gateway-with-az"
  }
}
`

const testAccVpnGatewayConfigWithASN = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-with-asn"
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id          = aws_vpc.test.id
  amazon_side_asn = 4294967294
}
`
