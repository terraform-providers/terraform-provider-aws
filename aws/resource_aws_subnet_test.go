package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// add sweeper to delete known test subnets
func init() {
	resource.AddTestSweepers("aws_subnet", &resource.Sweeper{
		Name: "aws_subnet",
		F:    testSweepSubnets,
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_batch_compute_environment",
			"aws_beanstalk_environment",
			"aws_db_instance",
			"aws_directory_service_directory",
			"aws_ec2_transit_gateway_vpc_attachment",
			"aws_eks_cluster",
			"aws_elasticache_cluster",
			"aws_elasticache_replication_group",
			"aws_elasticsearch_domain",
			"aws_elb",
			"aws_emr_cluster",
			"aws_lambda_function",
			"aws_lb",
			"aws_mq_broker",
			"aws_network_interface",
			"aws_redshift_cluster",
			"aws_spot_fleet_request",
			"aws_vpc_endpoint",
		},
	})
}

func testSweepSubnets(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	req := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag-value"),
				Values: []*string{
					aws.String("terraform-testacc-*"),
					aws.String("tf-acc-*"),
				},
			},
		},
	}
	resp, err := conn.DescribeSubnets(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Subnet sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing subnets: %s", err)
	}

	if len(resp.Subnets) == 0 {
		log.Print("[DEBUG] No aws subnets to sweep")
		return nil
	}

	for _, subnet := range resp.Subnets {
		if subnet == nil {
			continue
		}

		if aws.BoolValue(subnet.DefaultForAz) {
			continue
		}

		// delete the subnet
		_, err := conn.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: subnet.SubnetId,
		})
		if err != nil {
			return fmt.Errorf(
				"Error deleting Subnet (%s): %s",
				*subnet.SubnetId, err)
		}
	}

	return nil
}

func TestAccAWSSubnet_importBasic(t *testing.T) {
	resourceName := "aws_subnet.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSubnet_basic(t *testing.T) {
	var v ec2.Subnet

	testCheck := func(*terraform.State) error {
		if aws.StringValue(v.CidrBlock) != "10.1.1.0/24" {
			return fmt.Errorf("bad cidr: %s", aws.StringValue(v.CidrBlock))
		}

		if !aws.BoolValue(v.MapPublicIpOnLaunch) {
			return fmt.Errorf("bad MapPublicIpOnLaunch: %t", aws.BoolValue(v.MapPublicIpOnLaunch))
		}

		return nil
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_subnet.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &v),
					testCheck,
					// ipv6 should be empty if disabled so we can still use the property in conditionals
					resource.TestCheckResourceAttr(
						"aws_subnet.foo", "ipv6_cidr_block", ""),
					resource.TestMatchResourceAttr(
						"aws_subnet.foo",
						"arn",
						regexp.MustCompile(`^arn:[^:]+:ec2:[^:]+:\d{12}:subnet/subnet-.+`)),
					testAccCheckResourceAttrAccountID("aws_subnet.foo", "owner_id"),
					resource.TestCheckResourceAttrSet(
						"aws_subnet.foo", "availability_zone"),
					resource.TestCheckResourceAttrSet(
						"aws_subnet.foo", "availability_zone_id"),
				),
			},
		},
	})
}

func TestAccAWSSubnet_ipv6(t *testing.T) {
	var before, after ec2.Subnet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_subnet.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &before),
					testAccCheckAwsSubnetIpv6BeforeUpdate(t, &before),
				),
			},
			{
				Config: testAccSubnetConfigIpv6UpdateAssignIpv6OnCreation,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &after),
					testAccCheckAwsSubnetIpv6AfterUpdate(t, &after),
				),
			},
			{
				Config: testAccSubnetConfigIpv6UpdateIpv6Cidr,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &after),

					testAccCheckAwsSubnetNotRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSubnet_enableIpv6(t *testing.T) {
	var subnet ec2.Subnet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_subnet.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigPreIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &subnet),
				),
			},
			{
				Config: testAccSubnetConfigIpv6,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &subnet),
				),
			},
		},
	})
}

func TestAccAWSSubnet_availabilityZoneId(t *testing.T) {
	var v ec2.Subnet

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_subnet.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSubnetConfigAvailabilityZoneId,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &v),
					resource.TestCheckResourceAttrSet(
						"aws_subnet.foo", "availability_zone"),
					resource.TestCheckResourceAttr(
						"aws_subnet.foo", "availability_zone_id", "usw2-az3"),
				),
			},
		},
	})
}

func testAccCheckAwsSubnetIpv6BeforeUpdate(t *testing.T, subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if subnet.Ipv6CidrBlockAssociationSet == nil {
			return fmt.Errorf("Expected IPV6 CIDR Block Association")
		}

		if !aws.BoolValue(subnet.AssignIpv6AddressOnCreation) {
			return fmt.Errorf("bad AssignIpv6AddressOnCreation: %t", aws.BoolValue(subnet.AssignIpv6AddressOnCreation))
		}

		return nil
	}
}

func testAccCheckAwsSubnetIpv6AfterUpdate(t *testing.T, subnet *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.BoolValue(subnet.AssignIpv6AddressOnCreation) {
			return fmt.Errorf("bad AssignIpv6AddressOnCreation: %t", aws.BoolValue(subnet.AssignIpv6AddressOnCreation))
		}

		return nil
	}
}

func testAccCheckAwsSubnetNotRecreated(t *testing.T,
	before, after *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.SubnetId != *after.SubnetId {
			t.Fatalf("Expected SubnetIDs not to change, but both got before: %s and after: %s", *before.SubnetId, *after.SubnetId)
		}
		return nil
	}
}

func testAccCheckSubnetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_subnet" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.Subnets) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidSubnetID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckSubnetExists(n string, v *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.Subnets) == 0 {
			return fmt.Errorf("Subnet not found")
		}

		*v = *resp.Subnets[0]

		return nil
	}
}

const testAccSubnetConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-subnet"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	map_public_ip_on_launch = true
	tags = {
		Name = "tf-acc-subnet"
	}
}
`

const testAccSubnetConfigPreIpv6 = `
resource "aws_vpc" "foo" {
	cidr_block = "10.10.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-subnet-ipv6"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.10.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	map_public_ip_on_launch = true
	tags = {
		Name = "tf-acc-subnet-ipv6"
	}
}
`

const testAccSubnetConfigIpv6 = `
resource "aws_vpc" "foo" {
	cidr_block = "10.10.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-subnet-ipv6"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.10.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.foo.ipv6_cidr_block, 8, 1)}"
	map_public_ip_on_launch = true
	assign_ipv6_address_on_creation = true
	tags = {
		Name = "tf-acc-subnet-ipv6"
	}
}
`

const testAccSubnetConfigIpv6UpdateAssignIpv6OnCreation = `
resource "aws_vpc" "foo" {
	cidr_block = "10.10.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-subnet-assign-ipv6-on-creation"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.10.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.foo.ipv6_cidr_block, 8, 1)}"
	map_public_ip_on_launch = true
	assign_ipv6_address_on_creation = false
	tags = {
		Name = "tf-acc-subnet-assign-ipv6-on-creation"
	}
}
`

const testAccSubnetConfigIpv6UpdateIpv6Cidr = `
resource "aws_vpc" "foo" {
	cidr_block = "10.10.0.0/16"
	assign_generated_ipv6_cidr_block = true
	tags = {
		Name = "terraform-testacc-subnet-ipv6-update-cidr"
	}
}

resource "aws_subnet" "foo" {
	cidr_block = "10.10.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	ipv6_cidr_block = "${cidrsubnet(aws_vpc.foo.ipv6_cidr_block, 8, 3)}"
	map_public_ip_on_launch = true
	assign_ipv6_address_on_creation = false
	tags = {
		Name = "tf-acc-subnet-ipv6-update-cidr"
	}
}
`

const testAccSubnetConfigAvailabilityZoneId = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-subnet"
  }
}

resource "aws_subnet" "foo" {
  cidr_block = "10.1.1.0/24"
  vpc_id = "${aws_vpc.foo.id}"
  availability_zone_id = "usw2-az3"
  tags = {
    Name = "tf-acc-subnet"
  }
}
`
