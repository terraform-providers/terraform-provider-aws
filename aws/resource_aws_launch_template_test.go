package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLaunchTemplate_basic(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "default_version", "1"),
					resource.TestCheckResourceAttr(resName, "latest_version", "1"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_BlockDeviceMappings_EBS(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/sda1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
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

func TestAccAWSLaunchTemplate_data(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_data(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "disable_api_termination"),
					resource.TestCheckResourceAttr(resName, "elastic_gpu_specifications.#", "1"),
					resource.TestCheckResourceAttr(resName, "iam_instance_profile.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "image_id"),
					resource.TestCheckResourceAttrSet(resName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttr(resName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "instance_type"),
					resource.TestCheckResourceAttrSet(resName, "kernel_id"),
					resource.TestCheckResourceAttrSet(resName, "key_name"),
					resource.TestCheckResourceAttr(resName, "monitoring.#", "1"),
					resource.TestCheckResourceAttr(resName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resName, "placement.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "ram_disk_id"),
					resource.TestCheckResourceAttr(resName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resName, "tag_specifications.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_update(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "default_version", "1"),
					resource.TestCheckResourceAttr(resName, "latest_version", "1"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_data(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "default_version", "1"),
					resource.TestCheckResourceAttr(resName, "latest_version", "2"),
					resource.TestCheckResourceAttrSet(resName, "image_id"),
				),
			},
		},
	})
}

func testAccCheckAWSLaunchTemplateExists(n string, t *ec2.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Launch Template ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.LaunchTemplates) != 1 || *resp.LaunchTemplates[0].LaunchTemplateId != rs.Primary.ID {
			return fmt.Errorf("Launch Template not found")
		}

		*t = *resp.LaunchTemplates[0]

		return nil
	}
}

func testAccCheckAWSLaunchTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_template" {
			continue
		}

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(resp.LaunchTemplates) != 0 && *resp.LaunchTemplates[0].LaunchTemplateId == rs.Primary.ID {
				return fmt.Errorf("Launch Template still exists")
			}
		}

		if isAWSErr(err, "InvalidLaunchTemplateId.NotFound", "") {
			log.Printf("[WARN] launch template (%s) not found.", rs.Primary.ID)
			continue
		}
		return err
	}

	return nil
}

func testAccAWSLaunchTemplateConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "foo_%d"
}
`, rInt)
}

func testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_launch_template" "test" {
  image_id = "${data.aws_ami.test.id}"
  name     = "%s"

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 15
    }
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_data(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "foo_%d"

  block_device_mappings {
    device_name = "test"
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  elastic_gpu_specifications {
    type = "test"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-12a3b456"

  instance_initiated_shutdown_behavior = "terminate"

  instance_market_options {
    market_type = "spot"
  }

  instance_type = "t2.micro"

  kernel_id = "aki-a12bc3de"

  key_name = "test"

  monitoring {
    enabled = true
  }

  network_interfaces {
    associate_public_ip_address = true
    network_interface_id = "eni-123456ab"
    security_groups = ["sg-1a23bc45"]
  }

  placement {
    availability_zone = "us-west-2b"
  }

  ram_disk_id = "ari-a12bc3de"

  vpc_security_group_ids = ["sg-12a3b45c"]

  tag_specifications {
    resource_type = "instance"
    tags {
      Name = "test"
    }
  }
}
`, rInt)
}
