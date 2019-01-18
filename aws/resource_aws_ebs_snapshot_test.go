package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEBSSnapshot_basic(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-basic-%s", acctest.RandString(7))

	deleteSnapshot := func() {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []*string{aws.String(rName)},
				},
			},
		})
		if err != nil {
			t.Fatalf("Error fetching snapshot: %s", err)
		}
		if len(resp.Snapshots) == 0 {
			t.Fatalf("No snapshot exists with tag:Name = %s", rName)
		}
		snapshotId := resp.Snapshots[0].SnapshotId
		_, err = conn.DeleteSnapshot(&ec2.DeleteSnapshotInput{SnapshotId: snapshotId})
		if err != nil {
			t.Fatalf("Error deleting snapshot %s with tag:Name = %s: %s", *snapshotId, rName, err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEbsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists("aws_ebs_snapshot.test", &v),
					testAccCheckTags(&v.Tags, "Name", rName),
				),
			},
			{
				PreConfig: deleteSnapshot,
				Config:    testAccAwsEbsSnapshotConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists("aws_ebs_snapshot.test", &v),
					testAccCheckTags(&v.Tags, "Name", rName),
				),
			},
		},
	})
}

func TestAccAWSEBSSnapshot_withDescription(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-desc-%s", acctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEbsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotConfigWithDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists("aws_ebs_snapshot.test", &v),
					resource.TestCheckResourceAttr("aws_ebs_snapshot.test", "description", rName),
				),
			},
		},
	})
}

func TestAccAWSEBSSnapshot_withKms(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-kms-%s", acctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEbsSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotConfigWithKms(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists("aws_ebs_snapshot.test", &v),
					resource.TestMatchResourceAttr("aws_ebs_snapshot.test", "kms_key_id",
						regexp.MustCompile(`^arn:aws:kms:[a-z]{2}-[a-z]+-\d{1}:[0-9]{12}:key/[a-z0-9-]{36}$`)),
				),
			},
		},
	})
}

func testAccCheckSnapshotExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		request := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeSnapshots(request)
		if err == nil {
			if response.Snapshots != nil && len(response.Snapshots) > 0 {
				*v = *response.Snapshots[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
	}
}

func testAccCheckAWSEbsSnapshotDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_snapshot" {
			continue
		}
		input := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSnapshots(input)
		if err != nil {
			if isAWSErr(err, "InvalidSnapshot.NotFound", "") {
				continue
			}
			return err
		}
		if output != nil && len(output.Snapshots) > 0 && aws.StringValue(output.Snapshots[0].SnapshotId) == rs.Primary.ID {
			return fmt.Errorf("EBS Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAwsEbsSnapshotConfigBasic(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_region.current.name}a"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "%s"
  }
}
`, rName)
}

func testAccAwsEbsSnapshotConfigWithDescription(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ebs_volume" "description_test" {
  availability_zone = "${data.aws_region.current.name}a"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.description_test.id}"
  description = "%s"
}
`, rName)
}

func testAccAwsEbsSnapshotConfigWithKms(rName string) string {
	return fmt.Sprintf(`
variable "name" { default = "%s" }
data "aws_region" "current" {}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags = {
    Name = "${var.name}"
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_region.current.name}a"
  size              = 1
  encrypted         = true
  kms_key_id        = "${aws_kms_key.test.arn}"

  tags = {
    Name = "${var.name}"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "${var.name}"
  }
}
`, rName)
}
