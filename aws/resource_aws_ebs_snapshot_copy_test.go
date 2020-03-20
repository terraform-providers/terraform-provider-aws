package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSEbsSnapshotCopy_basic(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "testAccAwsEbsSnapshotCopyConfig"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotCopy_tags(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAwsEbsSnapshotCopyConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsEbsSnapshotCopyConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotCopy_withDescription(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfigWithDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "description", "Copy Snapshot Acceptance Test"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotCopy_withRegions(t *testing.T) {
	var providers []*schema.Provider
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionsPreCheck(t)
			testAccAlternateRegionPreCheck(t)
		},
		ProviderFactories: testAccProviderFactories(&providers),
		CheckDestroy:      testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfigWithRegions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
				),
			},
		},
	})

}

func TestAccAWSEbsSnapshotCopy_withKms(t *testing.T) {
	var snapshot ec2.Snapshot
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfigWithKms,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotCopy_disappears(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEbsSnapshotCopyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists(resourceName, &snapshot),
					testAccCheckEbsSnapshotCopyDisappears(&snapshot),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEbsSnapshotCopyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_snapshot_copy" {
			continue
		}

		resp, err := conn.DescribeSnapshots(&ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		})

		if isAWSErr(err, "InvalidSnapshot.NotFound", "") {
			continue
		}

		if err == nil {
			for _, snapshot := range resp.Snapshots {
				if aws.StringValue(snapshot.SnapshotId) == rs.Primary.ID {
					return fmt.Errorf("EBS Snapshot still exists")
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckEbsSnapshotCopyDisappears(snapshot *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DeleteSnapshot(&ec2.DeleteSnapshotInput{
			SnapshotId: snapshot.SnapshotId,
		})

		return err
	}
}

func testAccCheckEbsSnapshotCopyExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSnapshots(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.Snapshots) == 0 {
			return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
		}

		*v = *output.Snapshots[0]

		return nil
	}
}

const testAccAwsEbsSnapshotCopyConfig = `
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
  source_region      = "${data.aws_region.current.name}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
  }
}
`

func testAccAwsEbsSnapshotCopyConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
  source_region      = "${data.aws_region.current.name}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
	"%s" = "%s"
  }
}
`, tagKey1, tagValue1)
}

func testAccAwsEbsSnapshotCopyConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
  source_region      = "${data.aws_region.current.name}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
	"%s" = "%s"
	"%s" = "%s"
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

const testAccAwsEbsSnapshotCopyConfigWithDescription = `
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithDescription"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id   = "${aws_ebs_volume.test.id}"
  description = "EBS Snapshot Acceptance Test"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithDescription"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  description        = "Copy Snapshot Acceptance Test"
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
  source_region      = "${data.aws_region.current.name}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithDescription"
  }
}
`

var testAccAwsEbsSnapshotCopyConfigWithRegions = testAccAlternateRegionProviderConfig() + `
data "aws_availability_zones" "alternate_available" {
  provider = "aws.alternate"
  state    = "available"
}

data "aws_region" "alternate" {
  provider = "aws.alternate"
}

resource "aws_ebs_volume" "test" {
  provider          = "aws.alternate"
  availability_zone = "${data.aws_availability_zones.alternate_available.names[0]}"
  size              = 1

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithRegions"
  }
}

resource "aws_ebs_snapshot" "test" {
  provider  = "aws.alternate"
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithRegions"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
  source_region      = "${data.aws_region.alternate.name}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithRegions"
  }
}
`

const testAccAwsEbsSnapshotCopyConfigWithKms = `
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_region" "current" {}

resource "aws_kms_key" "test" {
  description             = "testAccAwsEbsSnapshotCopyConfigWithKms"
  deletion_window_in_days = 7
}

resource "aws_ebs_volume" "test" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithKms"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithKms"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
  source_region      = "${data.aws_region.current.name}"
  encrypted          = true
  kms_key_id         = "${aws_kms_key.test.arn}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithKms"
  }
}
`
