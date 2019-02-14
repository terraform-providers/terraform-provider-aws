package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSStorageGatewayLocalDiskDataSource_DiskNode(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSStorageGatewayLocalDiskDataSourceExists(dataSourceName),
					resource.TestCheckResourceAttrSet(dataSourceName, "disk_id"),
				),
			},
			{
				Config:      testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode_NonExistent(rName),
				ExpectError: regexp.MustCompile(`no results found`),
			},
		},
	})
}

func TestAccAWSStorageGatewayLocalDiskDataSource_DiskPath(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSStorageGatewayLocalDiskDataSourceExists(dataSourceName),
					resource.TestCheckResourceAttrSet(dataSourceName, "disk_id"),
				),
			},
			{
				Config:      testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath_NonExistent(rName),
				ExpectError: regexp.MustCompile(`no results found`),
			},
		},
	})
}

func testAccAWSStorageGatewayLocalDiskDataSourceExists(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("not found: %s", dataSourceName)
		}

		return nil
	}
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/sdb"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.test.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_node   = "${aws_volume_attachment.test.device_name}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}
`, rName)
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode_NonExistent(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/sdb"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.test.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_node   = "/dev/sdz"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}
`, rName)
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdb"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.test.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_path   = "${aws_volume_attachment.test.device_name}"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}
`, rName)
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath_NonExistent(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName) + fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = "${aws_instance.test.availability_zone}"
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdb"
  force_detach = true
  instance_id  = "${aws_instance.test.id}"
  volume_id    = "${aws_ebs_volume.test.id}"
}

data "aws_storagegateway_local_disk" "test" {
  disk_path   = "/dev/xvdz"
  gateway_arn = "${aws_storagegateway_gateway.test.arn}"
}
`, rName)
}
