package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func TestAccAWSAutoscalingGroupTag_basic(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, autoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoscalingGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", "value1"),
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

func TestAccAWSAutoscalingGroupTag_disappears(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, autoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoscalingGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAutoscalingGroupTag(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAutoscalingGroupTag_Value(t *testing.T) {
	resourceName := "aws_autoscaling_group_tag.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, autoscaling.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoscalingGroupTagDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAutoscalingGroupTagConfig("key1", "value1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoscalingGroupTagExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tag.0.key", "key1"),
					resource.TestCheckResourceAttr(resourceName, "tag.0.value", "value1updated"),
				),
			},
		},
	})
}

func testAccCheckAutoscalingGroupTagDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_group_tag" {
			continue
		}

		asgName, key, err := extractAutoscalingGroupNameAndKeyFromAutoscalingGroupTagID(rs.Primary.ID)

		if err != nil {
			return err
		}

		exists, _, err := keyvaluetags.AutoscalingGetTag(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, key)

		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("Tag (%s) for resource (%s) still exists", key, asgName)
		}
	}

	return nil
}

func testAccCheckAutoscalingGroupTagExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		asgName, key, err := extractAutoscalingGroupNameAndKeyFromAutoscalingGroupTagID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		exists, _, err := keyvaluetags.AutoscalingGetTag(conn, asgName, autoscalingTagResourceTypeAutoScalingGroup, key)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Tag (%s) for resource (%s) not found", key, asgName)
		}

		return nil
	}
}

func testAccAutoscalingGroupTagConfig(key string, value string) string {
	return fmt.Sprintf(`
data "aws_ami" "latest_al2" {
	owners      = ["amazon"]
	most_recent = true

	filter {
		name   = "name"
		values = ["amzn2-ami-hvm-*-x86_64-ebs"]
	}
}

resource "aws_launch_template" "test" {
	name_prefix   = "terraform-test-"
	image_id      = data.aws_ami.latest_al2.id
	instance_type = "t2.nano"
}

data "aws_availability_zones" "available" {
	state = "available"
}

resource "aws_autoscaling_group" "test" {
	lifecycle {
		ignore_changes = [tag]
	}

	availability_zones = data.aws_availability_zones.available.names

	min_size = 0
	max_size = 0

	launch_template {
		id      = aws_launch_template.test.id
		version = "$Latest"
	}
}

resource "aws_autoscaling_group_tag" "test" {
	asg_name = aws_autoscaling_group.test.name

	tag {
		key   = %[1]q
		value = %[2]q

		propagate_at_launch = true
	}
}
`, key, value)
}
