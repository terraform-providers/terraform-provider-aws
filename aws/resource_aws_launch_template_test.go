package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLaunchTemplate_importBasic(t *testing.T) {
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rInt),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLaunchTemplate_importData(t *testing.T) {
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_data(rInt),
			},
			{
				ResourceName:      resName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLaunchTemplate_basic(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrSet(resName, "arn"),
					resource.TestCheckResourceAttr(resName, "ebs_optimized", ""),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_disappears(t *testing.T) {
	var launchTemplate ec2.LaunchTemplate
	resourceName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &launchTemplate),
					testAccCheckAWSLaunchTemplateDisappears(&launchTemplate),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLaunchTemplate_BlockDeviceMappings_EBS(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
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

func TestAccAWSLaunchTemplate_BlockDeviceMappings_EBS_DeleteOnTermination(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/sda1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/sda1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", "false"),
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

func TestAccAWSLaunchTemplate_EbsOptimized(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "\"true\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "\"false\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "\"\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_data(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_data(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "disable_api_termination"),
					resource.TestCheckResourceAttr(resName, "ebs_optimized", "false"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_asg_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "default_version", "1"),
					resource.TestCheckResourceAttr(resName, "latest_version", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "1"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_asg_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "default_version", "1"),
					resource.TestCheckResourceAttr(resName, "latest_version", "2"),
					resource.TestCheckResourceAttrSet(resName, "instance_type"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_tags(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					testAccCheckTags(&template.Tags, "foo", "bar"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_tagsUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					testAccCheckTags(&template.Tags, "foo", ""),
					testAccCheckTags(&template.Tags, "bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_capacityReservation_preference(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_capacityReservation_preference(rInt, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_capacityReservation_target(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_capacityReservation_target(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_creditSpecification_nonBurstable(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resName := "aws_launch_template.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_creditSpecification(rName, "m1.small", "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_creditSpecification_t2(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resName := "aws_launch_template.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_creditSpecification(rName, "t2.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_creditSpecification_t3(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resName := "aws_launch_template.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_creditSpecification(rName, "t3.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.0.cpu_credits", "unlimited"),
				),
			},
		},
	})
}

// Reference: https://github.com/terraform-providers/terraform-provider-aws/issues/6757
func TestAccAWSLaunchTemplate_IamInstanceProfile_EmptyConfigurationBlock(t *testing.T) {
	var template1 ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfigIamInstanceProfileEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLaunchTemplate_networkInterface(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_networkInterface,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resName, "network_interfaces.0.associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_networkInterface_ipv6Addresses(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_networkInterface_ipv6Addresses,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resName, "network_interfaces.0.ipv6_addresses.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_networkInterface_ipv6AddressCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_ipv6_count(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resName, "network_interfaces.0.ipv6_address_count", "1"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_instanceMarketOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	var group autoscaling.Group
	templateName := "aws_launch_template.test"
	groupName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_instanceMarketOptions_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(templateName, &template),
					testAccCheckAWSAutoScalingGroupExists(groupName, &group),
					resource.TestCheckResourceAttr(templateName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(templateName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.0.version", "1"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_instanceMarketOptions_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(templateName, &template),
					testAccCheckAWSAutoScalingGroupExists(groupName, &group),
					resource.TestCheckResourceAttr(templateName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(templateName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.0.version", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_licenseSpecification(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.example"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_licenseSpecification(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "license_specification.#", "1"),
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

func testAccCheckAWSLaunchTemplateDisappears(launchTemplate *ec2.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteLaunchTemplateInput{
			LaunchTemplateId: launchTemplate.LaunchTemplateId,
		}

		_, err := conn.DeleteLaunchTemplate(input)

		return err
	}
}

func testAccAWSLaunchTemplateConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "foo_%d"

  tags = {
    foo = "bar"
  }
}
`, rInt)
}

func testAccAWSLaunchTemplateConfig_ipv6_count(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "set_ipv6_count_foo_%d"

  network_interfaces {
    ipv6_address_count = 1
  }
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

data "aws_availability_zones" "available" {}

resource "aws_launch_template" "test" {
  image_id = "${data.aws_ami.test.id}"
  name     = %q

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 15
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size         = 0
  min_size         = 0
  name             = %q

  launch_template {
    id      = "${aws_launch_template.test.id}"
    version = "${aws_launch_template.test.default_version}"
  }
}
`, rName, rName)
}

func testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName string, deleteOnTermination bool) string {
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

data "aws_availability_zones" "available" {}

resource "aws_launch_template" "test" {
  image_id = "${data.aws_ami.test.id}"
  name     = %q

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      delete_on_termination = %t
      volume_size           = 15
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size         = 0
  min_size         = 0
  name             = %q

  launch_template {
    id      = "${aws_launch_template.test.id}"
    version = "${aws_launch_template.test.default_version}"
  }
}
`, rName, deleteOnTermination, rName)
}

func testAccAWSLaunchTemplateConfig_EbsOptimized(rName, ebsOptimized string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  ebs_optimized = %s # allows "", false, true, "false", "true" values
  name          = %q
}
`, ebsOptimized, rName)
}

func testAccAWSLaunchTemplateConfig_data(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "foo_%d"

  block_device_mappings {
    device_name = "test"
  }

  disable_api_termination = true

  ebs_optimized = false

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
  tags = {
      Name = "test"
    }
  }
}
`, rInt)
}

func testAccAWSLaunchTemplateConfig_tagsUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "foo_%d"

  tags = {
    bar = "baz"
  }
}
`, rInt)
}

func testAccAWSLaunchTemplateConfig_capacityReservation_preference(rInt int, preference string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = "foo_%d"

	capacity_reservation_specification {
		capacity_reservation_preference = %q
	}
}
`, rInt, preference)
}

func testAccAWSLaunchTemplateConfig_capacityReservation_target(rInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_ec2_capacity_reservation" "test" {
	availability_zone = "${data.aws_availability_zones.available.names[0]}"
	instance_count    = 1
	instance_platform = "Linux/UNIX"
	instance_type     = "t2.micro"
}

resource "aws_launch_template" "foo" {
  name = "foo_%d"

	capacity_reservation_specification {
		capacity_reservation_target {
			capacity_reservation_id = "${aws_ec2_capacity_reservation.test.id}"
		}
	}
}
`, rInt)
}

func testAccAWSLaunchTemplateConfig_creditSpecification(rName, instanceType, cpuCredits string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  instance_type = %q
  name          = %q

  credit_specification {
    cpu_credits = %q
  }
}
`, instanceType, rName, cpuCredits)
}

func testAccAWSLaunchTemplateConfigIamInstanceProfileEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  iam_instance_profile {}
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_licenseSpecification(rInt int) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "example" {
  name                  = "Example"
  license_counting_type = "vCPU"
}

resource "aws_launch_template" "example" {
  name = "foo_%d"

	license_specification {
		license_configuration_arn = "${aws_licensemanager_license_configuration.example.id}"
	}
}
`, rInt)
}

const testAccAWSLaunchTemplateConfig_networkInterface = `
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.1.0.0/24"
}

resource "aws_network_interface" "test" {
  subnet_id = "${aws_subnet.test.id}"
}

resource "aws_launch_template" "test" {
  name = "network-interface-launch-template"

  network_interfaces {
    network_interface_id = "${aws_network_interface.test.id}"
    ipv4_address_count = 2
  }
}
`

const testAccAWSLaunchTemplateConfig_networkInterface_ipv6Addresses = `
resource "aws_launch_template" "test" {
  name = "network-interface-ipv6-addresses-launch-template"

  network_interfaces {
    ipv6_addresses = [
      "0:0:0:0:0:ffff:a01:5",
      "0:0:0:0:0:ffff:a01:6",
    ]
  }
}
`

const testAccAWSLaunchTemplateConfig_asg_basic = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foo" {
  name_prefix = "foobar"
  image_id = "${data.aws_ami.test_ami.id}"
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_template {
    id = "${aws_launch_template.foo.id}"
    version = "${aws_launch_template.foo.latest_version}"
  }
}
`

const testAccAWSLaunchTemplateConfig_asg_update = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foo" {
  name_prefix = "foobar"
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.nano"
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_template {
    id = "${aws_launch_template.foo.id}"
    version = "${aws_launch_template.foo.latest_version}"
  }
}
`

const testAccAWSLaunchTemplateConfig_instanceMarketOptions_basic = `
data "aws_ami" "test" {
  most_recent = true

  filter {
    name = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix = "instance_market_options"
  image_id = "${data.aws_ami.test.id}"

  instance_market_options {
    market_type = "spot"
    spot_options {
      spot_instance_type = "one-time"
    }
  }
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  min_size = 0
  max_size = 0

  launch_template {
    id = "${aws_launch_template.test.id}"
    version = "${aws_launch_template.test.latest_version}"
  }
}
`

const testAccAWSLaunchTemplateConfig_instanceMarketOptions_update = `
data "aws_ami" "test" {
  most_recent = true

  filter {
    name = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix = "instance_market_options"
  image_id = "${data.aws_ami.test.id}"
  instance_type = "t2.micro"

  instance_market_options {
    market_type = "spot"
    spot_options {
      spot_instance_type = "one-time"
    }
  }
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  min_size = 0
  max_size = 0

  launch_template {
    id = "${aws_launch_template.test.id}"
    version = "${aws_launch_template.test.latest_version}"
  }
}
`
