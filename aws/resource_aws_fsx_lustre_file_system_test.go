package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_fsx_lustre_file_system", &resource.Sweeper{
		Name: "aws_fsx_lustre_file_system",
		F:    testSweepFSXLustreFileSystems,
	})
}

func testSweepFSXLustreFileSystems(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).fsxconn
	input := &fsx.DescribeFileSystemsInput{}

	err = conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemType) != fsx.FileSystemTypeLustre {
				continue
			}

			input := &fsx.DeleteFileSystemInput{
				FileSystemId: fs.FileSystemId,
			}

			log.Printf("[INFO] Deleting FSx lustre filesystem: %s", aws.StringValue(fs.FileSystemId))
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
		log.Printf("[WARN] Skipping FSx Lustre Filesystem sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing FSx Lustre Filesystems: %s", err)
	}

	return nil

}

func TestAccAWSFsxLustreFileSystem_basic(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigSubnetIds1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "fsx", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, "dns_name", regexp.MustCompile(`fs-.+\.fsx\.`)),
					resource.TestCheckResourceAttr(resourceName, "export_path", ""),
					resource.TestCheckResourceAttr(resourceName, "import_path", ""),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "0"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_ids.#", "2"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "vpc_id", regexp.MustCompile(`^vpc-.+`)),
					resource.TestMatchResourceAttr(resourceName, "weekly_maintenance_start_time", regexp.MustCompile(`^\d:\d\d:\d\d$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
		},
	})
}

func TestAccAWSFsxLustreFileSystem_disappears(t *testing.T) {
	var filesystem fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigSubnetIds1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem),
					testAccCheckFsxLustreFileSystemDisappears(&filesystem),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSFsxLustreFileSystem_ExportPath(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigExportPath(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "export_path", fmt.Sprintf("s3://%s", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigExportPath(rName, "/prefix/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "export_path", fmt.Sprintf("s3://%s/prefix/", rName)),
				),
			},
		},
	})
}

// lintignore: AT002
func TestAccAWSFsxLustreFileSystem_ImportPath(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigImportPath(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "import_path", fmt.Sprintf("s3://%s", rName)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigImportPath(rName, "/prefix/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "import_path", fmt.Sprintf("s3://%s/prefix/", rName)),
				),
			},
		},
	})
}

// lintignore: AT002
func TestAccAWSFsxLustreFileSystem_ImportedFileChunkSize(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigImportedFileChunkSize(rName, 2048),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "2048"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigImportedFileChunkSize(rName, 4096),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "imported_file_chunk_size", "4096"),
				),
			},
		},
	})
}

func TestAccAWSFsxLustreFileSystem_SecurityGroupIds(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigSecurityGroupIds1(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigSecurityGroupIds2(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSFsxLustreFileSystem_StorageCapacity(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigStorageCapacity(7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "7200"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigStorageCapacity(1200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxLustreFileSystemRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "storage_capacity", "1200"),
				),
			},
		},
	})
}

func TestAccAWSFsxLustreFileSystem_Tags(t *testing.T) {
	var filesystem1, filesystem2, filesystem3 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem3),
					testAccCheckFsxLustreFileSystemNotRecreated(&filesystem2, &filesystem3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSFsxLustreFileSystem_WeeklyMaintenanceStartTime(t *testing.T) {
	var filesystem1, filesystem2 fsx.FileSystem
	resourceName := "aws_fsx_lustre_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFsxLustreFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsFsxLustreFileSystemConfigWeeklyMaintenanceStartTime("1:01:01"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem1),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "1:01:01"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"security_group_ids"},
			},
			{
				Config: testAccAwsFsxLustreFileSystemConfigWeeklyMaintenanceStartTime("2:02:02"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFsxLustreFileSystemExists(resourceName, &filesystem2),
					testAccCheckFsxLustreFileSystemNotRecreated(&filesystem1, &filesystem2),
					resource.TestCheckResourceAttr(resourceName, "weekly_maintenance_start_time", "2:02:02"),
				),
			},
		},
	})
}

func testAccCheckFsxLustreFileSystemExists(resourceName string, fs *fsx.FileSystem) resource.TestCheckFunc {
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

func testAccCheckFsxLustreFileSystemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).fsxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_fsx_lustre_file_system" {
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

func testAccCheckFsxLustreFileSystemDisappears(filesystem *fsx.FileSystem) resource.TestCheckFunc {
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

func testAccCheckFsxLustreFileSystemNotRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) != aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccCheckFsxLustreFileSystemRecreated(i, j *fsx.FileSystem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.FileSystemId) == aws.StringValue(j.FileSystemId) {
			return fmt.Errorf("FSx File System (%s) not recreated", aws.StringValue(i.FileSystemId))
		}

		return nil
	}
}

func testAccAwsFsxLustreFileSystemConfigBase() string {
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
`)
}

func testAccAwsFsxLustreFileSystemConfigExportPath(rName, exportPrefix string) string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  acl    = "private"
  bucket = %[1]q
}

resource "aws_fsx_lustre_file_system" "test" {
  export_path      = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  import_path      = "s3://${aws_s3_bucket.test.bucket}"
  storage_capacity = 1200
  subnet_ids       = ["${aws_subnet.test1.id}"]
}
`, rName, exportPrefix)
}

func testAccAwsFsxLustreFileSystemConfigImportPath(rName, importPrefix string) string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  acl    = "private"
  bucket = %[1]q
}

resource "aws_fsx_lustre_file_system" "test" {
  import_path      = "s3://${aws_s3_bucket.test.bucket}%[2]s"
  storage_capacity = 1200
  subnet_ids       = ["${aws_subnet.test1.id}"]
}
`, rName, importPrefix)
}

func testAccAwsFsxLustreFileSystemConfigImportedFileChunkSize(rName string, importedFileChunkSize int) string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  acl    = "private"
  bucket = %[1]q
}

resource "aws_fsx_lustre_file_system" "test" {
  import_path              = "s3://${aws_s3_bucket.test.bucket}"
  imported_file_chunk_size = %[2]d
  storage_capacity         = 1200
  subnet_ids               = ["${aws_subnet.test1.id}"]
}
`, rName, importedFileChunkSize)
}

func testAccAwsFsxLustreFileSystemConfigSecurityGroupIds1() string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
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

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = ["${aws_security_group.test1.id}"]
  storage_capacity   = 1200
  subnet_ids         = ["${aws_subnet.test1.id}"]
}
`)
}

func testAccAwsFsxLustreFileSystemConfigSecurityGroupIds2() string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
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

resource "aws_fsx_lustre_file_system" "test" {
  security_group_ids = ["${aws_security_group.test1.id}", "${aws_security_group.test2.id}"]
  storage_capacity   = 1200
  subnet_ids         = ["${aws_subnet.test1.id}"]
}
`)
}

func testAccAwsFsxLustreFileSystemConfigStorageCapacity(storageCapacity int) string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = %[1]d
  subnet_ids       = ["${aws_subnet.test1.id}"]
}
`, storageCapacity)
}

func testAccAwsFsxLustreFileSystemConfigSubnetIds1() string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = ["${aws_subnet.test1.id}"]
}
`)
}

func testAccAwsFsxLustreFileSystemConfigTags1(tagKey1, tagValue1 string) string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = ["${aws_subnet.test1.id}"]

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccAwsFsxLustreFileSystemConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = ["${aws_subnet.test1.id}"]

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAwsFsxLustreFileSystemConfigWeeklyMaintenanceStartTime(weeklyMaintenanceStartTime string) string {
	return testAccAwsFsxLustreFileSystemConfigBase() + fmt.Sprintf(`
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity              = 1200
  subnet_ids                    = ["${aws_subnet.test1.id}"]
  weekly_maintenance_start_time = %[1]q
}
`, weeklyMaintenanceStartTime)
}
