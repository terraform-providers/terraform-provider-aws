package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`image/ami-.+`)),
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
					testAccMatchResourceAttrAccountID(resourceName, "owner_id"),
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
					resource.TestCheckResourceAttr(resourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "usage_operation", "RunInstances"),
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
					testAccMatchResourceAttrAccountID(resourceName, "owner_id"),
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
					resource.TestMatchResourceAttr(resourceName, "platform_details", regexp.MustCompile(`Windows`)),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestMatchResourceAttr(resourceName, "usage_operation", regexp.MustCompile(`^RunInstances`)),
				),
			},
		},
	})
}

func TestAccAWSAmiDataSource_instanceStore(t *testing.T) {
	resourceName := "data.aws_ami.amzn-ami-minimal-hvm-instance-store"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLatestAmazonLinuxHvmInstanceStoreAmiConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID(resourceName),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "creation_date", regexp.MustCompile("^20[0-9]{2}-")),
					resource.TestCheckResourceAttr(resourceName, "hypervisor", "xen"),
					resource.TestMatchResourceAttr(resourceName, "image_id", regexp.MustCompile("^ami-")),
					resource.TestMatchResourceAttr(resourceName, "image_location", regexp.MustCompile("amzn-ami-minimal-hvm")),
					resource.TestCheckResourceAttr(resourceName, "image_type", "machine"),
					resource.TestCheckResourceAttr(resourceName, "most_recent", "true"),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("amzn-ami-minimal-hvm")),
					testAccMatchResourceAttrAccountID(resourceName, "owner_id"),
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
					resource.TestCheckResourceAttr(resourceName, "platform_details", "Linux/UNIX"),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "usage_operation", "RunInstances"),
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

func TestAccAWSAmiDataSource_Gp3BlockDevice(t *testing.T) {
	resourceName := "aws_ami.test"
	datasourceName := "data.aws_ami.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiDataSourceConfigGp3BlockDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "architecture", resourceName, "architecture"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "block_device_mappings.#", resourceName, "ebs_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "image_id", resourceName, "id"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_device_name", resourceName, "root_device_name"),
					resource.TestCheckResourceAttr(datasourceName, "root_device_type", "ebs"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_snapshot_id", resourceName, "root_snapshot_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "sriov_net_support", resourceName, "sriov_net_support"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "virtualization_type", resourceName, "virtualization_type"),
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
    name   = "name"
    values = ["amzn-ami-vpc-nat*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "block-device-mapping.volume-type"
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
    name   = "name"
    values = ["Windows_Server-2012-R2*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "block-device-mapping.volume-type"
    values = ["gp2"]
  }
}
`

// Testing name_regex parameter
const testAccCheckAwsAmiDataSourceNameRegexConfig = `
data "aws_ami" "name_regex_filtered_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
  }

  name_regex = "^amzn-ami-min[a-z]{4}-hvm"
}
`

func testAccAmiDataSourceConfigGp3BlockDevice(rName string) string {
	return composeConfig(
		testAccAmiConfigGp3BlockDevice(rName),
		`
data "aws_caller_identity" "current" {}

data "aws_ami" "test" {
  owners = [data.aws_caller_identity.current.account_id]

  filter {
    name   = "image-id"
    values = [aws_ami.test.id]
  }
}
`)
}
