package aws

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEIP_importEc2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resourceName := "aws_eip.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPInstanceEc2Classic,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEIP_importVpc(t *testing.T) {
	resourceName := "aws_eip.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPNetworkInterfaceConfig,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEIP_basic(t *testing.T) {
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_eip.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &conf),
					testAccCheckAWSEIPAttributes(&conf),
				),
			},
		},
	})
}

func TestAccAWSEIP_instance(t *testing.T) {
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_eip.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPInstanceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &conf),
					testAccCheckAWSEIPAttributes(&conf),
				),
			},

			{
				Config: testAccAWSEIPInstanceConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &conf),
					testAccCheckAWSEIPAttributes(&conf),
				),
			},
		},
	})
}

func TestAccAWSEIP_network_interface(t *testing.T) {
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_eip.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPNetworkInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPAssociated(&conf),
				),
			},
		},
	})
}

func TestAccAWSEIP_twoEIPsOneNetworkInterface(t *testing.T) {
	var one, two ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_eip.one",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPMultiNetworkInterfaceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.one", &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
					testAccCheckAWSEIPExists("aws_eip.two", &two),
					testAccCheckAWSEIPAttributes(&two),
					testAccCheckAWSEIPAssociated(&two),
				),
			},
		},
	})
}

// This test is an expansion of TestAccAWSEIP_instance, by testing the
// associated Private EIPs of two instances
func TestAccAWSEIP_associated_user_private_ip(t *testing.T) {
	var one ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_eip.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPInstanceConfig_associated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
				),
			},

			{
				Config: testAccAWSEIPInstanceConfig_associated_switch,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &one),
					testAccCheckAWSEIPAttributes(&one),
					testAccCheckAWSEIPAssociated(&one),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform/issues/3429 (now
// https://github.com/terraform-providers/terraform-provider-aws/issues/42)
func TestAccAWSEIP_classic_disassociate(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIP_classic_disassociate("instance-store"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"aws_eip.ip.0",
						"instance"),
					resource.TestCheckResourceAttrSet(
						"aws_eip.ip.1",
						"instance"),
				),
			},
			{
				Config: testAccAWSEIP_classic_disassociate("ebs"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"aws_eip.ip.0",
						"instance"),
					resource.TestCheckResourceAttrSet(
						"aws_eip.ip.1",
						"instance"),
				),
			},
		},
	})
}

func TestAccAWSEIP_disappears(t *testing.T) {
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &conf),
					testAccCheckAWSEIPDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEIPAssociate_not_associated(t *testing.T) {
	var conf ec2.Address

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_eip.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPAssociate_not_associated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &conf),
					testAccCheckAWSEIPAttributes(&conf),
				),
			},

			{
				Config: testAccAWSEIPAssociate_associated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.bar", &conf),
					testAccCheckAWSEIPAttributes(&conf),
					testAccCheckAWSEIPAssociated(&conf),
				),
			},
		},
	})
}

func TestAccAWSEIP_tags(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.bar"
	rName1 := fmt.Sprintf("%s-%d", t.Name(), acctest.RandInt())
	rName2 := fmt.Sprintf("%s-%d", t.Name(), acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_eip.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig_tags(rName1, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName1),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", t.Name()),
				),
			},
			{
				Config: testAccAWSEIPConfig_tags(rName2, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.RandomName", rName2),
					resource.TestCheckResourceAttr(resourceName, "tags.TestName", t.Name()),
				),
			},
		},
	})
}

func TestAccAWSEIP_PublicIpv4Pool_default(t *testing.T) {
	var conf ec2.Address
	resourceName := "aws_eip.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig_PublicIpv4Pool_default,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", "amazon"),
				),
			},
		},
	})
}

func TestAccAWSEIP_PublicIpv4Pool_custom(t *testing.T) {
	if os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL") == "" {
		t.Skip("Environment variable AWS_EC2_EIP_PUBLIC_IPV4_POOL is not set")
	}

	var conf ec2.Address
	resourceName := "aws_eip.bar"

	poolName := os.Getenv("AWS_EC2_EIP_PUBLIC_IPV4_POOL")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEIPConfig_PublicIpv4Pool_custom(poolName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists(resourceName, &conf),
					testAccCheckAWSEIPAttributes(&conf),
					resource.TestCheckResourceAttr(resourceName, "public_ipv4_pool", poolName),
				),
			},
		},
	})
}

func testAccCheckAWSEIPDisappears(v *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_eip" {
				continue
			}

			_, err := conn.ReleaseAddress(&ec2.ReleaseAddressInput{
				AllocationId: aws.String(rs.Primary.ID),
			})
			return err
		}
		return nil
	}
}

func testAccCheckAWSEIPDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_eip" {
			continue
		}

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			req := &ec2.DescribeAddressesInput{
				AllocationIds: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				// Verify the error is what we want
				if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidAllocationID.NotFound" || ae.Code() == "InvalidAddress.NotFound" {
					continue
				}
				return err
			}

			if len(describe.Addresses) > 0 {
				return fmt.Errorf("still exists")
			}
		} else {
			req := &ec2.DescribeAddressesInput{
				PublicIps: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				// Verify the error is what we want
				if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidAllocationID.NotFound" || ae.Code() == "InvalidAddress.NotFound" {
					continue
				}
				return err
			}

			if len(describe.Addresses) > 0 {
				return fmt.Errorf("still exists")
			}
		}
	}

	return nil
}

func testAccCheckAWSEIPAttributes(conf *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *conf.PublicIp == "" {
			return fmt.Errorf("empty public_ip")
		}

		return nil
	}
}

func testAccCheckAWSEIPAssociated(conf *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if conf.AssociationId == nil || *conf.AssociationId == "" {
			return fmt.Errorf("empty association_id")
		}

		return nil
	}
}

func testAccCheckAWSEIPExists(n string, res *ec2.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		if strings.Contains(rs.Primary.ID, "eipalloc") {
			req := &ec2.DescribeAddressesInput{
				AllocationIds: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				return err
			}

			if len(describe.Addresses) != 1 ||
				*describe.Addresses[0].AllocationId != rs.Primary.ID {
				return fmt.Errorf("EIP not found")
			}
			*res = *describe.Addresses[0]

		} else {
			req := &ec2.DescribeAddressesInput{
				PublicIps: []*string{aws.String(rs.Primary.ID)},
			}
			describe, err := conn.DescribeAddresses(req)
			if err != nil {
				return err
			}

			if len(describe.Addresses) != 1 ||
				*describe.Addresses[0].PublicIp != rs.Primary.ID {
				return fmt.Errorf("EIP not found")
			}
			*res = *describe.Addresses[0]
		}

		return nil
	}
}

const testAccAWSEIPConfig = `
resource "aws_eip" "bar" {
}
`

func testAccAWSEIPConfig_tags(rName, testName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "bar" {
  tags = {
    RandomName = "%[1]s"
    TestName   = "%[2]s"
  }
}
`, rName, testName)
}

const testAccAWSEIPConfig_PublicIpv4Pool_default = `
resource "aws_eip" "bar" {
   vpc = true
}
`

func testAccAWSEIPConfig_PublicIpv4Pool_custom(poolName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "bar" {
   vpc = true
   public_ipv4_pool = "%s"
}
`, poolName)
}

const testAccAWSEIPInstanceEc2Classic = `
provider "aws" {
	region = "us-east-1"
}

data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "foo" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
	tags = {
		Name = "testAccAWSEIPInstanceEc2Classic"
	}
}

resource "aws_eip" "bar" {
	instance = "${aws_instance.foo.id}"
}
`

const testAccAWSEIPInstanceConfig = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "foo" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "bar" {
	instance = "${aws_instance.foo.id}"
}
`

const testAccAWSEIPInstanceConfig2 = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "bar" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "bar" {
	instance = "${aws_instance.bar.id}"
}
`

const testAccAWSEIPInstanceConfig_associated = `
data "aws_ami" "amzn-ami-minimal-hvm" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-eip-instance-associated"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.default.id}"

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "tf_test_subnet" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = ["aws_internet_gateway.gw"]

  tags = {
    Name = "tf-acc-eip-instance-associated"
  }
}

resource "aws_instance" "foo" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm.id}"
  instance_type = "t2.micro"

  private_ip = "10.0.0.12"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "foo instance"
  }
}

resource "aws_instance" "bar" {
  ami = "${data.aws_ami.amzn-ami-minimal-hvm.id}"

  instance_type = "t2.micro"

  private_ip = "10.0.0.19"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "bar instance"
  }
}

resource "aws_eip" "bar" {
  vpc = true

  instance                  = "${aws_instance.bar.id}"
  associate_with_private_ip = "10.0.0.19"
}
`
const testAccAWSEIPInstanceConfig_associated_switch = `
data "aws_ami" "amzn-ami-minimal-hvm" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-eip-instance-associated"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.default.id}"

  tags = {
    Name = "main"
  }
}

resource "aws_subnet" "tf_test_subnet" {
  vpc_id                  = "${aws_vpc.default.id}"
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = ["aws_internet_gateway.gw"]

  tags = {
    Name = "tf-acc-eip-instance-associated"
  }
}

resource "aws_instance" "foo" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm.id}"
  instance_type = "t2.micro"

  private_ip = "10.0.0.12"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "foo instance"
  }
}

resource "aws_instance" "bar" {
  ami = "${data.aws_ami.amzn-ami-minimal-hvm.id}"

  instance_type = "t2.micro"

  private_ip = "10.0.0.19"
  subnet_id  = "${aws_subnet.tf_test_subnet.id}"

  tags = {
    Name = "bar instance"
  }
}

resource "aws_eip" "bar" {
  vpc = true

  instance                  = "${aws_instance.foo.id}"
  associate_with_private_ip = "10.0.0.12"
}
`

const testAccAWSEIPNetworkInterfaceConfig = `
resource "aws_vpc" "bar" {
	cidr_block = "10.0.0.0/24"
	tags = {
		Name = "terraform-testacc-eip-network-interface"
	}
}

resource "aws_internet_gateway" "bar" {
	vpc_id = "${aws_vpc.bar.id}"
}

resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.bar.id}"
  availability_zone = "us-west-2a"
  cidr_block = "10.0.0.0/24"
  tags = {
	Name = "tf-acc-eip-network-interface"
  }
}

resource "aws_network_interface" "bar" {
  subnet_id = "${aws_subnet.bar.id}"
	private_ips = ["10.0.0.10"]
  security_groups = [ "${aws_vpc.bar.default_security_group_id}" ]
}

resource "aws_eip" "bar" {
	vpc = "true"
	network_interface = "${aws_network_interface.bar.id}"
	depends_on = ["aws_internet_gateway.bar"]
}
`

const testAccAWSEIPMultiNetworkInterfaceConfig = `
resource "aws_vpc" "bar" {
  cidr_block = "10.0.0.0/24"
	tags = {
		Name = "terraform-testacc-eip-multi-network-interface"
	}
}

resource "aws_internet_gateway" "bar" {
  vpc_id = "${aws_vpc.bar.id}"
}

resource "aws_subnet" "bar" {
  vpc_id            = "${aws_vpc.bar.id}"
  availability_zone = "us-west-2a"
  cidr_block        = "10.0.0.0/24"
  tags = {
	Name = "tf-acc-eip-multi-network-interface"
  }
}

resource "aws_network_interface" "bar" {
  subnet_id       = "${aws_subnet.bar.id}"
  private_ips     = ["10.0.0.10", "10.0.0.11"]
  security_groups = ["${aws_vpc.bar.default_security_group_id}"]
}

resource "aws_eip" "one" {
  vpc                       = "true"
  network_interface         = "${aws_network_interface.bar.id}"
  associate_with_private_ip = "10.0.0.10"
  depends_on                = ["aws_internet_gateway.bar"]
}

resource "aws_eip" "two" {
  vpc                       = "true"
  network_interface         = "${aws_network_interface.bar.id}"
  associate_with_private_ip = "10.0.0.11"
  depends_on                = ["aws_internet_gateway.bar"]
}
`

func testAccAWSEIP_classic_disassociate(rootDeviceType string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

variable "server_count" {
  default = 2
}

data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = [%q]
  }
}

resource "aws_eip" "ip" {
  count    = "${var.server_count}"
  instance = "${element(aws_instance.example.*.id, count.index)}"
  vpc      = true
}

resource "aws_instance" "example" {
  count = "${var.server_count}"

  ami                         = "${data.aws_ami.amzn-ami-minimal-pv.id}"
  instance_type               = "m1.small"
  associate_public_ip_address = true
  subnet_id                   = "${aws_subnet.us-east-1b-public.id}"
  availability_zone           = "${aws_subnet.us-east-1b-public.availability_zone}"

  tags = {
    Name = "testAccAWSEIP_classic_disassociate"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-eip-classic-disassociate"
	}
}

resource "aws_internet_gateway" "example" {
  vpc_id = "${aws_vpc.example.id}"
}

resource "aws_subnet" "us-east-1b-public" {
  vpc_id = "${aws_vpc.example.id}"

  cidr_block        = "10.0.0.0/24"
  availability_zone = "us-east-1b"
  tags = {
    Name = "tf-acc-eip-classic-disassociate"
  }
}

resource "aws_route_table" "us-east-1-public" {
  vpc_id = "${aws_vpc.example.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.example.id}"
  }
}

resource "aws_route_table_association" "us-east-1b-public" {
  subnet_id      = "${aws_subnet.us-east-1b-public.id}"
  route_table_id = "${aws_route_table.us-east-1-public.id}"
}`, rootDeviceType)
}

const testAccAWSEIPAssociate_not_associated = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "foo" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "bar" {
}
`

const testAccAWSEIPAssociate_associated = `
data "aws_ami" "amzn-ami-minimal-pv" {
  most_recent = true
  filter {
    name = "owner-alias"
    values = ["amazon"]
  }
  filter {
    name = "name"
    values = ["amzn-ami-minimal-pv-*"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}

resource "aws_instance" "foo" {
	ami = "${data.aws_ami.amzn-ami-minimal-pv.id}"
	instance_type = "m1.small"
}

resource "aws_eip" "bar" {
	instance = "${aws_instance.foo.id}"
}
`
