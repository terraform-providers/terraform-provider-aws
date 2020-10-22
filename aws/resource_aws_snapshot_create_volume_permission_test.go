package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSnapshotCreateVolumePermission_basic(t *testing.T) {
	var snapshotId string
	accountId := "111122223333"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSSnapshotCreateVolumePermissionDestroy,
		Steps: []resource.TestStep{
			// Scaffold everything
			{
				Config: testAccAWSSnapshotCreateVolumePermissionConfig(true, accountId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceGetAttr("aws_ebs_snapshot.test", "id", &snapshotId),
					testAccAWSSnapshotCreateVolumePermissionExists(&accountId, &snapshotId),
				),
			},
			// Drop just create volume permission to test destruction
			{
				Config: testAccAWSSnapshotCreateVolumePermissionConfig(false, accountId),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSSnapshotCreateVolumePermissionDestroyed(&accountId, &snapshotId),
				),
			},
		},
	})
}

func TestAccAWSSnapshotCreateVolumePermission_disappears(t *testing.T) {
	var snapshotId string
	accountId := "111122223333"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAWSSnapshotCreateVolumePermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSnapshotCreateVolumePermissionConfig(true, accountId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceGetAttr("aws_ebs_snapshot.test", "id", &snapshotId),
					testAccAWSSnapshotCreateVolumePermissionExists(&accountId, &snapshotId),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSnapshotCreateVolumePermission(), "aws_snapshot_create_volume_permission.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSSnapshotCreateVolumePermissionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_snapshot_create_volume_permission" {
			continue
		}

		snapshotID, accountID, err := resourceAwsSnapshotCreateVolumePermissionParseID(rs.Primary.ID)
		if err != nil {
			return err
		}
		if has, err := hasCreateVolumePermission(conn, snapshotID, accountID); err != nil {
			return err
		} else if has {
			return fmt.Errorf("create volume permission still exist for '%s' on '%s'", accountID, snapshotID)
		}
	}

	return nil
}

func testAccAWSSnapshotCreateVolumePermissionExists(accountId, snapshotId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		if has, err := hasCreateVolumePermission(conn, aws.StringValue(snapshotId), aws.StringValue(accountId)); err != nil {
			return err
		} else if !has {
			return fmt.Errorf("create volume permission does not exist for '%s' on '%s'", aws.StringValue(snapshotId), aws.StringValue(accountId))
		}
		return nil
	}
}

func testAccAWSSnapshotCreateVolumePermissionDestroyed(accountId, snapshotId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		if has, err := hasCreateVolumePermission(conn, aws.StringValue(snapshotId), aws.StringValue(accountId)); err != nil {
			return err
		} else if has {
			return fmt.Errorf("create volume permission still exists for '%s' on '%s'", aws.StringValue(snapshotId), aws.StringValue(accountId))
		}
		return nil
	}
}

func testAccAWSSnapshotCreateVolumePermissionConfig(includeCreateVolumePermission bool, accountID string) string {
	base := `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = "ebs_snap_perm"
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}
`

	if !includeCreateVolumePermission {
		return base
	}

	return base + fmt.Sprintf(`
resource "aws_snapshot_create_volume_permission" "test" {
  snapshot_id = aws_ebs_snapshot.test.id
  account_id  = %q
}
`, accountID)
}
