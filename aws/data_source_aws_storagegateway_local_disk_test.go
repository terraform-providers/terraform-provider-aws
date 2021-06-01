package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSStorageGatewayLocalDiskDataSource_DiskNode(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode_NonExistent(rName),
				ExpectError: regexp.MustCompile(`no results found`),
			},
			{
				Config: testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSStorageGatewayLocalDiskDataSourceExists(dataSourceName),
					resource.TestCheckResourceAttrSet(dataSourceName, "disk_id"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayLocalDiskDataSource_DiskPath(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_storagegateway_local_disk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath_NonExistent(rName),
				ExpectError: regexp.MustCompile(`no results found`),
			},
			{
				Config: testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSStorageGatewayLocalDiskDataSourceExists(dataSourceName),
					resource.TestCheckResourceAttrSet(dataSourceName, "disk_id"),
				),
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

func testAccAWSStorageGatewayLocalDiskDataSourceConfigBase(rName string) string {
	return composeConfig(
		testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName),
		fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = aws_instance.test.availability_zone
  size              = "10"
  type              = "gp2"

  tags = {
    Name = %[1]q
  }
}

resource "aws_volume_attachment" "test" {
  device_name  = "/dev/xvdb"
  force_detach = true
  instance_id  = aws_instance.test.id
  volume_id    = aws_ebs_volume.test.id
}
`, rName))
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode(rName string) string {
	return composeConfig(
		testAccAWSStorageGatewayLocalDiskDataSourceConfigBase(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_node   = aws_volume_attachment.test.device_name
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskNode_NonExistent(rName string) string {
	return composeConfig(
		testAccAWSStorageGatewayLocalDiskDataSourceConfigBase(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_node   = replace(aws_volume_attachment.test.device_name, "xvdb", "nonexistent")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath(rName string) string {
	return composeConfig(
		testAccAWSStorageGatewayLocalDiskDataSourceConfigBase(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_path   = split(".", aws_instance.test.instance_type)[0] == "m4" ? aws_volume_attachment.test.device_name : replace(aws_volume_attachment.test.device_name, "xvdb", "nvme1n1")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}

func testAccAWSStorageGatewayLocalDiskDataSourceConfig_DiskPath_NonExistent(rName string) string {
	return composeConfig(
		testAccAWSStorageGatewayLocalDiskDataSourceConfigBase(rName),
		`
data "aws_storagegateway_local_disk" "test" {
  disk_path   = replace(aws_volume_attachment.test.device_name, "xvdb", "nonexistent")
  gateway_arn = aws_storagegateway_gateway.test.arn
}
`)
}
