package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSLaunchConfigurationDataSource_basic(t *testing.T) {
	rInt := acctest.RandInt()
	rName := "data.aws_launch_configuration.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLaunchConfigurationDataSourceConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(rName, "image_id"),
					resource.TestCheckResourceAttrSet(rName, "instance_type"),
					resource.TestCheckResourceAttrSet(rName, "associate_public_ip_address"),
					resource.TestCheckResourceAttrSet(rName, "user_data"),
					resource.TestCheckResourceAttr(rName, "root_block_device.#", "1"),
					resource.TestCheckResourceAttr(rName, "ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr(rName, "ephemeral_block_device.#", "1"),
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

func testAccLaunchConfigurationDataSourceConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_configuration" "foo" {
  name = "terraform-test-%d"
  image_id = "ami-21f78e11"
  instance_type = "m1.small"
  associate_public_ip_address = true
  user_data = "foobar-user-data"

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
    iops = 100
  }
  ephemeral_block_device {
    device_name = "/dev/sde"
    virtual_name = "ephemeral0"
  }
}

data "aws_launch_configuration" "foo" {
  name = "${aws_launch_configuration.foo.name}"
}
`, rInt)
}

func testAccLaunchConfigurationDataSourceConfig_securityGroups(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_security_group" "test" {
  name = "terraform-test_%d"
  vpc_id = "${aws_vpc.test.id}"
}

resource "aws_launch_configuration" "test" {
  name = "terraform-test-%d"
  image_id = "ami-21f78e11"
  instance_type = "m1.small"
  security_groups = ["${aws_security_group.test.id}"]
}

data "aws_launch_configuration" "foo" {
  name = "${aws_launch_configuration.test.name}"
}
`, rInt, rInt)
}
