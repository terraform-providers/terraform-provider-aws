package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSLaunchConfigurationDataSource_basic(t *testing.T) {
	resourceName := "aws_launch_configuration.test"
	datasourceName := "data.aws_launch_configuration.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "image_id", resourceName, "image_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "associate_public_ip_address", resourceName, "associate_public_ip_address"),
					// Resource and data source user_data have differing representations in state.
					resource.TestCheckResourceAttrSet(datasourceName, "user_data"),
					resource.TestCheckResourceAttrPair(datasourceName, "root_block_device.#", resourceName, "root_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_block_device.#", resourceName, "ebs_block_device.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "ephemeral_block_device.#", resourceName, "ephemeral_block_device.#"),
					testAccMatchResourceAttrRegionalARN(datasourceName, "arn", "autoscaling", regexp.MustCompile(`launchConfiguration:.+`)),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfigurationDataSource_securityGroups(t *testing.T) {
	rInt := acctest.RandInt()
	rName := "data.aws_launch_configuration.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_securityGroups(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rName, "security_groups.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfigurationDataSource_ebsNoDevice(t *testing.T) {
	resourceName := "aws_launch_configuration.test"
	datasourceName := "data.aws_launch_configuration.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfigEbsNoDevice(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "image_id", resourceName, "image_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "ebs_block_device.#", resourceName, "ebs_block_device.#"),
				),
			},
		},
	})
}

func TestAccLaunchConfigurationDataSource_metadataOptions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_launch_configuration.test"
	resourceName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.#", resourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_endpoint", resourceName, "metadata_options.0.http_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_tokens", resourceName, "metadata_options.0.http_tokens"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_put_response_hop_limit", resourceName, "metadata_options.0.http_put_response_hop_limit"),
				),
			},
		},
	})
}

func testAccLaunchConfigurationDataSourceConfig_basic(rName string) string {
	return composeConfig(testAccLatestAmazonLinuxHvmEbsAmiConfig(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name                        = %[1]q
  image_id                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = "m1.small"
  associate_public_ip_address = true
  user_data                   = "foobar-user-data"

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }

  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
  }

  ebs_block_device {
    device_name = "/dev/sdc"
    volume_size = 10
    volume_type = "io1"
    iops        = 100
  }

  ephemeral_block_device {
    device_name  = "/dev/sde"
    virtual_name = "ephemeral0"
  }
}

data "aws_launch_configuration" "test" {
  name = aws_launch_configuration.test.name
}
`, rName))
}

func testAccLaunchConfigurationDataSourceConfig_securityGroups(rInt int) string {
	return testAccLatestAmazonLinuxHvmEbsAmiConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_security_group" "test" {
  name   = "terraform-test_%d"
  vpc_id = aws_vpc.test.id
}

resource "aws_launch_configuration" "test" {
  name            = "terraform-test-%d"
  image_id        = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type   = "m1.small"
  security_groups = [aws_security_group.test.id]
}

data "aws_launch_configuration" "foo" {
  name = aws_launch_configuration.test.name
}
`, rInt, rInt)
}

func testAccLaunchConfigurationDataSourceConfig_metadataOptions(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t3.nano"
  name          = %[1]q
  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }
}

data "aws_launch_configuration" "test" {
  name = aws_launch_configuration.test.name
}
`, rName))
}

func testAccLaunchConfigurationDataSourceConfigEbsNoDevice(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "m1.small"

  ebs_block_device {
    device_name = "/dev/sda2"
    no_device   = true
  }
}

data "aws_launch_configuration" "test" {
  name = aws_launch_configuration.test.name
}
`, rName))
}
