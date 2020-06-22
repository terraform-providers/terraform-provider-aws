package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
)

func TestAccAWSRoute53ZoneAssociation_basic(t *testing.T) {
	var vpc route53.VPC
	resourceName := "aws_route53_zone_association.foobar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName, &vpc),
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

func TestAccAWSRoute53ZoneAssociation_disappears(t *testing.T) {
	var vpc route53.VPC
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone_association.foobar"
	route53ZoneResourceName := "aws_route53_zone.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(route53ZoneResourceName, &zone),
					testAccCheckRoute53ZoneAssociationExists(resourceName, &vpc),
					testAccCheckRoute53ZoneAssociationDisappears(&zone, &vpc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_disappears_VPC(t *testing.T) {
	var ec2Vpc ec2.Vpc
	var route53Vpc route53.VPC
	resourceName := "aws_route53_zone_association.foobar"
	vpcResourceName := "aws_vpc.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName, &route53Vpc),
					testAccCheckVpcExists(vpcResourceName, &ec2Vpc),
					testAccCheckVpcDisappears(&ec2Vpc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_disappears_Zone(t *testing.T) {
	var vpc route53.VPC
	var zone route53.GetHostedZoneOutput
	resourceName := "aws_route53_zone_association.foobar"
	route53ZoneResourceName := "aws_route53_zone.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneExists(route53ZoneResourceName, &zone),
					testAccCheckRoute53ZoneAssociationExists(resourceName, &vpc),
					testAccCheckRoute53ZoneDisappears(&zone),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ZoneAssociation_region(t *testing.T) {
	var vpc route53.VPC
	resourceName := "aws_route53_zone_association.test"

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckRoute53ZoneAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ZoneAssociationRegionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ZoneAssociationExists(resourceName, &vpc),
				),
			},
			{
				Config:            testAccRoute53ZoneAssociationRegionConfig(),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRoute53ZoneAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_zone_association" {
			continue
		}

		zoneID, vpcID, err := resourceAwsRoute53ZoneAssociationParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		vpc, err := route53GetZoneAssociation(conn, zoneID, vpcID)

		if isAWSErr(err, route53.ErrCodeNoSuchHostedZone, "") {
			continue
		}

		if err != nil {
			return err
		}

		if vpc != nil {
			return fmt.Errorf("Route 53 Hosted Zone (%s) Association (%s) still exists", zoneID, vpcID)
		}
	}
	return nil
}

func testAccCheckRoute53ZoneAssociationExists(n string, vpc *route53.VPC) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No zone association ID is set")
		}

		zoneID, vpcID, err := resourceAwsRoute53ZoneAssociationParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).r53conn

		associationVPC, err := route53GetZoneAssociation(conn, zoneID, vpcID)

		if err != nil {
			return err
		}

		if associationVPC == nil {
			return fmt.Errorf("Route 53 Hosted Zone (%s) Association (%s) not found", zoneID, vpcID)
		}

		*vpc = *associationVPC

		return nil
	}
}

func testAccCheckRoute53ZoneAssociationDisappears(zone *route53.GetHostedZoneOutput, vpc *route53.VPC) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53conn

		input := &route53.DisassociateVPCFromHostedZoneInput{
			HostedZoneId: zone.HostedZone.Id,
			VPC:          vpc,
			Comment:      aws.String("Managed by Terraform"),
		}

		_, err := conn.DisassociateVPCFromHostedZone(input)

		return err
	}
}

const testAccRoute53ZoneAssociationConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.6.0.0/16"
	enable_dns_hostnames = true
	enable_dns_support = true
	tags = {
		Name = "terraform-testacc-route53-zone-association-foo"
	}
}

resource "aws_vpc" "bar" {
	cidr_block = "10.7.0.0/16"
	enable_dns_hostnames = true
	enable_dns_support = true
	tags = {
		Name = "terraform-testacc-route53-zone-association-bar"
	}
}

resource "aws_route53_zone" "foo" {
	name = "foo.com"
	vpc {
		vpc_id = "${aws_vpc.foo.id}"
	}
	lifecycle {
		ignore_changes = ["vpc"]
	}
}

resource "aws_route53_zone_association" "foobar" {
	zone_id = "${aws_route53_zone.foo.id}"
	vpc_id  = "${aws_vpc.bar.id}"
}
`

func testAccRoute53ZoneAssociationRegionConfig() string {
	return testAccAlternateRegionProviderConfig() + `
data "aws_region" "alternate" {
  provider = "aws.alternate"
}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block           = "10.6.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "terraform-testacc-route53-zone-association-region-foo"
  }
}

resource "aws_vpc" "alternate" {
  provider = "aws.alternate"

  cidr_block           = "10.7.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "terraform-testacc-route53-zone-association-region-bar"
  }
}

resource "aws_route53_zone" "test" {
  name = "foo.com"

  vpc {
    vpc_id     = aws_vpc.test.id
    vpc_region = data.aws_region.current.name
  }

  lifecycle {
    ignore_changes = [vpc]
  }
}

resource "aws_route53_zone_association" "test" {
  vpc_id     = aws_vpc.alternate.id
  vpc_region = data.aws_region.alternate.name
  zone_id    = aws_route53_zone.test.id
}
`
}
