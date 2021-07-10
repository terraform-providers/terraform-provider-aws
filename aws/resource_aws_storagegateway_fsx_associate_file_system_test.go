package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"username", "password"},
			},
		},
	})
}

func TestAccAWSStorageGatewayFsxAssociateFileSystem_tags(t *testing.T) {
	var fsxFileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_fsx_associate_file_system.test"
	username := "Admin"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(storagegateway.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, storagegateway.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsStorageGatewayFsxAssociateFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsStorageGatewayFsxAssociateFileSystemConfigTags1(rName, username, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsStorageGatewayFsxAssociateFileSystemExists(resourceName, &fsxFileSystemAssociation),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"username", "password"},
			},
			{
				Config: testAccAwsStorageGatewayFsxAssociateFileSystemConfigTags2(rName, username, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsStorageGatewayFsxAssociateFileSystemExists(resourceName, &fsxFileSystemAssociation),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsStorageGatewayFsxAssociateFileSystemConfigTags1(rName, username, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsStorageGatewayFsxAssociateFileSystemExists(resourceName, &fsxFileSystemAssociation),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`fs-association/fsa-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayFsxAssociateFileSystem_cacheAttributes(t *testing.T) {
}

func TestAccAWSStorageGatewayFsxAssociateFileSystem_disappears(t *testing.T) {
	var fsxFileSystemAssociation storagegateway.FileSystemAssociationInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_fsx_associate_file_system.test"
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
					testAccCheckResourceDisappears(testAccProvider, resourceAwsStorageGatewayFsxAssociateFileSystem(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
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
			if tfawserr.ErrCodeEquals(err, storagegateway.ErrCodeInvalidGatewayRequestException) {
				var igrex *storagegateway.InvalidGatewayRequestException
				if ok := errors.As(err, &igrex); ok {
					if err := igrex.Error_; err != nil {
						if aws.StringValue(err.ErrorCode) == "FileSystemAssociationNotFound" {
							continue
						}
					}
				}
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
	return testAccAWSStorageGatewayFsxAssociateFileSystemBase(rName, username) + fmt.Sprintf(`
resource "aws_storagegateway_fsx_associate_file_system" "test" {
  gateway_arn = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username = %[1]q
  password = aws_directory_service_directory.test.password
}
`, username)
}

func testAccAwsStorageGatewayFsxAssociateFileSystemConfigTags1(rName, username, tagKey1, tagValue1 string) string {
	return testAccAWSStorageGatewayFsxAssociateFileSystemBase(rName, username) + fmt.Sprintf(`
resource "aws_storagegateway_fsx_associate_file_system" "test" {
  gateway_arn = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username = %[1]q
  password = aws_directory_service_directory.test.password

  tags = {
    %[2]q = %[3]q
  }
}
`, username, tagKey1, tagValue1)
}

func testAccAwsStorageGatewayFsxAssociateFileSystemConfigTags2(rName, username, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSStorageGatewayFsxAssociateFileSystemBase(rName, username) + fmt.Sprintf(`
resource "aws_storagegateway_fsx_associate_file_system" "test" {
  gateway_arn = aws_storagegateway_gateway.test.arn
  location_arn = aws_fsx_windows_file_system.test.arn
  username = %[1]q
  password = aws_directory_service_directory.test.password

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, username, tagKey1, tagValue1, tagKey2, tagValue2)
}
