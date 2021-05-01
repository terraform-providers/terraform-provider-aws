package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsEfsMountTarget_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, efs.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEfsMountTargetConfigByMountTargetId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups", resourceName, "security_groups"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEfsMountTarget_byAccessPointID(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, efs.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSMountTargetConfigByAccessPointId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups", resourceName, "security_groups"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEfsMountTarget_byFileSystemID(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_efs_mount_target.test"
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, efs.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSMountTargetConfigByFileSystemId(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_arn", resourceName, "file_system_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "file_system_id", resourceName, "file_system_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_id", resourceName, "subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interface_id", resourceName, "network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "mount_target_dns_name", resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_name", resourceName, "availability_zone_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups", resourceName, "security_groups"),
				),
			},
		},
	})
}

func testAccAwsEfsMountTargetDataSourceBaseConfig(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAwsEfsMountTargetConfigByMountTargetId(rName string) string {
	return composeConfig(testAccAwsEfsMountTargetDataSourceBaseConfig(rName), `
data "aws_efs_mount_target" "test" {
  mount_target_id = aws_efs_mount_target.test.id
}
`)
}

func testAccAWSEFSMountTargetConfigByAccessPointId(rName string) string {
	return composeConfig(testAccAwsEfsMountTargetDataSourceBaseConfig(rName), `
resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}

data "aws_efs_mount_target" "test" {
  access_point_id = aws_efs_access_point.test.id
}
`)
}

func testAccAWSEFSMountTargetConfigByFileSystemId(rName string) string {
	return composeConfig(testAccAwsEfsMountTargetDataSourceBaseConfig(rName), `
data "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id

  depends_on = [aws_efs_mount_target.test]
}
`)
}
