package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	req := &ec2.DescribeVpnGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-value"),
				Values: []*string{
					aws.String("terraform-testacc-*"),
					aws.String("tf-acc-test-*"),
				},
			},
		},
	}
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
		for _, vpcAttachment := range vpng.VpcAttachments {
			input := &ec2.DetachVpnGatewayInput{
				VpcId:        vpcAttachment.VpcId,
				VpnGatewayId: vpng.VpnGatewayId,
			}

			log.Printf("[DEBUG] Detaching VPN Gateway: %s", input)
			_, err := conn.DetachVpnGateway(input)
			if err != nil {
				return fmt.Errorf("error detaching VPN Gateway (%s) from VPC (%s): %s", aws.StringValue(vpng.VpnGatewayId), aws.StringValue(vpcAttachment.VpcId), err)
			}

			stateConf := &resource.StateChangeConf{
				Pending: []string{"attached", "detaching"},
				Target:  []string{"detached"},
				Refresh: vpnGatewayAttachmentStateRefresh(conn, aws.StringValue(vpcAttachment.VpcId), aws.StringValue(vpng.VpnGatewayId)),
				Timeout: 10 * time.Minute,
			}

			log.Printf("[DEBUG] Waiting for VPN Gateway (%s) to detach from VPC (%s)", aws.StringValue(vpng.VpnGatewayId), aws.StringValue(vpcAttachment.VpcId))
			if _, err = stateConf.WaitForState(); err != nil {
				return fmt.Errorf("error waiting for VPN Gateway (%s) to detach from VPC (%s): %s", aws.StringValue(vpng.VpnGatewayId), aws.StringValue(vpcAttachment.VpcId), err)
			}
		}

		input := &ec2.DeleteVpnGatewayInput{
			VpnGatewayId: vpng.VpnGatewayId,
		}

		log.Printf("[DEBUG] Deleting VPN Gateway: %s", input)
		_, err := conn.DeleteVpnGateway(input)
		if err != nil {
			return fmt.Errorf("error deleting VPN Gateway (%s): %s", aws.StringValue(vpng.VpnGatewayId), err)
		}
	}

	return nil
}

func TestAccAWSVpnGateway_importBasic(t *testing.T) {
	resourceName := "aws_vpn_gateway.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVpnGateway_basic(t *testing.T) {
	var v, v2 ec2.VpnGateway

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
		IDRefreshName: "aws_vpn_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.foo", &v),
				),
			},

			{
				Config: testAccVpnGatewayConfigChangeVPC,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.foo", &v2),
					testNotEqual,
				),
			},
		},
	})
}

func TestAccAWSVpnGateway_withAvailabilityZoneSetToState(t *testing.T) {
	var v ec2.VpnGateway

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfigWithAZ,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists("aws_vpn_gateway.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_vpn_gateway.foo", "availability_zone", "us-west-2a"),
				),
			},
		},
	})
}
func TestAccAWSVpnGateway_withAmazonSideAsnSetToState(t *testing.T) {
	var v ec2.VpnGateway

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfigWithASN,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists("aws_vpn_gateway.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_vpn_gateway.foo", "amazon_side_asn", "4294967294"),
				),
			},
		},
	})
}

func TestAccAWSVpnGateway_disappears(t *testing.T) {
	var v ec2.VpnGateway

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists("aws_vpn_gateway.foo", &v),
					testAccAWSVpnGatewayDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpnGateway_reattach(t *testing.T) {
	var vpc1, vpc2 ec2.Vpc
	var vgw1, vgw2 ec2.VpnGateway

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
		IDRefreshName: "aws_vpn_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVpnGatewayConfigReattach,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc1),
					testAccCheckVpcExists("aws_vpc.bar", &vpc2),
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.foo", &vgw1),
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.bar", &vgw2),
					testAttachmentFunc(&vgw1, &vpc1),
					testAttachmentFunc(&vgw2, &vpc2),
				),
			},
			{
				Config: testAccCheckVpnGatewayConfigReattachChange,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.foo", &vgw1),
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.bar", &vgw2),
					testAttachmentFunc(&vgw2, &vpc1),
					testAttachmentFunc(&vgw1, &vpc2),
				),
			},
			{
				Config: testAccCheckVpnGatewayConfigReattach,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.foo", &vgw1),
					testAccCheckVpnGatewayExists(
						"aws_vpn_gateway.bar", &vgw2),
					testAttachmentFunc(&vgw1, &vpc1),
					testAttachmentFunc(&vgw2, &vpc2),
				),
			},
		},
	})
}

func TestAccAWSVpnGateway_delete(t *testing.T) {
	var vpnGateway ec2.VpnGateway

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
		IDRefreshName: "aws_vpn_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpnGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists("aws_vpn_gateway.foo", &vpnGateway)),
			},
			{
				Config: testAccNoVpnGatewayConfig,
				Check:  resource.ComposeTestCheckFunc(testDeleted("aws_vpn_gateway.foo")),
			},
		},
	})
}

func TestAccAWSVpnGateway_tags(t *testing.T) {
	var v ec2.VpnGateway

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpn_gateway.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVpnGatewayConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists("aws_vpn_gateway.foo", &v),
					testAccCheckTags(&v.Tags, "Name", "terraform-testacc-vpn-gateway-tags"),
				),
			},
			{
				Config: testAccCheckVpnGatewayConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpnGatewayExists("aws_vpn_gateway.foo", &v),
					testAccCheckTags(&v.Tags, "foo", ""),
					testAccCheckTags(&v.Tags, "Name", "terraform-testacc-vpn-gateway-tags-updated"),
				),
			},
		},
	})
}

func testAccAWSVpnGatewayDisappears(gateway *ec2.VpnGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DetachVpnGateway(&ec2.DetachVpnGatewayInput{
			VpnGatewayId: gateway.VpnGatewayId,
			VpcId:        gateway.VpcAttachments[0].VpcId,
		})
		if err != nil {
			ec2err, ok := err.(awserr.Error)
			if ok {
				if ec2err.Code() == "InvalidVpnGatewayID.NotFound" {
					return nil
				} else if ec2err.Code() == "InvalidVpnGatewayAttachment.NotFound" {
					return nil
				}
			}

			if err != nil {
				return err
			}
		}

		opts := &ec2.DeleteVpnGatewayInput{
			VpnGatewayId: gateway.VpnGatewayId,
		}
		if _, err := conn.DeleteVpnGateway(opts); err != nil {
			return err
		}
		return resource.Retry(40*time.Minute, func() *resource.RetryError {
			opts := &ec2.DescribeVpnGatewaysInput{
				VpnGatewayIds: []*string{gateway.VpnGatewayId},
			}
			resp, err := conn.DescribeVpnGateways(opts)
			if err != nil {
				cgw, ok := err.(awserr.Error)
				if ok && cgw.Code() == "InvalidVpnGatewayID.NotFound" {
					return nil
				}
				if ok && cgw.Code() == "IncorrectState" {
					return resource.RetryableError(fmt.Errorf(
						"Waiting for VPN Gateway to be in the correct state: %v", gateway.VpnGatewayId))
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error retrieving VPN Gateway: %s", err))
			}
			if *resp.VpnGateways[0].State == "deleted" {
				return nil
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for VPN Gateway: %v", gateway.VpnGatewayId))
		})
	}
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
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-removed"
  }
}
`

const testAccVpnGatewayConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-basic"
  }
}
`

const testAccVpnGatewayConfigChangeVPC = `
resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-change-vpc"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.bar.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-basic"
  }
}
`

const testAccCheckVpnGatewayConfigTags = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-tags"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-tags"
  }
}
`

const testAccCheckVpnGatewayConfigTagsUpdate = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-tags"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-tags-updated"
  }
}
`

const testAccCheckVpnGatewayConfigReattach = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-foo"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-bar"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}

resource "aws_vpn_gateway" "bar" {
  vpc_id = "${aws_vpc.bar.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}
`

const testAccCheckVpnGatewayConfigReattachChange = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-foo"
  }
}

resource "aws_vpc" "bar" {
  cidr_block = "10.2.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach-bar"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.bar.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}

resource "aws_vpn_gateway" "bar" {
  vpc_id = "${aws_vpc.foo.id}"
  tags = {
    Name = "terraform-testacc-vpn-gateway-reattach"
  }
}
`

const testAccVpnGatewayConfigWithAZ = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-with-az"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  availability_zone = "us-west-2a"
  tags = {
    Name = "terraform-testacc-vpn-gateway-with-az"
  }
}
`

const testAccVpnGatewayConfigWithASN = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-vpn-gateway-with-asn"
  }
}

resource "aws_vpn_gateway" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  amazon_side_asn = 4294967294
}
`
