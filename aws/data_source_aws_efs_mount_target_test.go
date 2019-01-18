package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsEfsMountTargetByMountTargetId(t *testing.T) {
	rName := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEfsMountTargetConfigByMountTargetId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_efs_mount_target.by_mount_target_id", "file_system_arn", "aws_efs_mount_target.alpha", "file_system_arn"),
					resource.TestCheckResourceAttrSet("data.aws_efs_mount_target.by_mount_target_id", "file_system_id"),
					resource.TestCheckResourceAttrSet("data.aws_efs_mount_target.by_mount_target_id", "ip_address"),
					resource.TestCheckResourceAttrSet("data.aws_efs_mount_target.by_mount_target_id", "subnet_id"),
					resource.TestCheckResourceAttrSet("data.aws_efs_mount_target.by_mount_target_id", "network_interface_id"),
					resource.TestCheckResourceAttrSet("data.aws_efs_mount_target.by_mount_target_id", "dns_name"),
				),
			},
		},
	})
}

func testAccAwsEfsMountTargetConfigByMountTargetId(ct string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "foo" {
	creation_token = "%s"
}

resource "aws_efs_mount_target" "alpha" {
	file_system_id = "${aws_efs_file_system.foo.id}"
	subnet_id = "${aws_subnet.alpha.id}"
}

resource "aws_vpc" "foo" {
	cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-efs-mount-target"
	}
}

resource "aws_subnet" "alpha" {
	vpc_id = "${aws_vpc.foo.id}"
	availability_zone = "us-west-2a"
	cidr_block = "10.0.1.0/24"
	tags = {
		Name = "tf-acc-efs-mount-target"
	}
}

data "aws_efs_mount_target" "by_mount_target_id" {
	mount_target_id = "${aws_efs_mount_target.alpha.id}"
}
`, ct)
}
