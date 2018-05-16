package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDefaultSubnet_basic(t *testing.T) {
	var v ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSubnetConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_default_subnet.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.Name", "terraform-testacc-default-subnet"),
				),
			},
		},
	})
}

func TestAccAWSDefaultSubnet_publicIp(t *testing.T) {
	var v ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultSubnetConfigPublicIp,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_default_subnet.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.Name", "terraform-testacc-default-subnet"),
				),
			},
			{
				Config: testAccAWSDefaultSubnetConfigNoPublicIp,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_default_subnet.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "map_public_ip_on_launch", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.Name", "terraform-testacc-default-subnet"),
				),
			},
		},
	})
}

func testAccCheckAWSDefaultSubnetDestroy(s *terraform.State) error {
	// We expect subnet to still exist
	return nil
}

const testAccAWSDefaultSubnetConfigBasic = `
resource "aws_default_subnet" "foo" {
  availability_zone = "us-west-2a"
  tags {
    Name = "terraform-testacc-default-subnet"
  }
}
`

const testAccAWSDefaultSubnetConfigPublicIp = `
resource "aws_default_subnet" "foo" {
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = true
  tags {
    Name = "terraform-testacc-default-subnet"
  }
}
`

const testAccAWSDefaultSubnetConfigNoPublicIp = `
resource "aws_default_subnet" "foo" {
  availability_zone = "us-west-2a"
  map_public_ip_on_launch = false
  tags {
    Name = "terraform-testacc-default-subnet"
  }
}
`
