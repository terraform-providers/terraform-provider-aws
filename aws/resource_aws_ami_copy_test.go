package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAMICopy_basic(t *testing.T) {
	var image ec2.Image
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMICopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMICopyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMICopyExists(resourceName, &image),
					testAccCheckAWSAMICopyAttributes(&image, rName),
				),
			},
		},
	})
}

func TestAccAWSAMICopy_Description(t *testing.T) {
	var image ec2.Image
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMICopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMICopyConfigDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMICopyExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAWSAMICopyConfigDescription(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMICopyExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccAWSAMICopy_EnaSupport(t *testing.T) {
	var image ec2.Image
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ami_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAMICopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAMICopyConfigENASupport(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAMICopyExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "ena_support", "true"),
				),
			},
		},
	})
}

func testAccCheckAWSAMICopyExists(resourceName string, image *ec2.Image) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		output, err := conn.DescribeImages(input)
		if err != nil {
			return err
		}

		if len(output.Images) == 0 || aws.StringValue(output.Images[0].ImageId) != rs.Primary.ID {
			return fmt.Errorf("AMI %q not found", rs.Primary.ID)
		}

		*image = *output.Images[0]

		return nil
	}
}

func testAccCheckAWSAMICopyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami_copy" {
			continue
		}

		input := &ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(rs.Primary.ID)},
		}
		output, err := conn.DescribeImages(input)
		if err != nil {
			return err
		}

		if output != nil && len(output.Images) > 0 && aws.StringValue(output.Images[0].ImageId) == rs.Primary.ID {
			return fmt.Errorf("AMI %q still exists in state: %s", rs.Primary.ID, aws.StringValue(output.Images[0].State))
		}
	}

	// Check for managed EBS snapshots
	return testAccCheckAWSEbsSnapshotDestroy(s)
}

func testAccCheckAWSAMICopyAttributes(image *ec2.Image, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if expected := "available"; aws.StringValue(image.State) != expected {
			return fmt.Errorf("invalid image state; expected %s, got %s", expected, aws.StringValue(image.State))
		}
		if expected := "machine"; aws.StringValue(image.ImageType) != expected {
			return fmt.Errorf("wrong image type; expected %s, got %s", expected, aws.StringValue(image.ImageType))
		}
		if expected := expectedName; aws.StringValue(image.Name) != expected {
			return fmt.Errorf("wrong name; expected %s, got %s", expected, aws.StringValue(image.Name))
		}

		snapshots := []string{}
		for _, bdm := range image.BlockDeviceMappings {
			// The snapshot ID might not be set,
			// even for a block device that is an
			// EBS volume.
			if bdm.Ebs != nil && bdm.Ebs.SnapshotId != nil {
				snapshots = append(snapshots, aws.StringValue(bdm.Ebs.SnapshotId))
			}
		}

		if expected := 1; len(snapshots) != expected {
			return fmt.Errorf("wrong number of snapshots; expected %v, got %v", expected, len(snapshots))
		}

		return nil
	}
}

func testAccAWSAMICopyConfigBase() string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {}
data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1

  tags = {
    Name = "tf-acc-test-ami-copy"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "tf-acc-test-ami-copy"
  }
}
`)
}

func testAccAWSAMICopyConfig(rName string) string {
	return testAccAWSAMICopyConfigBase() + fmt.Sprintf(`
resource "aws_ami" "test" {
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = "${aws_ebs_snapshot.test.id}"
  }
}

resource "aws_ami_copy" "test" {
  name              = %q
  source_ami_id     = "${aws_ami.test.id}"
  source_ami_region = "${data.aws_region.current.name}"
}
`, rName, rName)
}

func testAccAWSAMICopyConfigDescription(rName, description string) string {
	return testAccAWSAMICopyConfigBase() + fmt.Sprintf(`
resource "aws_ami" "test" {
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = "${aws_ebs_snapshot.test.id}"
  }
}

resource "aws_ami_copy" "test" {
  description       = %q
  name              = %q
  source_ami_id     = "${aws_ami.test.id}"
  source_ami_region = "${data.aws_region.current.name}"
}
`, rName, description, rName)
}

func testAccAWSAMICopyConfigENASupport(rName string) string {
	return testAccAWSAMICopyConfigBase() + fmt.Sprintf(`
resource "aws_ami" "test" {
  ena_support         = true
  name                = "%s-source"
  virtualization_type = "hvm"
  root_device_name    = "/dev/sda1"

  ebs_block_device {
    device_name = "/dev/sda1"
    snapshot_id = "${aws_ebs_snapshot.test.id}"
  }
}

resource "aws_ami_copy" "test" {
    name              = "%s-copy"
    source_ami_id     = "${aws_ami.test.id}"
    source_ami_region = "${data.aws_region.current.name}"
}
`, rName, rName)
}
