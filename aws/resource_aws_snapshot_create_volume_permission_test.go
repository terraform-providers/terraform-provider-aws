package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSnapshotCreateVolumePermission_Basic(t *testing.T) {
	var snapshotId string
	accountId := "111122223333"

	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Scaffold everything
			{
				Config: testAccAWSSnapshotCreateVolumePermissionConfig(true, accountId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceGetAttr("aws_ebs_snapshot.example_snapshot", "id", &snapshotId),
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

func testAccAWSSnapshotCreateVolumePermissionExists(accountId, snapshotId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		if has, err := hasCreateVolumePermission(conn, *snapshotId, *accountId); err != nil {
			return err
		} else if !has {
			return fmt.Errorf("create volume permission does not exist for '%s' on '%s'", *accountId, *snapshotId)
		}
		return nil
	}
}

func testAccAWSSnapshotCreateVolumePermissionDestroyed(accountId, snapshotId *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		if has, err := hasCreateVolumePermission(conn, *snapshotId, *accountId); err != nil {
			return err
		} else if has {
			return fmt.Errorf("create volume permission still exists for '%s' on '%s'", *accountId, *snapshotId)
		}
		return nil
	}
}

func testAccAWSSnapshotCreateVolumePermissionConfig(includeCreateVolumePermission bool, accountID string) string {
	base := `
data "aws_availability_zones" "available" {}

resource "aws_ebs_volume" "example" {
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  size              = 1

  tags = {
    Name = "ebs_snap_perm"
  }
}

resource "aws_ebs_snapshot" "example_snapshot" {
  volume_id = "${aws_ebs_volume.example.id}"
}
`

	if !includeCreateVolumePermission {
		return base
	}

	return base + fmt.Sprintf(`
resource "aws_snapshot_create_volume_permission" "self-test" {
  snapshot_id = "${aws_ebs_snapshot.example_snapshot.id}"
  account_id  = %q
}
`, accountID)
}
