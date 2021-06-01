package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSNetworkInterfaceAttachment_basic(t *testing.T) {
	var conf ec2.NetworkInterface
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_network_interface.bar",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSENIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNetworkInterfaceAttachmentConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSENIExists("aws_network_interface.bar", &conf),
					resource.TestCheckResourceAttr(
						"aws_network_interface_attachment.test", "device_index", "1"),
					resource.TestCheckResourceAttrSet(
						"aws_network_interface_attachment.test", "instance_id"),
					resource.TestCheckResourceAttrSet(
						"aws_network_interface_attachment.test", "network_interface_id"),
					resource.TestCheckResourceAttrSet(
						"aws_network_interface_attachment.test", "attachment_id"),
					resource.TestCheckResourceAttrSet(
						"aws_network_interface_attachment.test", "status"),
				),
			},
		},
	})
}

func testAccAWSNetworkInterfaceAttachmentConfig_basic(rInt int) string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-network-iface-attachment-basic"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = "172.16.10.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-network-iface-attachment-basic"
  }
}

resource "aws_security_group" "foo" {
  vpc_id      = aws_vpc.foo.id
  description = "foo"
  name        = "foo-%d"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }
}

resource "aws_network_interface" "bar" {
  subnet_id       = aws_subnet.foo.id
  private_ips     = ["172.16.10.100"]
  security_groups = [aws_security_group.foo.id]
  description     = "Managed by Terraform"

  tags = {
    Name = "bar_interface"
  }
}

resource "aws_instance" "foo" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t2.micro"
  subnet_id     = aws_subnet.foo.id

  tags = {
    Name = "foo-%d"
  }
}

resource "aws_network_interface_attachment" "test" {
  device_index         = 1
  instance_id          = aws_instance.foo.id
  network_interface_id = aws_network_interface.bar.id
}
`, rInt, rInt)
}
