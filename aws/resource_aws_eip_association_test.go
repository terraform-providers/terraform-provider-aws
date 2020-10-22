package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEIPAssociation_instance(t *testing.T) {
	resourceName := "aws_eip_association.test"
	var a ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfig_instance(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &a),
					testAccCheckAWSEIPAssociationExists(resourceName, &a),
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

func TestAccAWSEIPAssociation_networkInterface(t *testing.T) {
	resourceName := "aws_eip_association.test"
	var a ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfig_networkInterface,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &a),
					testAccCheckAWSEIPAssociationExists(resourceName, &a),
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

func TestAccAWSEIPAssociation_basic(t *testing.T) {
	var a ec2.Address
	resourceName := "aws_eip_association.by_allocation_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2VPCOnlyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test.0", &a),
					testAccCheckAWSEIPAssociationExists("aws_eip_association.by_allocation_id", &a),
					testAccCheckAWSEIPExists("aws_eip.test.1", &a),
					testAccCheckAWSEIPAssociationExists("aws_eip_association.by_public_ip", &a),
					testAccCheckAWSEIPExists("aws_eip.test.2", &a),
					testAccCheckAWSEIPAssociationExists("aws_eip_association.to_eni", &a),
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

func TestAccAWSEIPAssociation_ec2Classic(t *testing.T) {
	var a ec2.Address
	resourceName := "aws_eip_association.test"

	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	// This test cannot run in parallel with the other EIP Association tests
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfig_ec2Classic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &a),
					testAccCheckAWSEIPAssociationExists(resourceName, &a),
					resource.TestCheckResourceAttrSet(resourceName, "public_ip"),
					resource.TestCheckResourceAttr(resourceName, "allocation_id", ""),
					testAccCheckAWSEIPAssociationHasIpBasedId(resourceName),
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

func TestAccAWSEIPAssociation_spotInstance(t *testing.T) {
	var a ec2.Address
	rInt := acctest.RandInt()
	resourceName := "aws_eip_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfig_spotInstance(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &a),
					testAccCheckAWSEIPAssociationExists(resourceName, &a),
					resource.TestCheckResourceAttrSet(resourceName, "allocation_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
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

func TestAccAWSEIPAssociation_disappears(t *testing.T) {
	var a ec2.Address

	resourceName := "aws_eip_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfigDisappears(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &a),
					testAccCheckAWSEIPAssociationExists(resourceName, &a),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEipAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEIPAssociationExists(name string, res *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP Association ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		platforms := testAccProvider.Meta().(*AWSClient).supportedplatforms

		request, err := describeAddressesById(rs.Primary.ID, platforms)
		if err != nil {
			return err
		}

		describe, err := conn.DescribeAddresses(request)
		if err != nil {
			return err
		}

		if len(describe.Addresses) != 1 ||
			(!hasEc2Classic(platforms) && *describe.Addresses[0].AssociationId != *res.AssociationId) {
			return fmt.Errorf("EIP Association not found")
		}

		return nil
	}
}

func testAccCheckAWSEIPAssociationHasIpBasedId(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP Association ID is set")
		}

		if rs.Primary.ID != rs.Primary.Attributes["public_ip"] {
			return fmt.Errorf("Expected EIP Association ID to be equal to Public IP (%q), given: %q",
				rs.Primary.Attributes["public_ip"], rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckAWSEIPAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eip_association" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP Association ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		request := &ec2.DescribeAddressesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("association-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		}
		describe, err := conn.DescribeAddresses(request)
		if err != nil {
			return err
		}

		if len(describe.Addresses) > 0 {
			return fmt.Errorf("EIP Association still exists")
		}
	}
	return nil
}

func testAccAWSEIPAssociationConfig() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/24"
  tags = {
    Name = "terraform-testacc-eip-association"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/25"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-eip-association"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  count             = 2
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "t2.small"
  subnet_id         = aws_subnet.test.id
  private_ip        = "192.168.0.${count.index + 10}"
}

resource "aws_eip" "test" {
  count = 3
  vpc   = true
}

resource "aws_eip_association" "by_allocation_id" {
  allocation_id = aws_eip.test[0].id
  instance_id   = aws_instance.test[0].id
  depends_on    = [aws_instance.test]
}

resource "aws_eip_association" "by_public_ip" {
  public_ip   = aws_eip.test[1].public_ip
  instance_id = aws_instance.test[1].id
  depends_on  = [aws_instance.test]
}

resource "aws_eip_association" "to_eni" {
  allocation_id        = aws_eip.test[2].id
  network_interface_id = aws_network_interface.test.id
}

resource "aws_network_interface" "test" {
  subnet_id   = aws_subnet.test.id
  private_ips = ["192.168.0.50"]
  depends_on  = [aws_instance.test]

  attachment {
    instance     = aws_instance.test[0].id
    device_index = 1
  }
}
`)
}

func testAccAWSEIPAssociationConfigDisappears() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "192.168.0.0/24"
  tags = {
    Name = "terraform-testacc-eip-association-disappears"
  }
}

resource "aws_subnet" "sub" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "192.168.0.0/25"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-eip-association-disappears"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.main.id
}

resource "aws_instance" "test" {
  ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_type     = "t2.small"
  subnet_id         = aws_subnet.sub.id
}

resource "aws_eip" "test" {
  vpc = true
}

resource "aws_eip_association" "test" {
  allocation_id = aws_eip.test.id
  instance_id   = aws_instance.test.id
}
`)
}

const testAccAWSEIPAssociationConfig_ec2Classic = `
provider "aws" {
  region = "us-east-1"
}

resource "aws_eip" "test" {}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/ebs/ubuntu-trusty-14.04-i386-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["paravirtual"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "test" {
  ami               = data.aws_ami.ubuntu.id
  availability_zone = data.aws_availability_zones.available.names[0]

  # tflint-ignore: aws_instance_previous_type
  instance_type = "t1.micro"
}

resource "aws_eip_association" "test" {
  public_ip   = aws_eip.test.public_ip
  instance_id = aws_instance.test.id
}
`

func testAccAWSEIPAssociationConfig_spotInstance(rInt int) string {
	return composeConfig(
		testAccAWSSpotInstanceRequestConfig(rInt), `
resource "aws_eip" "test" {
  vpc = true
}

resource "aws_eip_association" "test" {
  allocation_id = aws_eip.test.id
  instance_id   = aws_spot_instance_request.test.spot_instance_id
}
`)
}

func testAccAWSEIPAssociationConfig_instance() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.small"
}

resource "aws_eip" "test" {}

resource "aws_eip_association" "test" {
  allocation_id = aws_eip.test.id
  instance_id   = aws_instance.test.id
}
`))
}

const testAccAWSEIPAssociationConfig_networkInterface = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.1.0/24"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_eip" "test" {}

resource "aws_eip_association" "test" {
  allocation_id        = aws_eip.test.id
  network_interface_id = aws_network_interface.test.id
}
`
