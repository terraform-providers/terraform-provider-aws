package aws

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSStorageGatewayFsxAssociateFileSystem_basic(t *testing.T) {
	var fsxFileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_fsx_associate_file_system.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	fsxResourceName := "aws_fsx_windows_file_system.test"
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(storagegateway.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, storagegateway.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsStorageGatewayFsxAssociateFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsStorageGatewayFsxAssociateFileSystemConfig_Required(rName, username),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsStorageGatewayFsxAssociateFileSystemExists(resourceName, &fsxFileSystemAssociation),

					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", fsxResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "username", username),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
				Destroy: false,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSStorageGatewayFsxAssociateFileSystem_tags1(t *testing.T) {
}

func TestAccAWSStorageGatewayFsxAssociateFileSystem_tags2(t *testing.T) {
}

func TestAccAWSStorageGatewayFsxAssociateFileSystem_cacheAttributes(t *testing.T) {
}

func TestAccAWSStorageGatewayFsxAssociateFileSystem_disappears(t *testing.T) {
}

func testAccCheckAwsStorageGatewayFsxAssociateFileSystemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_fsx_associate_file_system" {
			continue
		}

		input := &storagegateway.DescribeFileSystemAssociationsInput{
			FileSystemAssociationARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeFileSystemAssociations(input)

		if err != nil {
			if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified fsx file share was not found.") {
				continue
			}
			return err
		}

		if output != nil && len(output.FileSystemAssociationInfoList) > 0 && output.FileSystemAssociationInfoList[0] != nil {
			return fmt.Errorf("Storage Gateway Fsx File System %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsStorageGatewayFsxAssociateFileSystemExists(resourceName string, fsxFileSystemAssociation *storagegateway.FileSystemAssociationInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn
		input := &storagegateway.DescribeFileSystemAssociationsInput{
			FileSystemAssociationARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeFileSystemAssociations(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.FileSystemAssociationInfoList) == 0 || output.FileSystemAssociationInfoList[0] == nil {
			return fmt.Errorf("Storage Gateway Fsx File System %q does not exist", rs.Primary.ID)
		}

		*fsxFileSystemAssociation = *output.FileSystemAssociationInfoList[0]

		return nil
	}
}

func testAccAWSStorageGatewayFsxAssociateFileSystemBase(rName, username string) string {
	return composeConfig(
		testAccAWSStorageGatewayGatewayConfigSmbActiveDirectorySettingsBase(rName),
		testAccAWSStorageGatewayGatewayConfig_DirectoryServiceMicrosoftAD(rName),
		fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_FSX_SMB"

  smb_active_directory_settings {
    domain_name        = aws_directory_service_directory.test.name
    password           = aws_directory_service_directory.test.password
    username           = %[2]q
  }

  tags = {
    Name = %[1]q
  }
}

`, rName, username))
}

func testAccAwsStorageGatewayFsxAssociateFileSystemConfig_Required(rName, username string) string {
	tempVar := testAccAWSStorageGatewayFsxAssociateFileSystemBase(rName, username) + fmt.Sprintf(`
resource "aws_storagegateway_fsx_associate_file_system" "test" {
  gateway_arn = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username = %[1]q
  password = aws_directory_service_directory.test.password
}
`, username)

	data := []byte(tempVar)
	err := ioutil.WriteFile("/tmp/tf_resources.tf", data, 0644)
	if err != nil {
		return ""
	}
	return tempVar
}
