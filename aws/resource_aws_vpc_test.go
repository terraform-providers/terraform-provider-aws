package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// add sweeper to delete known test vpcs
func init() {
	resource.AddTestSweepers("aws_vpc", &resource.Sweeper{
		Name: "aws_vpc",
		Dependencies: []string{
			"aws_internet_gateway",
			"aws_nat_gateway",
			"aws_network_acl",
			"aws_security_group",
			"aws_subnet",
			"aws_vpn_gateway",
		},
		F: testSweepVPCs,
	})
}

func testSweepVPCs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	req := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-value"),
				Values: []*string{
					aws.String("terraform-testacc-*"),
				},
			},
		},
	}
	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing vpcs: %s", err)
	}

	if len(resp.Vpcs) == 0 {
		log.Print("[DEBUG] No aws vpcs to sweep")
		return nil
	}

	for _, vpc := range resp.Vpcs {
		// delete the vpc
		_, err := conn.DeleteVpc(&ec2.DeleteVpcInput{
			VpcId: vpc.VpcId,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting VPC (%s): %s",
				*vpc.VpcId, err)
		}
	}

	return nil
}

func TestAccAWSVpc_basic(t *testing.T) {
	var vpc ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(
						"aws_vpc.foo", "default_route_table_id"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "enable_dns_support", "true"),
				),
			},
		},
	})
}

func TestAccAWSVpc_enableIpv6(t *testing.T) {
	var vpc ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigIpv6Enabled,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(
						"aws_vpc.foo", "ipv6_association_id"),
					resource.TestCheckResourceAttrSet(
						"aws_vpc.foo", "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "assign_generated_ipv6_cidr_block", "true"),
				),
			},
			{
				Config: testAccVpcConfigIpv6Disabled,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "assign_generated_ipv6_cidr_block", "false"),
				),
			},
			{
				Config: testAccVpcConfigIpv6Enabled,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
					resource.TestCheckResourceAttrSet(
						"aws_vpc.foo", "ipv6_association_id"),
					resource.TestCheckResourceAttrSet(
						"aws_vpc.foo", "ipv6_cidr_block"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "assign_generated_ipv6_cidr_block", "true"),
				),
			},
		},
	})
}

func TestAccAWSVpc_dedicatedTenancy(t *testing.T) {
	var vpc ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcDedicatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.bar", &vpc),
					resource.TestCheckResourceAttr(
						"aws_vpc.bar", "instance_tenancy", "dedicated"),
				),
			},
		},
	})
}

func TestAccAWSVpc_tags(t *testing.T) {
	var vpc ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
					testAccCheckTags(&vpc.Tags, "foo", "bar"),
				),
			},

			{
				Config: testAccVpcConfigTagsUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckTags(&vpc.Tags, "foo", ""),
					testAccCheckTags(&vpc.Tags, "bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSVpc_update(t *testing.T) {
	var vpc ec2.Vpc

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
				),
			},
			{
				Config: testAccVpcConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "enable_dns_hostnames", "true"),
				),
			},
		},
	})
}

func testAccCheckVpcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc" {
			continue
		}

		// Try to find the VPC
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err == nil {
			if len(resp.Vpcs) > 0 {
				return fmt.Errorf("VPCs still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpcID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckVpcCidr(vpc *ec2.Vpc, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		CIDRBlock := vpc.CidrBlock
		if *CIDRBlock != expected {
			return fmt.Errorf("Bad cidr: %s", *vpc.CidrBlock)
		}

		return nil
	}
}

func testAccCheckVpcExists(n string, vpc *ec2.Vpc) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		DescribeVpcOpts := &ec2.DescribeVpcsInput{
			VpcIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.Vpcs) == 0 {
			return fmt.Errorf("VPC not found")
		}

		*vpc = *resp.Vpcs[0]

		return nil
	}
}

// https://github.com/hashicorp/terraform/issues/1301
func TestAccAWSVpc_bothDnsOptionsSet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_BothDnsOptions,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_vpc.bar", "enable_dns_hostnames", "true"),
					resource.TestCheckResourceAttr(
						"aws_vpc.bar", "enable_dns_support", "true"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform/issues/10168
func TestAccAWSVpc_DisabledDnsSupport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_DisabledDnsSupport,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_vpc.bar", "enable_dns_support", "false"),
				),
			},
		},
	})
}

func TestAccAWSVpc_classiclinkOptionSet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_ClassiclinkOption,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_vpc.bar", "enable_classiclink", "true"),
				),
			},
		},
	})
}

func TestAccAWSVpc_classiclinkDnsSupportOptionSet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcConfig_ClassiclinkDnsSupportOption,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_vpc.bar", "enable_classiclink_dns_support", "true"),
				),
			},
		},
	})
}

const testAccVpcConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "terraform-testacc-vpc"
	}
}
`

const testAccVpcConfigIpv6Enabled = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags {
		Name = "terraform-testacc-vpc-ipv6"
	}
}
`

const testAccVpcConfigIpv6Disabled = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags {
		Name = "terraform-testacc-vpc-ipv6"
	}
}
`

const testAccVpcConfigUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	enable_dns_hostnames = true
	tags {
		Name = "terraform-testacc-vpc"
	}
}
`

const testAccVpcConfigTags = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		foo = "bar"
		Name = "terraform-testacc-vpc-tags"
	}
}
`

const testAccVpcConfigTagsUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		bar = "baz"
		Name = "terraform-testacc-vpc-tags"
	}
}
`
const testAccVpcDedicatedConfig = `
resource "aws_vpc" "bar" {
	instance_tenancy = "dedicated"
	cidr_block = "10.2.0.0/16"
	tags {
		Name = "terraform-testacc-vpc-dedicated"
	}
}
`

const testAccVpcConfig_BothDnsOptions = `
provider "aws" {
	region = "eu-central-1"
}

resource "aws_vpc" "bar" {
	cidr_block = "10.2.0.0/16"
	enable_dns_hostnames = true
	enable_dns_support = true
	tags {
		Name = "terraform-testacc-vpc-both-dns-opts"
	}
}
`

const testAccVpcConfig_DisabledDnsSupport = `
resource "aws_vpc" "bar" {
	cidr_block = "10.2.0.0/16"
	enable_dns_support = false
	tags {
		Name = "terraform-testacc-vpc-disabled-dns-support"
	}
}
`

const testAccVpcConfig_ClassiclinkOption = `
resource "aws_vpc" "bar" {
	cidr_block = "172.2.0.0/16"
	enable_classiclink = true
	tags {
		Name = "terraform-testacc-vpc-classic-link"
	}
}
`

const testAccVpcConfig_ClassiclinkDnsSupportOption = `
resource "aws_vpc" "bar" {
	cidr_block = "172.2.0.0/16"
	enable_classiclink = true
	enable_classiclink_dns_support = true
	tags {
		Name = "terraform-testacc-vpc-classic-link-support"
	}
}
`
