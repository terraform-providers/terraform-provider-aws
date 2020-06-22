package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_fsx_windows_file_system", &resource.Sweeper{
		Name: "aws_fsx_windows_file_system",
		F:    testSweepFSXWindowsFileSystems,
	})
}

func testSweepFSXWindowsFileSystems(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).fsxconn
	input := &fsx.DescribeFileSystemsInput{}

	err = conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemType) != fsx.FileSystemTypeWindows {
				continue
			}

			input := &fsx.DeleteFileSystemInput{
				ClientRequestToken: aws.String(resource.UniqueId()),
				FileSystemId:       fs.FileSystemId,
				WindowsConfiguration: &fsx.DeleteFileSystemWindowsConfiguration{
					SkipFinalBackup: aws.Bool(true),
				},
			}

			log.Printf("[INFO] Deleting FSx windows filesystem: %s", aws.StringValue(fs.FileSystemId))
			_, err := conn.DeleteFileSystem(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting FSx filesystem: %s", err)
				continue
			}

			if err := waitForFsxFileSystemDeletion(conn, aws.StringValue(fs.FileSystemId), 30*time.Minute); err != nil {
				log.Printf("[ERROR] Error waiting for filesystem (%s) to delete: %s", aws.StringValue(fs.FileSystemId), err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping FSx Windows Filesystem sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing FSx Windows Filesystems: %s", err)
	}

	return nil

}

func TestAccAWSFsxWindowsFileSystem_basic(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigSubnetIds1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
					resource.TestMatchResourceAttr(resourceName, "daily_automatic_backup_start_time", regexp.MustCompile(`^\d\d:\d\d$`)),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(`fs-.+\..+`)),
					resource.TestMatchResourceAttr(resourceName, "kms_key_id", regexp.MustCompile(`^arn:`)),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "1"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "skip_final_backup", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "32"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "8"),
					resource.TestMatchResourceAttr(resourceName, "vpc_id", regexp.MustCompile(`^vpc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_disappears(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigSubnetIds1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem),
					testAccCheckFsxWindowsFileSystemDisappears(&filesystem),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_AutomaticBackupRetentionDays(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigAutomaticBackupRetentionDays(35),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "35"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigAutomaticBackupRetentionDays(0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "0"),
				),
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigAutomaticBackupRetentionDays(14),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem3),
					testAccCheckFsxWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "automatic_backup_retention_days", "14"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_CopyTagsToBackups(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigCopyTagsToBackups(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigCopyTagsToBackups(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "copy_tags_to_backups", "false"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_DailyAutomaticBackupStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigDailyAutomaticBackupStartTime("01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigDailyAutomaticBackupStartTime("02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "daily_automatic_backup_start_time", "02:02"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_KmsKeyId(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	kmsKeyResourceName1 := "aws_kms_key.test1"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigKmsKeyId1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName1, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigKmsKeyId2(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_SecurityGroupIds(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds2(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_SelfManagedActiveDirectory(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigSelfManagedActiveDirectory(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"self_managed_active_directory",
					"skip_final_backup",
				},
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_SelfManagedActiveDirectory_Username(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigSelfManagedActiveDirectoryUsername("Admin"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"self_managed_active_directory",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigSelfManagedActiveDirectoryUsername("Administrator"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem),
					resource.TestCheckResourceAttr(resourceName, "self_managed_active_directory.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_StorageCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigStorageCapacity(33),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "33"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigStorageCapacity(34),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "34"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_Tags(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem3),
					testAccCheckFsxWindowsFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_ThroughputCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigThroughputCapacity(16),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "16"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigThroughputCapacity(32),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "throughput_capacity", "32"),
				),
			},
		},
	})
}

func TestAccAWSFsxWindowsFileSystem_WeeklyMaintenanceStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_windows_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxWindowsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxWindowsFileSystemConfigWeeklyMaintenanceStartTime("1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"security_group_ids",
					"skip_final_backup",
				},
			},
			{
				Config: testAccAwsFsxWindowsFileSystemConfigWeeklyMaintenanceStartTime("2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxWindowsFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxWindowsFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func testAccCheckFsxWindowsFileSystemExists(resourceName string, fs *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).fsxconn

		filesystem, err := describeFsxFileSystem(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if filesystem == nil {
			return fmt.Errorf("FSx File System (%s) not found", rs.Primary.ID)
		}

		*fs = *filesystem

		return nil
	}
}

func testAccCheckFsxWindowsFileSystemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fsxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_windows_file_system" {
			continue
		}

		filesystem, err := describeFsxFileSystem(conn, rs.Primary.ID)

		if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
			continue
		}

		if err != nil {
			return err
		}

		if filesystem != nil {
			return fmt.Errorf("FSx File System (%s) still exists", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckFsxWindowsFileSystemDisappears(filesystem *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).fsxconn

		input := &fsx.DeleteFileSystemInput{
			FileSystemId: filesystem.FileSystemId,
		}

		_, err := conn.DeleteFileSystem(input)

		if err != nil {
			return err
		}

		return waitForFsxFileSystemDeletion(conn, aws.StringValue(filesystem.FileSystemId), 30*time.Minute)
	}
}

func testAccCheckFsxWindowsFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckFsxWindowsFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccAwsFsxWindowsFileSystemConfigBase() string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "test2" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = ["${aws_subnet.test1.id}", "${aws_subnet.test2.id}"]
    vpc_id     = "${aws_vpc.test.id}"
  }
}
`)
}

func testAccAwsFsxWindowsFileSystemConfigAutomaticBackupRetentionDays(automaticBackupRetentionDays int) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id             = "${aws_directory_service_directory.test.id}"
  automatic_backup_retention_days = %[1]d
  skip_final_backup               = true
  storage_capacity                = 32
  subnet_ids                      = ["${aws_subnet.test1.id}"]
  throughput_capacity             = 8
}
`, automaticBackupRetentionDays)
}

func testAccAwsFsxWindowsFileSystemConfigCopyTagsToBackups(copyTagsToBackups bool) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id  = "${aws_directory_service_directory.test.id}"
  copy_tags_to_backups = %[1]t
  skip_final_backup    = true
  storage_capacity     = 32
  subnet_ids           = ["${aws_subnet.test1.id}"]
  throughput_capacity  = 8
}
`, copyTagsToBackups)
}

func testAccAwsFsxWindowsFileSystemConfigDailyAutomaticBackupStartTime(dailyAutomaticBackupStartTime string) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id               = "${aws_directory_service_directory.test.id}"
  daily_automatic_backup_start_time = %[1]q
  skip_final_backup                 = true
  storage_capacity                  = 32
  subnet_ids                        = ["${aws_subnet.test1.id}"]
  throughput_capacity               = 8
}
`, dailyAutomaticBackupStartTime)
}

func testAccAwsFsxWindowsFileSystemConfigKmsKeyId1() string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_kms_key" "test1" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = "${aws_directory_service_directory.test.id}"
  kms_key_id          = "${aws_kms_key.test1.arn}"
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8
}
`)
}

func testAccAwsFsxWindowsFileSystemConfigKmsKeyId2() string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_kms_key" "test2" {
  description             = "FSx KMS Testing key"
  deletion_window_in_days = 7
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = "${aws_directory_service_directory.test.id}"
  kms_key_id          = "${aws_kms_key.test2.arn}"
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8
}
`)
}

func testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds1() string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    cidr_blocks = ["${aws_vpc.test.cidr_block}"]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id  = "${aws_directory_service_directory.test.id}"
  security_group_ids   = ["${aws_security_group.test1.id}"]
  skip_final_backup    = true
  storage_capacity     = 32
  subnet_ids           = ["${aws_subnet.test1.id}"]
  throughput_capacity  = 8
}
`)
}

func testAccAwsFsxWindowsFileSystemConfigSecurityGroupIds2() string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_security_group" "test1" {
  description = "security group for FSx testing"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    cidr_blocks = ["${aws_vpc.test.cidr_block}"]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_security_group" "test2" {
  description = "security group for FSx testing"
  vpc_id      = "${aws_vpc.test.id}"

  ingress {
    cidr_blocks = ["${aws_vpc.test.cidr_block}"]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id  = "${aws_directory_service_directory.test.id}"
  security_group_ids   = ["${aws_security_group.test1.id}", "${aws_security_group.test2.id}"]
  skip_final_backup    = true
  storage_capacity     = 32
  subnet_ids           = ["${aws_subnet.test1.id}"]
  throughput_capacity  = 8
}
`)
}

func testAccAwsFsxWindowsFileSystemConfigSelfManagedActiveDirectory() string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8

  self_managed_active_directory {
    dns_ips     = aws_directory_service_directory.test.dns_ip_addresses
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Admin"
  }
}
`)
}

func testAccAwsFsxWindowsFileSystemConfigSelfManagedActiveDirectoryUsername(username string) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8
  self_managed_active_directory {
    dns_ips     = aws_directory_service_directory.test.dns_ip_addresses
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = %[1]q
  }
}
`, username)
}

func testAccAwsFsxWindowsFileSystemConfigStorageCapacity(storageCapacity int) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = "${aws_directory_service_directory.test.id}"
  skip_final_backup   = true
  storage_capacity    = %[1]d
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8
}
`, storageCapacity)
}

func testAccAwsFsxWindowsFileSystemConfigSubnetIds1() string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = "${aws_directory_service_directory.test.id}"
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8
}
`)
}

func testAccAwsFsxWindowsFileSystemConfigTags1(tagKey1, tagValue1 string) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = "${aws_directory_service_directory.test.id}"
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAwsFsxWindowsFileSystemConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = "${aws_directory_service_directory.test.id}"
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = 8

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsFsxWindowsFileSystemConfigThroughputCapacity(throughputCapacity int) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = "${aws_directory_service_directory.test.id}"
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = ["${aws_subnet.test1.id}"]
  throughput_capacity = %[1]d
}
`, throughputCapacity)
}

func testAccAwsFsxWindowsFileSystemConfigWeeklyMaintenanceStartTime(weeklyMaintenanceStartTime string) string {
	return testAccAwsFsxWindowsFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_windows_file_system" "test" {
  active_directory_id           = "${aws_directory_service_directory.test.id}"
  skip_final_backup             = true
  storage_capacity              = 32
  subnet_ids                    = ["${aws_subnet.test1.id}"]
  throughput_capacity           = 8
  weekly_maintenance_start_time = %[1]q
}
`, weeklyMaintenanceStartTime)
}
