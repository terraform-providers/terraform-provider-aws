package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSVolumeAttachment_basic(t *testing.T) {
	var i ec2.Instance
	var v ec2.Volume

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_volume_attachment.ebs_att", "device_name", "/dev/sdh"),
					testAccCheckInstanceExists(
						"aws_instance.web", &i),
					testAccCheckVolumeExists(
						"aws_ebs_volume.example", &v),
					testAccCheckVolumeAttachmentExists(
						"aws_volume_attachment.ebs_att", &i, &v),
				),
			},
		},
	})
}

func TestAccAWSVolumeAttachment_skipDestroy(t *testing.T) {
	var i ec2.Instance
	var v ec2.Volume

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfigSkipDestroy,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_volume_attachment.ebs_att", "device_name", "/dev/sdh"),
					testAccCheckInstanceExists(
						"aws_instance.web", &i),
					testAccCheckVolumeExists(
						"aws_ebs_volume.example", &v),
					testAccCheckVolumeAttachmentExists(
						"aws_volume_attachment.ebs_att", &i, &v),
				),
			},
		},
	})
}

func TestAccAWSVolumeAttachment_attachStopped(t *testing.T) {
	var i ec2.Instance
	var v ec2.Volume

	stopInstance := func() {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.StopInstances(&ec2.StopInstancesInput{
			InstanceIds: []*string{i.InstanceId},
		})
		if err != nil {
			t.Fatalf("error stopping instance (%s): %s", aws.StringValue(i.InstanceId), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"pending", "running", "stopping"},
			Target:     []string{"stopped"},
			Refresh:    InstanceStateRefreshFunc(conn, *i.InstanceId, []string{}),
			Timeout:    10 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			t.Fatalf("Error waiting for instance(%s) to stop: %s", *i.InstanceId, err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfigInstanceOnly,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(
						"aws_instance.web", &i),
				),
			},
			{
				PreConfig: stopInstance,
				Config:    testAccVolumeAttachmentConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_volume_attachment.ebs_att", "device_name", "/dev/sdh"),
					testAccCheckInstanceExists(
						"aws_instance.web", &i),
					testAccCheckVolumeExists(
						"aws_ebs_volume.example", &v),
					testAccCheckVolumeAttachmentExists(
						"aws_volume_attachment.ebs_att", &i, &v),
				),
			},
		},
	})
}

func TestAccAWSVolumeAttachment_update(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVolumeAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeAttachmentConfig_update(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_volume_attachment.ebs_att", "force_detach", "false"),
					resource.TestCheckResourceAttr(
						"aws_volume_attachment.ebs_att", "skip_destroy", "false"),
				),
			},
			{
				Config: testAccVolumeAttachmentConfig_update(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_volume_attachment.ebs_att", "force_detach", "true"),
					resource.TestCheckResourceAttr(
						"aws_volume_attachment.ebs_att", "skip_destroy", "true"),
				),
			},
		},
	})
}

func testAccCheckVolumeAttachmentExists(n string, i *ec2.Instance, v *ec2.Volume) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		for _, b := range i.BlockDeviceMappings {
			if rs.Primary.Attributes["device_name"] == *b.DeviceName {
				if b.Ebs.VolumeId != nil && rs.Primary.Attributes["volume_id"] == *b.Ebs.VolumeId {
					// pass
					return nil
				}
			}
		}

		return fmt.Errorf("Error finding instance/volume")
	}
}

func testAccCheckVolumeAttachmentDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		log.Printf("\n\n----- This is never called")
		if rs.Type != "aws_volume_attachment" {
			continue
		}
	}
	return nil
}

const testAccVolumeAttachmentConfigInstanceOnly = `
resource "aws_instance" "web" {
  ami = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type = "t1.micro"
  tags = {
    Name = "HelloWorld"
  }
}
`

const testAccVolumeAttachmentConfig = `
resource "aws_instance" "web" {
  ami = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type = "t1.micro"
  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size = 1
}

resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sdh"
  volume_id = "${aws_ebs_volume.example.id}"
  instance_id = "${aws_instance.web.id}"
}
`

const testAccVolumeAttachmentConfigSkipDestroy = `
resource "aws_instance" "web" {
  ami = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type = "t1.micro"
  tags = {
    Name = "HelloWorld"
  }
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size = 1
  tags = {
    Name = "TestVolume"
  }
}

data "aws_ebs_volume" "ebs_volume" {
  filter {
    name = "size"
    values = ["${aws_ebs_volume.example.size}"]
  }
  filter {
    name = "availability-zone"
    values = ["${aws_ebs_volume.example.availability_zone}"]
  }
  filter {
    name = "tag:Name"
    values = ["TestVolume"]
  }
}

resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sdh"
  volume_id = "${data.aws_ebs_volume.ebs_volume.id}"
  instance_id = "${aws_instance.web.id}"
  skip_destroy = true
}
`

func testAccVolumeAttachmentConfig_update(detach bool) string {
	return fmt.Sprintf(`
resource "aws_instance" "web" {
  ami = "ami-21f78e11"
  availability_zone = "us-west-2a"
  instance_type = "t1.micro"
}

resource "aws_ebs_volume" "example" {
  availability_zone = "us-west-2a"
  size = 1
}

resource "aws_volume_attachment" "ebs_att" {
  device_name = "/dev/sdh"
  volume_id = "${aws_ebs_volume.example.id}"
  instance_id = "${aws_instance.web.id}"
  force_detach = %t
  skip_destroy = %t
}
`, detach, detach)
}
