package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_efs_access_point", &resource.Sweeper{
		Name: "aws_efs_access_point",
		F:    testSweepEfsAccessPoints,
	})
}

func testSweepEfsAccessPoints(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).efsconn

	var errors error
	input := &efs.DescribeFileSystemsInput{}
	err = conn.DescribeFileSystemsPages(input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, filesystem := range page.FileSystems {
			id := aws.StringValue(filesystem.FileSystemId)
			log.Printf("[INFO] Deleting access points for EFS File System: %s", id)

			var errors error
			input := &efs.DescribeAccessPointsInput{
				FileSystemId: filesystem.FileSystemId,
			}
			for {
				out, err := conn.DescribeAccessPoints(input)
				if err != nil {
					errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS access points on File System %q: %w", id, err))
					break
				}

				if out == nil || len(out.AccessPoints) == 0 {
					log.Printf("[INFO] No EFS access points to sweep on File System %q", id)
					break
				}

				for _, AccessPoint := range out.AccessPoints {
					id := aws.StringValue(AccessPoint.AccessPointId)

					log.Printf("[INFO] Deleting EFS access point: %s", id)
					_, err := conn.DeleteAccessPoint(&efs.DeleteAccessPointInput{
						AccessPointId: AccessPoint.AccessPointId,
					})
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error deleting EFS access point %q: %w", id, err))
						continue
					}

					err = waitForDeleteEfsAccessPoint(conn, id, 10*time.Minute)
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error waiting for EFS access point %q to delete: %w", id, err))
						continue
					}
				}

				if out.NextToken == nil {
					break
				}
				input.NextToken = out.NextToken
			}
		}
		return true
	})
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS File Systems: %w", err))
	}

	return errors
}

func TestAccAWSEFSAccessPoint_basic(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_efs_access_point.test"
	fsResourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_arn", fsResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "file_system_id", fsResourceName, "id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticfilesystem", regexp.MustCompile(`access-point/fsap-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEFSAccessPoint_root_directory(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSAccessPointConfigRootDirectory(rName, "/home/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/home/test"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEFSAccessPoint_root_directory_creation_info(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSAccessPointConfigRootDirectoryCreationInfo(rName, "/home/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "root_directory.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.path", "/home/test"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.owner_gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.owner_uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "root_directory.0.creation_info.0.permissions", "755"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEFSAccessPoint_posix_user(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSAccessPointConfigPosixUser(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.secondary_gids.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEFSAccessPoint_posix_user_secondary_gids(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSAccessPointConfigPosixUserSecondaryGids(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "posix_user.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.gid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.uid", "1001"),
					resource.TestCheckResourceAttr(resourceName, "posix_user.0.secondary_gids.#", "1")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEFSAccessPoint_tags(t *testing.T) {
	var ap efs.AccessPointDescription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_efs_access_point.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSAccessPointConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEFSAccessPointConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEFSAccessPointConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSEFSAccessPoint_disappears(t *testing.T) {
	var ap efs.AccessPointDescription
	resourceName := "aws_efs_access_point.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsAccessPointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSAccessPointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsAccessPointExists(resourceName, &ap),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEfsAccessPoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEfsAccessPointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).efsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_access_point" {
			continue
		}

		resp, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
			AccessPointId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, efs.ErrCodeAccessPointNotFound, "") {
				continue
			}
			return fmt.Errorf("Error describing EFS access point in tests: %s", err)
		}
		if len(resp.AccessPoints) > 0 {
			return fmt.Errorf("EFS access point %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckEfsAccessPointExists(resourceID string, mount *efs.AccessPointDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		fs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		mt, err := conn.DescribeAccessPoints(&efs.DescribeAccessPointsInput{
			AccessPointId: aws.String(fs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if *mt.AccessPoints[0].AccessPointId != fs.Primary.ID {
			return fmt.Errorf("access point ID mismatch: %q != %q",
				*mt.AccessPoints[0].AccessPointId, fs.Primary.ID)
		}

		*mount = *mt.AccessPoints[0]

		return nil
	}
}

func testAccAWSEFSAccessPointConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = "%s"
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
}
`, rName)
}

func testAccAWSEFSAccessPointConfigRootDirectory(rName, dir string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  root_directory {
    path = %[2]q
  }
}
`, rName, dir)
}

func testAccAWSEFSAccessPointConfigRootDirectoryCreationInfo(rName, dir string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  root_directory {
    path = %[2]q
    creation_info {
      owner_gid   = 1001
      owner_uid   = 1001
      permissions = "755"
    }
  }
}
`, rName, dir)
}

func testAccAWSEFSAccessPointConfigPosixUser(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = "%s"
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  posix_user {
    gid = 1001
    uid = 1001
  }
}
`, rName)
}

func testAccAWSEFSAccessPointConfigPosixUserSecondaryGids(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = "%s"
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id
  posix_user {
    gid            = 1001
    uid            = 1001
    secondary_gids = [1002]
  }
}
`, rName)
}

func testAccAWSEFSAccessPointConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSEFSAccessPointConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
