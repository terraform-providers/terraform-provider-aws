package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_launch_configuration", &resource.Sweeper{
		Name:         "aws_launch_configuration",
		Dependencies: []string{"aws_autoscaling_group"},
		F:            testSweepLaunchConfigurations,
	})
}

func testSweepLaunchConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	autoscalingconn := client.(*AWSClient).autoscalingconn

	resp, err := autoscalingconn.DescribeLaunchConfigurations(&autoscaling.DescribeLaunchConfigurationsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping AutoScaling Launch Configuration sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving launch configuration: %s", err)
	}

	if len(resp.LaunchConfigurations) == 0 {
		log.Print("[DEBUG] No aws launch configurations to sweep")
		return nil
	}

	prefixes := []string{
		"foobar",
		"terraform-",
		"tf-acc-",
		"TestAcc",
	}
	for _, lc := range resp.LaunchConfigurations {
		name := *lc.LaunchConfigurationName
		skip := true
		for _, prefix := range prefixes {
			if strings.HasPrefix(name, prefix) {
				skip = false
			}
		}

		if skip {
			log.Printf("[INFO] Skipping Launch Configuration: %s", name)
			continue
		}

		log.Printf("[INFO] Deleting Launch Configuration: %s", name)
		_, err := autoscalingconn.DeleteLaunchConfiguration(
			&autoscaling.DeleteLaunchConfigurationInput{
				LaunchConfigurationName: aws.String(name),
			})
		if err != nil {
			if isAWSErr(err, "InvalidConfiguration.NotFound", "") || isAWSErr(err, "ValidationError", "") {
				return nil
			}
			return err
		}
	}

	return nil
}

func TestAccAWSLaunchConfiguration_importBasic(t *testing.T) {
	resourceName := "aws_launch_configuration.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationNoNameConfig(),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"associate_public_ip_address"},
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_basic(t *testing.T) {
	var conf autoscaling.LaunchConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationNoNameConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					testAccCheckAWSLaunchConfigurationGeneratedNamePrefix("aws_launch_configuration.bar", "terraform-"),
				),
			},
			{
				Config: testAccAWSLaunchConfigurationPrefixNameConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.baz", &conf),
					testAccCheckAWSLaunchConfigurationGeneratedNamePrefix("aws_launch_configuration.baz", "tf-acc-test-"),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_withBlockDevices(t *testing.T) {
	var conf autoscaling.LaunchConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					testAccCheckAWSLaunchConfigurationAttributes(&conf),
					resource.TestMatchResourceAttr("aws_launch_configuration.bar", "image_id", regexp.MustCompile("^ami-[0-9a-z]+")),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "instance_type", "m1.small"),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "associate_public_ip_address", "true"),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "spot_price", ""),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_updateRootBlockDevice(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationConfigWithRootBlockDevice(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "root_block_device.0.volume_size", "11"),
				),
			},
			{
				Config: testAccAWSLaunchConfigurationConfigWithRootBlockDeviceUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "root_block_device.0.volume_size", "20"),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_withSpotPrice(t *testing.T) {
	var conf autoscaling.LaunchConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationWithSpotPriceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "spot_price", "0.01"),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_withVpcClassicLink(t *testing.T) {
	var vpc ec2.Vpc
	var group ec2.SecurityGroup
	var conf autoscaling.LaunchConfiguration
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationConfig_withVpcClassicLink(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.foo", &conf),
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckAWSSecurityGroupExists("aws_security_group.foo", &group),
				),
			},
			{
				ResourceName:      "aws_launch_configuration.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_withIAMProfile(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationConfig_withIAMProfile(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
				),
			},
		},
	})
}

func testAccCheckAWSLaunchConfigurationWithEncryption(conf *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Map out the block devices by name, which should be unique.
		blockDevices := make(map[string]*autoscaling.BlockDeviceMapping)
		for _, blockDevice := range conf.BlockDeviceMappings {
			blockDevices[*blockDevice.DeviceName] = blockDevice
		}

		// Check if the root block device exists.
		if _, ok := blockDevices["/dev/sda1"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/sda1")
		} else if blockDevices["/dev/sda1"].Ebs.Encrypted != nil {
			return fmt.Errorf("root device should not include value for Encrypted")
		}

		// Check if the secondary block device exists.
		if _, ok := blockDevices["/dev/sdb"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/sdb")
		} else if !*blockDevices["/dev/sdb"].Ebs.Encrypted {
			return fmt.Errorf("block device isn't encrypted as expected: /dev/sdb")
		}

		return nil
	}
}

func TestAccAWSLaunchConfiguration_withEncryption(t *testing.T) {
	var conf autoscaling.LaunchConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationWithEncryption(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.baz", &conf),
					testAccCheckAWSLaunchConfigurationWithEncryption(&conf),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_updateEbsBlockDevices(t *testing.T) {
	var conf autoscaling.LaunchConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationWithEncryption(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.baz", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.baz", "ebs_block_device.1393547169.volume_size", "9"),
				),
			},
			{
				Config: testAccAWSLaunchConfigurationWithEncryptionUpdated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.baz", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.baz", "ebs_block_device.4131155854.volume_size", "10"),
				),
			},
		},
	})
}

func TestAccAWSLaunchConfiguration_ebs_noDevice(t *testing.T) {
	var conf autoscaling.LaunchConfiguration
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationConfigEbsNoDevice(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "ebs_block_device.#", "1"),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "ebs_block_device.3099842682.device_name", "/dev/sda2"),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "ebs_block_device.3099842682.no_device", "true"),
				),
			},
		},
	})
}
func TestAccAWSLaunchConfiguration_userData(t *testing.T) {
	var conf autoscaling.LaunchConfiguration

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchConfigurationConfig_userData(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "user_data", "3dc39dda39be1205215e776bad998da361a5955d"),
				),
			},
			{
				Config: testAccAWSLaunchConfigurationConfig_userDataBase64(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.bar", &conf),
					resource.TestCheckResourceAttr("aws_launch_configuration.bar", "user_data_base64", "aGVsbG8gd29ybGQ="),
				),
			},
		},
	})
}

func testAccCheckAWSLaunchConfigurationGeneratedNamePrefix(
	resource, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Resource not found")
		}
		name, ok := r.Primary.Attributes["name"]
		if !ok {
			return fmt.Errorf("Name attr not found: %#v", r.Primary.Attributes)
		}
		if !strings.HasPrefix(name, prefix) {
			return fmt.Errorf("Name: %q, does not have prefix: %q", name, prefix)
		}
		return nil
	}
}

func testAccCheckAWSLaunchConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_configuration" {
			continue
		}

		describe, err := conn.DescribeLaunchConfigurations(
			&autoscaling.DescribeLaunchConfigurationsInput{
				LaunchConfigurationNames: []*string{aws.String(rs.Primary.ID)},
			})

		if err == nil {
			if len(describe.LaunchConfigurations) != 0 &&
				*describe.LaunchConfigurations[0].LaunchConfigurationName == rs.Primary.ID {
				return fmt.Errorf("Launch Configuration still exists")
			}
		}

		// Verify the error
		if !isAWSErr(err, "InvalidLaunchConfiguration.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSLaunchConfigurationAttributes(conf *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*conf.LaunchConfigurationName, "terraform-") && !strings.HasPrefix(*conf.LaunchConfigurationName, "tf-acc-test-") {
			return fmt.Errorf("Bad name: %s", *conf.LaunchConfigurationName)
		}

		if *conf.InstanceType != "m1.small" {
			return fmt.Errorf("Bad instance_type: %s", *conf.InstanceType)
		}

		// Map out the block devices by name, which should be unique.
		blockDevices := make(map[string]*autoscaling.BlockDeviceMapping)
		for _, blockDevice := range conf.BlockDeviceMappings {
			blockDevices[*blockDevice.DeviceName] = blockDevice
		}

		// Check if the root block device exists.
		if _, ok := blockDevices["/dev/sda1"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/sda1")
		}

		// Check if the secondary block device exists.
		if _, ok := blockDevices["/dev/sdb"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/sdb")
		}

		// Check if the third block device exists.
		if _, ok := blockDevices["/dev/sdc"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/sdc")
		}

		// Check if the secondary block device exists.
		if _, ok := blockDevices["/dev/sdb"]; !ok {
			return fmt.Errorf("block device doesn't exist: /dev/sdb")
		}

		return nil
	}
}

func testAccCheckAWSLaunchConfigurationExists(n string, res *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Launch Configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		describeOpts := autoscaling.DescribeLaunchConfigurationsInput{
			LaunchConfigurationNames: []*string{aws.String(rs.Primary.ID)},
		}
		describe, err := conn.DescribeLaunchConfigurations(&describeOpts)

		if err != nil {
			return err
		}

		if len(describe.LaunchConfigurations) != 1 ||
			*describe.LaunchConfigurations[0].LaunchConfigurationName != rs.Primary.ID {
			return fmt.Errorf("Launch Configuration Group not found")
		}

		*res = *describe.LaunchConfigurations[0]

		return nil
	}
}

func testAccAWSLaunchConfigurationConfig_ami() string {
	return fmt.Sprintf(`
data "aws_ami" "ubuntu" {
  most_recent = true
  owners = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/ebs/ubuntu-precise-12.04-i386-server-2017*"]
  }
}
`)
}

func testAccAWSLaunchConfigurationConfigWithRootBlockDevice(rInt int) string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  name_prefix = "tf-acc-test-%d"
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "m1.small"
  user_data = "foobar-user-data"
  associate_public_ip_address = true

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }
}
`, rInt)
}

func testAccAWSLaunchConfigurationConfigWithRootBlockDeviceUpdated(rInt int) string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  name_prefix = "tf-acc-test-%d"
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "m1.small"
  user_data = "foobar-user-data"
  associate_public_ip_address = true

  root_block_device {
    volume_type = "gp2"
    volume_size = 20
  }
}
`, rInt)
}

func testAccAWSLaunchConfigurationConfig() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  name = "tf-acc-test-%d"
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "m1.small"
  user_data = "foobar-user-data"
  associate_public_ip_address = true

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
`, acctest.RandInt())
}

func testAccAWSLaunchConfigurationWithSpotPriceConfig() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  name = "tf-acc-test-%d"
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  spot_price = "0.01"
}
`, acctest.RandInt())
}

func testAccAWSLaunchConfigurationNoNameConfig() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  user_data = "foobar-user-data-change"
  associate_public_ip_address = false
}
`)
}

func testAccAWSLaunchConfigurationPrefixNameConfig() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "baz" {
  name_prefix = "tf-acc-test-"
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  user_data = "foobar-user-data-change"
  associate_public_ip_address = false
}
`)
}

func testAccAWSLaunchConfigurationWithEncryption() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "baz" {
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  associate_public_ip_address = false

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }
  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 9
    encrypted = true
  }
}
`)
}

func testAccAWSLaunchConfigurationWithEncryptionUpdated() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "baz" {
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  associate_public_ip_address = false

  root_block_device {
    volume_type = "gp2"
    volume_size = 11
  }
  ebs_block_device {
    device_name = "/dev/sdb"
    volume_size = 10
    encrypted = true
  }
}
`)
}

func testAccAWSLaunchConfigurationConfig_withVpcClassicLink(rInt int) string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "10.0.0.0/16"
    enable_classiclink = true
  tags = {
        Name = "terraform-testacc-launch-configuration-with-vpc-classic-link"
    }
}

resource "aws_security_group" "foo" {
  name = "tf-acc-test-%[1]d"
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_launch_configuration" "foo" {
  name = "tf-acc-test-%[1]d"
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"

  vpc_classic_link_id = "${aws_vpc.foo.id}"
  vpc_classic_link_security_groups = ["${aws_security_group.foo.id}"]
}
`, rInt)
}

func testAccAWSLaunchConfigurationConfig_withIAMProfile(rInt int) string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_iam_role" "role" {
  name  = "tf-acc-test-%[1]d"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_instance_profile" "profile" {
  name  = "tf-acc-test-%[1]d"
  roles = ["${aws_iam_role.role.name}"]
}

resource "aws_launch_configuration" "bar" {
  image_id             = "${data.aws_ami.ubuntu.id}"
  instance_type        = "t2.nano"
  iam_instance_profile = "${aws_iam_instance_profile.profile.name}"
}
`, rInt)
}

func testAccAWSLaunchConfigurationConfigEbsNoDevice(rInt int) string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  name_prefix = "tf-acc-test-%d"
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "m1.small"
  ebs_block_device {
    device_name = "/dev/sda2"
    no_device = true
  }
}
`, rInt)
}

func testAccAWSLaunchConfigurationConfig_userData() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  user_data = "foo:-with-character's"
  associate_public_ip_address = false
}
`)
}

func testAccAWSLaunchConfigurationConfig_userDataBase64() string {
	return testAccAWSLaunchConfigurationConfig_ami() + fmt.Sprintf(`
resource "aws_launch_configuration" "bar" {
  image_id = "${data.aws_ami.ubuntu.id}"
  instance_type = "t2.micro"
  user_data_base64 = "${base64encode("hello world")}"
  associate_public_ip_address = false
}
`)
}
