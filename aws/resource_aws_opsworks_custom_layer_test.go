package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
)

// These tests assume the existence of predefined Opsworks IAM roles named `aws-opsworks-ec2-role`
// and `aws-opsworks-service-role`, and Opsworks stacks named `tf-acc`.

func TestAccAWSOpsworksCustomLayer_basic(t *testing.T) {
	name := acctest.RandString(10)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.tf-acc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksCustomLayerConfigVpcCreate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksCustomLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", "true"),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.1368285564", "git"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.2937857443", "golang"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.number_of_disks", "2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.mount_point", "/home"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.size", "100"),
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

func TestAccAWSOpsworksCustomLayer_noVPC(t *testing.T) {
	stackName := fmt.Sprintf("tf-%d", acctest.RandInt())
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.tf-acc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksCustomLayerConfigNoVpcCreate(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksCustomLayerExists(resourceName, &opslayer),
					testAccCheckAWSOpsworksCreateLayerAttributes(&opslayer, stackName),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", "true"),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.1368285564", "git"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.2937857443", "golang"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.number_of_disks", "2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.mount_point", "/home"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.size", "100"),
				),
			},
			{
				Config: testAccAwsOpsworksCustomLayerConfigUpdate(stackName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "120"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.1368285564", "git"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.2937857443", "golang"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.4101929740", "subversion"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.type", "gp2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.number_of_disks", "2"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.mount_point", "/home"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.3575749636.size", "100"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.1266957920.type", "io1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.1266957920.number_of_disks", "4"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.1266957920.mount_point", "/var"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.1266957920.size", "100"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.1266957920.raid_level", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.1266957920.iops", "3000"),
					resource.TestCheckResourceAttr(resourceName, "custom_json", `{"layer_key":"layer_value2"}`),
				),
			},
		},
	})
}

func TestAccAWSOpsworksCustomLayer_autoscaling(t *testing.T) {
	stackName := fmt.Sprintf("tf-%d", acctest.RandInt())
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.tf-acc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksCustomLayerAutoscalingGroup(stackName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksCustomLayerExists(resourceName, &opslayer),
					testAccCheckAWSOpsworksCreateLayerAttributes(&opslayer, stackName),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "enable_load_based_autoscaling", "false"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.#", "0"),
				),
			},
			{
				Config: testAccAwsOpsworksCustomLayerAutoscalingGroup(stackName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksCustomLayerExists(resourceName, &opslayer),
					testAccCheckAWSOpsworksCreateLayerAttributes(&opslayer, stackName),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "enable_load_based_autoscaling", "true"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.downscaling.0.cpu_threshold", "20"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.downscaling.0.ignore_metrics_time", "15"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.downscaling.0.instance_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.downscaling.0.load_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.downscaling.0.memory_threshold", "20"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.downscaling.0.thresholds_wait_time", "30"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.upscaling.0.cpu_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.upscaling.0.ignore_metrics_time", "15"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.upscaling.0.instance_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.upscaling.0.load_threshold", "10"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.upscaling.0.memory_threshold", "80"),
					resource.TestCheckResourceAttr(resourceName, "load_based_autoscaling.0.upscaling.0.thresholds_wait_time", "30"),
				),
			},
		},
	})

}

func testAccCheckAWSOpsworksCustomLayerExists(
	n string, opslayer *opsworks.Layer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).opsworksconn

		params := &opsworks.DescribeLayersInput{
			LayerIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeLayers(params)

		if err != nil {
			return err
		}

		if v := len(resp.Layers); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opslayer = *resp.Layers[0]

		return nil
	}
}

func testAccCheckAWSOpsworksCreateLayerAttributes(
	opslayer *opsworks.Layer, stackName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opslayer.Name != stackName {
			return fmt.Errorf("Unexpected name: %s", *opslayer.Name)
		}

		if *opslayer.AutoAssignElasticIps {
			return fmt.Errorf(
				"Unexpected AutoAssignElasticIps: %t", *opslayer.AutoAssignElasticIps)
		}

		if !*opslayer.EnableAutoHealing {
			return fmt.Errorf(
				"Unexpected EnableAutoHealing: %t", *opslayer.EnableAutoHealing)
		}

		if !*opslayer.LifecycleEventConfiguration.Shutdown.DelayUntilElbConnectionsDrained {
			return fmt.Errorf(
				"Unexpected DelayUntilElbConnectionsDrained: %t",
				*opslayer.LifecycleEventConfiguration.Shutdown.DelayUntilElbConnectionsDrained)
		}

		if *opslayer.LifecycleEventConfiguration.Shutdown.ExecutionTimeout != 300 {
			return fmt.Errorf(
				"Unexpected ExecutionTimeout: %d",
				*opslayer.LifecycleEventConfiguration.Shutdown.ExecutionTimeout)
		}

		if v := len(opslayer.CustomSecurityGroupIds); v != 2 {
			return fmt.Errorf("Expected 2 customSecurityGroupIds, got %d", v)
		}

		expectedPackages := []*string{
			aws.String("git"),
			aws.String("golang"),
		}

		if !reflect.DeepEqual(expectedPackages, opslayer.Packages) {
			return fmt.Errorf("Unexpected Packages: %v", aws.StringValueSlice(opslayer.Packages))
		}

		expectedEbsVolumes := []*opsworks.VolumeConfiguration{
			{
				Encrypted:     aws.Bool(false),
				MountPoint:    aws.String("/home"),
				NumberOfDisks: aws.Int64(2),
				RaidLevel:     aws.Int64(0),
				Size:          aws.Int64(100),
				VolumeType:    aws.String("gp2"),
			},
		}

		if !reflect.DeepEqual(expectedEbsVolumes, opslayer.VolumeConfigurations) {
			return fmt.Errorf("Unnexpected VolumeConfiguration: %s", opslayer.VolumeConfigurations)
		}

		return nil
	}
}

func testAccCheckAwsOpsworksCustomLayerDestroy(s *terraform.State) error {
	opsworksconn := testAccProvider.Meta().(*AWSClient).opsworksconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_custom_layer" {
			continue
		}
		req := &opsworks.DescribeLayersInput{
			LayerIds: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		_, err := opsworksconn.DescribeLayers(req)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				if awserr.Code() == "ResourceNotFoundException" {
					// not found, good to go
					return nil
				}
			}
			return err
		}
	}

	return fmt.Errorf("Fall through error on OpsWorks custom layer test")
}

func testAccAwsOpsworksCustomLayerSecurityGroups(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-layer1" {
  name = "%s-layer1"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-layer2" {
  name = "%s-layer2"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, name, name)
}

func testAccAwsOpsworksCustomLayerConfigNoVpcCreate(name string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id               = "${aws_opsworks_stack.tf-acc.id}"
  name                   = "%s"
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    "${aws_security_group.tf-ops-acc-layer1.id}",
    "${aws_security_group.tf-ops-acc-layer2.id}",
  ]
  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300
  system_packages = [
    "git",
    "golang",
  ]
  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }
}

%s

%s 
`, name, testAccAwsOpsworksStackConfigNoVpcCreate(name), testAccAwsOpsworksCustomLayerSecurityGroups(name))
}

func testAccAwsOpsworksCustomLayerConfigVpcCreate(name string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id               = "${aws_opsworks_stack.tf-acc.id}"
  name                   = "%s"
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = [
    "${aws_security_group.tf-ops-acc-layer1.id}",
    "${aws_security_group.tf-ops-acc-layer2.id}",
  ]

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  system_packages = [
    "git",
    "golang",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }
}

%s


%s

`, name, testAccAwsOpsworksStackConfigVpcCreate(name), testAccAwsOpsworksCustomLayerSecurityGroups(name))
}

func testAccAwsOpsworksCustomLayerConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-layer3" {
  name = "tf-ops-acc-layer-%[1]s"
  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id               = "${aws_opsworks_stack.tf-acc.id}"
  name                   = "%[1]s"
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    "${aws_security_group.tf-ops-acc-layer1.id}",
    "${aws_security_group.tf-ops-acc-layer2.id}",
    "${aws_security_group.tf-ops-acc-layer3.id}",
  ]
  drain_elb_on_shutdown     = false
  instance_shutdown_timeout = 120
  system_packages = [
    "git",
    "golang",
    "subversion",
  ]
  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }
  ebs_volume {
    type            = "io1"
    number_of_disks = 4
    mount_point     = "/var"
    size            = 100
    raid_level      = 1
    iops            = 3000
  }
  custom_json = "{\"layer_key\": \"layer_value2\"}"
}

%s

%s 
`, name, testAccAwsOpsworksStackConfigNoVpcCreate(name), testAccAwsOpsworksCustomLayerSecurityGroups(name))
}

func testAccAwsOpsworksCustomLayerAutoscalingGroup(name string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id               = "${aws_opsworks_stack.tf-acc.id}"
  name                   = "%s"
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    "${aws_security_group.tf-ops-acc-layer1.id}",
    "${aws_security_group.tf-ops-acc-layer2.id}",
  ]
  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300
  system_packages = [
    "git",
    "golang",
  ]
  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }

  enable_load_based_autoscaling = %t
  load_based_autoscaling {
    downscaling {
      cpu_threshold        = 20
      ignore_metrics_time  = 15
      instance_count       = 2
      load_threshold       = 5
      memory_threshold     = 20
      thresholds_wait_time = 30
    }

    upscaling {
      cpu_threshold        = 80
      ignore_metrics_time  = 15
      instance_count       = 3
      load_threshold       = 10
      memory_threshold     = 80
      thresholds_wait_time = 30
    }
  }
}

%s

%s 
`, name, enable, testAccAwsOpsworksStackConfigNoVpcCreate(name), testAccAwsOpsworksCustomLayerSecurityGroups(name))
}
