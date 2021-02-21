package aws

import (
	"fmt"
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPAssociationDestroy,
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPAssociationDestroy,
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
		PreCheck:          func() { testAccPreCheck(t); testAccEC2VPCOnlyPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPAssociationDestroy,
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfig_ec2Classic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPEc2ClassicExists("aws_eip.test", &a),
					testAccCheckAWSEIPAssociationEc2ClassicExists(resourceName, &a),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_eip_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociationConfig_spotInstance(rName),
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSEIPAssociationDestroy,
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

func testAccCheckAWSEIPAssociationEc2ClassicExists(name string, res *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP Association ID is set")
		}

		conn := testAccProviderEc2Classic.Meta().(*AWSClient).ec2conn
		platforms := testAccProviderEc2Classic.Meta().(*AWSClient).supportedplatforms

		request, err := describeAddressesById(rs.Primary.ID, platforms)

		if err != nil {
			return err
		}

		describe, err := conn.DescribeAddresses(request)

		if err != nil {
			return fmt.Errorf("error describing EC2 Address (%s): %w", rs.Primary.ID, err)
		}

		if len(describe.Addresses) != 1 || aws.StringValue(describe.Addresses[0].PublicIp) != rs.Primary.ID {
			return fmt.Errorf("EC2 Address (%s) not found", rs.Primary.ID)
		}

		*res = *describe.Addresses[0]

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
		testAccAvailableAZsNoOptInConfig(),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
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
		testAccAvailableAZsNoOptInConfig(),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(), `
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

func testAccAWSEIPAssociationConfig_ec2Classic() string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		testAccLatestAmazonLinuxPvEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t1.micro", "m3.medium", "m3.large", "c3.large", "r3.large"),
		`
resource "aws_eip" "test" {}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-pv-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_eip_association" "test" {
  public_ip   = aws_eip.test.public_ip
  instance_id = aws_instance.test.id
}
`)
}

func testAccAWSEIPAssociationConfig_spotInstance(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  key_name             = aws_key_pair.test.key_name
  spot_price           = "0.10"
  wait_for_fulfillment = true
  subnet_id            = aws_subnet.test.id
}

resource "aws_eip" "test" {
  vpc = true
}

resource "aws_eip_association" "test" {
  allocation_id = aws_eip.test.id
  instance_id   = aws_spot_instance_request.test.spot_instance_id
}
`, rName))
}

func testAccAWSEIPAssociationConfig_instance() string {
	return composeConfig(
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		testAccAvailableAZsNoOptInConfig(),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id
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

resource "aws_eip" "test" {
  vpc = true
}

resource "aws_eip_association" "test" {
  allocation_id        = aws_eip.test.id
  network_interface_id = aws_network_interface.test.id
}
`
