package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAMI_basic(t *testing.T) {
	var ami ec2.Image
	resourceName := "aws_ami.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestMatchResourceAttr(resourceName, "root_snapshot_id", regexp.MustCompile("^snap-")),
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

func TestAccAWSAMI_snapshotSize(t *testing.T) {
	var ami ec2.Image
	var bd ec2.BlockDeviceMapping
	resourceName := "aws_ami.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	expectedDevice := &ec2.EbsBlockDevice{
		DeleteOnTermination: aws.Bool(true),
		Encrypted:           aws.Bool(false),
		Iops:                aws.Int64(0),
		VolumeSize:          aws.Int64(20),
		VolumeType:          aws.String("standard"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAmiConfig_snapshotSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAmiExists(resourceName, &ami),
					testAccCheckAmiBlockDevice(&ami, &bd, "/dev/sda1"),
					testAccCheckAmiEbsBlockDevice(&bd, expectedDevice),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func testAccCheckAmiDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami" {
			continue
		}

		// Try to find the AMI
		log.Printf("AMI-ID: %s", rs.Primary.ID)
		DescribeAmiOpts := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeImages(DescribeAmiOpts)
		if err != nil {
			if isAWSErr(err, "InvalidAMIID", "NotFound") {
				log.Printf("[DEBUG] AMI not found, passing")
				return nil
			}
			return err
		}

		if len(resp.Images) > 0 {
			state := resp.Images[0].State
			return fmt.Errorf("AMI %s still exists in the state: %s.", *resp.Images[0].ImageId, *state)
		}
	}
	return nil
}

func testAccCheckAmiExists(n string, ami *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("AMI Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMI ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		var resp *ec2.DescribeImagesOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			opts := &ec2.DescribeImagesInput{
				ImageIds: []*string{aws.String(rs.Primary.ID)},
			}
			var err error
			resp, err = conn.DescribeImages(opts)
			if err != nil {
				// This can be just eventual consistency
				awsErr, ok := err.(awserr.Error)
				if ok && awsErr.Code() == "InvalidAMIID.NotFound" {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("Unable to find AMI after retries: %s", err)
		}

		if len(resp.Images) == 0 {
			return fmt.Errorf("AMI not found")
		}
		*ami = *resp.Images[0]
		return nil
	}
}

func testAccCheckAmiBlockDevice(ami *ec2.Image, blockDevice *ec2.BlockDeviceMapping, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		devices := make(map[string]*ec2.BlockDeviceMapping)
		for _, device := range ami.BlockDeviceMappings {
			devices[*device.DeviceName] = device
		}

		// Check if the block device exists
		if _, ok := devices[n]; !ok {
			return fmt.Errorf("block device doesn't exist: %s", n)
		}

		*blockDevice = *devices[n]
		return nil
	}
}

func testAccCheckAmiEbsBlockDevice(bd *ec2.BlockDeviceMapping, ed *ec2.EbsBlockDevice) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Test for things that ed has, don't care about unset values
		cd := bd.Ebs
		if ed.VolumeType != nil {
			if *ed.VolumeType != *cd.VolumeType {
				return fmt.Errorf("Volume type mismatch. Expected: %s Got: %s",
					*ed.VolumeType, *cd.VolumeType)
			}
		}
		if ed.DeleteOnTermination != nil {
			if *ed.DeleteOnTermination != *cd.DeleteOnTermination {
				return fmt.Errorf("DeleteOnTermination mismatch. Expected: %t Got: %t",
					*ed.DeleteOnTermination, *cd.DeleteOnTermination)
			}
		}
		if ed.Encrypted != nil {
			if *ed.Encrypted != *cd.Encrypted {
				return fmt.Errorf("Encrypted mismatch. Expected: %t Got: %t",
					*ed.Encrypted, *cd.Encrypted)
			}
		}
		// Integer defaults need to not be `0` so we don't get a panic
		if ed.Iops != nil && *ed.Iops != 0 {
			if *ed.Iops != *cd.Iops {
				return fmt.Errorf("IOPS mismatch. Expected: %d Got: %d",
					*ed.Iops, *cd.Iops)
			}
		}
		if ed.VolumeSize != nil && *ed.VolumeSize != 0 {
			if *ed.VolumeSize != *cd.VolumeSize {
				return fmt.Errorf("Volume Size mismatch. Expected: %d Got: %d",
					*ed.VolumeSize, *cd.VolumeSize)
			}
		}

		return nil
	}
}

func testAccAmiConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "foo" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size = 8
  tags = {
    Name = "testAccAmiConfig_basic"
  }
}

resource "aws_ebs_snapshot" "foo" {
  volume_id = "${aws_ebs_volume.foo.id}"

  tags = {
    Name = "testAccAmiConfig_basic"
  }
}

resource "aws_ami" "test" {
  ena_support         = true
  name                = %q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = "${aws_ebs_snapshot.foo.id}"
  }
}
`, rName)
}

func testAccAmiConfig_snapshotSize(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "foo" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size = 20
  tags = {
    Name = "testAccAmiConfig_snapshotSize"
  }
}

resource "aws_ebs_snapshot" "foo" {
  volume_id = "${aws_ebs_volume.foo.id}"

  tags = {
    Name = "TestAccAWSAMI_snapshotSize"
  }
}

resource "aws_ami" "test" {
  name                = %q
  root_device_name    = "/dev/sda1"
  virtualization_type = "hvm"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = "${aws_ebs_snapshot.foo.id}"
  }
}
`, rName)
}
