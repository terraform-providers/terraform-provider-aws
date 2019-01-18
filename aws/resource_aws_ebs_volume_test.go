package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ebs_volume", &resource.Sweeper{
		Name: "aws_ebs_volume",
		Dependencies: []string{
			"aws_instance",
		},
		F: testSweepEbsVolumes,
	})
}

func testSweepEbsVolumes(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	err = conn.DescribeVolumesPages(&ec2.DescribeVolumesInput{}, func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
		for _, volume := range page.Volumes {
			var nameTag string
			id := aws.StringValue(volume.VolumeId)

			if aws.StringValue(volume.State) != ec2.VolumeStateAvailable {
				log.Printf("[INFO] Skipping unavailable EC2 EBS Volume: %s", id)
				continue
			}

			for _, volumeTag := range volume.Tags {
				if aws.StringValue(volumeTag.Key) == "Name" {
					nameTag = aws.StringValue(volumeTag.Value)
					break
				}
			}

			if !strings.HasPrefix(nameTag, "tf-acc-test-") {
				log.Printf("[INFO] Skipping EC2 EBS Volume: %s", id)
				continue
			}

			input := &ec2.DeleteVolumeInput{
				VolumeId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 EBS Volume: %s", id)
			_, err := conn.DeleteVolume(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting EC2 EBS Volume (%s): %s", id, err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 EBS Volume sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving EC2 EBS Volumes: %s", err)
	}

	return nil
}

func TestAccAWSEBSVolume_basic(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`volume/vol-.+`)),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "iops"),
					resource.TestCheckNoResourceAttr(resourceName, "kms_key_id"),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
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

func TestAccAWSEBSVolume_updateAttachedEbsVolume(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsAttachedVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsAttachedVolumeConfigUpdateSize,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "20"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_updateSize(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigUpdateSize,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "size", "10"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_updateType(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "gp2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigUpdateType,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "type", "sc1"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_updateIops(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigWithIops,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEbsVolumeConfigWithIopsUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "iops", "200"),
				),
			},
		},
	})
}

func TestAccAWSEBSVolume_kmsKey(t *testing.T) {
	var v ec2.Volume
	ri := acctest.RandInt()
	config := fmt.Sprintf(testAccAwsEbsVolumeConfigWithKmsKey, ri)
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccAWSEBSVolume_NoIops(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigWithNoIops,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"iops"},
			},
		},
	})
}

func TestAccAWSEBSVolume_withTags(t *testing.T) {
	var v ec2.Volume
	resourceName := "aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsVolumeConfigWithTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "TerraformTest"),
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

func testAccCheckVolumeExists(n string, v *ec2.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		request := &ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeVolumes(request)
		if err == nil {
			if response.Volumes != nil && len(response.Volumes) > 0 {
				*v = *response.Volumes[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding EC2 volume %s", rs.Primary.ID)
	}
}

const testAccAwsEbsVolumeConfig = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "gp2"
  size = 1
}
`

const testAccAwsEbsAttachedVolumeConfig = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  owners = ["amazon"]
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.debian_jessie_latest.id}"
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name    = "test-terraform"
  }
}

data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  depends_on = ["aws_instance.test"]
  availability_zone = "${aws_instance.test.availability_zone}"
  type = "gp2"
  size = "10"
}

resource "aws_volume_attachment" "test" {
  depends_on  = ["aws_ebs_volume.test"]
  device_name = "/dev/xvdg"
  volume_id   = "${aws_ebs_volume.test.id}"
  instance_id = "${aws_instance.test.id}"
}
`

const testAccAwsEbsAttachedVolumeConfigUpdateSize = `
data "aws_ami" "debian_jessie_latest" {
  most_recent = true

  filter {
    name   = "name"
    values = ["amzn-ami-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  owners = ["amazon"]
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.debian_jessie_latest.id}"
  instance_type = "t2.medium"

  root_block_device {
    volume_size           = "10"
    volume_type           = "standard"
    delete_on_termination = true
  }

  tags = {
    Name    = "test-terraform"
  }
}

resource "aws_ebs_volume" "test" {
  depends_on = ["aws_instance.test"]
  availability_zone = "${aws_instance.test.availability_zone}"
  type = "gp2"
  size = "20"
}

resource "aws_volume_attachment" "test" {
  depends_on  = ["aws_ebs_volume.test"]
  device_name = "/dev/xvdg"
  volume_id   = "${aws_ebs_volume.test.id}"
  instance_id = "${aws_instance.test.id}"
}
`

const testAccAwsEbsVolumeConfigUpdateSize = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "gp2"
  size = 10
  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigUpdateType = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "sc1"
  size = 500
  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithIops = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "io1"
  size = 4
  iops = 100
  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithIopsUpdated = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  type = "io1"
  size = 4
  iops = 200
  tags = {
    Name = "tf-acc-test-ebs-volume-test"
  }
}
`

const testAccAwsEbsVolumeConfigWithKmsKey = `
resource "aws_kms_key" "test" {
  description = "Terraform acc test %d"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1
  encrypted         = true
  kms_key_id        = "${aws_kms_key.test.arn}"
}
`

const testAccAwsEbsVolumeConfigWithTags = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size = 1
  tags = {
    Name = "TerraformTest"
  }
}
`

const testAccAwsEbsVolumeConfigWithNoIops = `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size = 10
  type = "gp2"
  iops = 0
  tags = {
    Name = "TerraformTest"
  }
}
`
