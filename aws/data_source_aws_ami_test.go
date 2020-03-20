package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSAmiDataSource_natInstance(t *testing.T) {
	resourceName := "data.aws_ami.nat_ami"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAmiDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID(resourceName),
					// Check attributes. Some attributes are tough to test - any not contained here should not be considered
					// stable and should not be used in interpolation. Exception to block_device_mappings which should both
					// show up consistently and break if certain references are not available. However modification of the
					// snapshot ID which is bound to happen on the NAT AMIs will cause testing to break consistently, so
					// deep inspection is not included, simply the count is checked.
					// Tags and product codes may need more testing, but I'm having a hard time finding images with
					// these attributes set.
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestMatchResourceAttr(resourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr(resourceName, "description", regexp.MustCompile("^Amazon Linux AMI")),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					resource.TestMatchResourceAttr(resourceName, "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr(resourceName, "image_location", regexp.MustCompile("^amazon/")),
					resource.TestCheckResourceAttr(resourceName, "image_owner_alias", "amazon"),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "most_recent", "true"),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^amzn-ami-vpc-nat")),
					resource.TestCheckResourceAttr(resourceName, "owner_id", "137112412989"),
					resource.TestCheckResourceAttr(resourceName, "public", "true"),
					resource.TestCheckResourceAttr(resourceName, "product_codes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "root_device_type", "ebs"),
					resource.TestMatchResourceAttr(resourceName, "root_snapshot_id", regexp.MustCompile("^snap-")),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
					resource.TestCheckResourceAttr(resourceName, "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr(resourceName, "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}
func TestAccAWSAmiDataSource_windowsInstance(t *testing.T) {
	resourceName := "data.aws_ami.windows_ami"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAmiDataSourceWindowsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "27"),
					resource.TestMatchResourceAttr(resourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestMatchResourceAttr(resourceName, "description", regexp.MustCompile("^Microsoft Windows Server")),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					resource.TestMatchResourceAttr(resourceName, "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr(resourceName, "image_location", regexp.MustCompile("^amazon/")),
					resource.TestCheckResourceAttr(resourceName, "image_owner_alias", "amazon"),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "most_recent", "true"),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^Windows_Server-2012-R2")),
					resource.TestCheckResourceAttr(resourceName, "owner_id", "801119661308"),
					resource.TestCheckResourceAttr(resourceName, "platform", "windows"),
					resource.TestCheckResourceAttr(resourceName, "public", "true"),
					resource.TestCheckResourceAttr(resourceName, "product_codes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "root_device_name", "/dev/sda1"),
					resource.TestCheckResourceAttr(resourceName, "root_device_type", "ebs"),
					resource.TestMatchResourceAttr(resourceName, "root_snapshot_id", regexp.MustCompile("^snap-")),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
					resource.TestCheckResourceAttr(resourceName, "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr(resourceName, "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccAWSAmiDataSource_instanceStore(t *testing.T) {
	resourceName := "data.aws_ami.instance_store_ami"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAmiDataSourceInstanceStoreConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					resource.TestMatchResourceAttr(resourceName, "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr(resourceName, "image_location", regexp.MustCompile("ubuntu-trusty-14.04-amd64-server")),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "most_recent", "true"),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^ubuntu/images/hvm-instance/ubuntu-trusty-14.04-amd64-server")),
					resource.TestCheckResourceAttr(resourceName, "owner_id", "099720109477"),
					resource.TestCheckResourceAttr(resourceName, "public", "true"),
					resource.TestCheckResourceAttr(resourceName, "product_codes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "root_device_type", "instance-store"),
					resource.TestCheckResourceAttr(resourceName, "root_snapshot_id", ""),
					resource.TestCheckResourceAttr(resourceName, "sriov_net_support", "simple"),
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
					resource.TestCheckResourceAttr(resourceName, "state_reason.code", "UNSET"),
					resource.TestCheckResourceAttr(resourceName, "state_reason.message", "UNSET"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "virtualization_type", "hvm"),
				),
			},
		},
	})
}

func TestAccAWSAmiDataSource_localNameFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAmiDataSourceNameRegexConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami.name_regex_filtered_ami"),
					resource.TestMatchResourceAttr("data.aws_ami.name_regex_filtered_ami", "image_id", regexp.MustCompile("^ami-")),
				),
			},
		},
	})
}

func testAccCheckAwsAmiDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AMI data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AMI data source ID not set")
		}
		return nil
	}
}

// Using NAT AMIs for testing - I would expect with NAT gateways now a thing,
// that this will possibly be deprecated at some point in time. Other candidates
// for testing this after that may be Ubuntu's AMI's, or Amazon's regular
// Amazon Linux AMIs.
const testAccCheckAwsAmiDataSourceConfig = `
data "aws_ami" "nat_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-vpc-nat*"]
  }
  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
  filter {
    name = "block-device-mapping.volume-type"
    values = ["standard"]
  }
}
`

// Windows image test.
const testAccCheckAwsAmiDataSourceWindowsConfig = `
data "aws_ami" "windows_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["Windows_Server-2012-R2*"]
  }
  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
  filter {
    name = "block-device-mapping.volume-type"
    values = ["gp2"]
  }
}
`

// Instance store test - using Ubuntu images
const testAccCheckAwsAmiDataSourceInstanceStoreConfig = `
data "aws_ami" "instance_store_ami" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name = "name"
    values = ["ubuntu/images/hvm-instance/ubuntu-trusty-14.04-amd64-server*"]
  }
  filter {
    name = "virtualization-type"
    values = ["hvm"]
  }
  filter {
    name = "root-device-type"
    values = ["instance-store"]
  }
}
`

// Testing name_regex parameter
const testAccCheckAwsAmiDataSourceNameRegexConfig = `
data "aws_ami" "name_regex_filtered_ami" {
  most_recent = true
  owners = ["amazon"]
  filter {
    name = "name"
    values = ["amzn-ami-*"]
  }
  name_regex = "^amzn-ami-\\d{3}[5].*-ecs-optimized"
}
`
