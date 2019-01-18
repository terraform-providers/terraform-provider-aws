package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/efs"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestResourceAWSEFSFileSystem_validateReferenceName(t *testing.T) {
	var value string
	var errors []error

	value = acctest.RandString(128)
	_, errors = validateReferenceName(value, "reference_name")
	if len(errors) == 0 {
		t.Fatalf("Expected to trigger a validation error")
	}

	value = acctest.RandString(32)
	_, errors = validateReferenceName(value, "reference_name")
	if len(errors) != 0 {
		t.Fatalf("Expected not to trigger a validation error")
	}
}

func TestResourceAWSEFSFileSystem_hasEmptyFileSystems(t *testing.T) {
	fs := &efs.DescribeFileSystemsOutput{
		FileSystems: []*efs.FileSystemDescription{},
	}

	actual := hasEmptyFileSystems(fs)
	if !actual {
		t.Fatalf("Expected return value to be true, got %t", actual)
	}

	// Add an empty file system.
	fs.FileSystems = append(fs.FileSystems, &efs.FileSystemDescription{})

	actual = hasEmptyFileSystems(fs)
	if actual {
		t.Fatalf("Expected return value to be false, got %t", actual)
	}

}

func TestAccAWSEFSFileSystem_importBasic(t *testing.T) {
	resourceName := "aws_efs_file_system.foo-with-tags"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigWithTags(rInt),
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name", "creation_token"},
			},
		},
	})
}

func TestAccAWSEFSFileSystem_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN("aws_efs_file_system.foo", "arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestCheckResourceAttr(
						"aws_efs_file_system.foo",
						"performance_mode",
						"generalPurpose"),
					resource.TestCheckResourceAttr(
						"aws_efs_file_system.foo",
						"throughput_mode",
						efs.ThroughputModeBursting),
					testAccCheckEfsFileSystem(
						"aws_efs_file_system.foo",
					),
					testAccCheckEfsFileSystemPerformanceMode(
						"aws_efs_file_system.foo",
						"generalPurpose",
					),
					resource.TestMatchResourceAttr(
						"aws_efs_file_system.foo",
						"dns_name",
						regexp.MustCompile("^[^.]+.efs.us-west-2.amazonaws.com$"),
					),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(
						"aws_efs_file_system.foo-with-tags",
					),
					testAccCheckEfsFileSystemPerformanceMode(
						"aws_efs_file_system.foo-with-tags",
						"generalPurpose",
					),
					testAccCheckEfsFileSystemTags(
						"aws_efs_file_system.foo-with-tags",
						map[string]string{
							"Name":    fmt.Sprintf("foo-efs-%d", rInt),
							"Another": "tag",
						},
					),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfigWithPerformanceMode,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(
						"aws_efs_file_system.foo-with-performance-mode",
					),
					testAccCheckEfsCreationToken(
						"aws_efs_file_system.foo-with-performance-mode",
						"supercalifragilisticexpialidocious",
					),
					testAccCheckEfsFileSystemPerformanceMode(
						"aws_efs_file_system.foo-with-performance-mode",
						"maxIO",
					),
				),
			},
		},
	})
}

func TestAccAWSEFSFileSystem_pagedTags(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigPagedTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_efs_file_system.foo",
						"tags.%",
						"11"),
					//testAccCheckEfsFileSystem(
					//	"aws_efs_file_system.foo",
					//),
					//testAccCheckEfsFileSystemPerformanceMode(
					//	"aws_efs_file_system.foo",
					//	"generalPurpose",
					//),
				),
			},
		},
	})
}

func TestAccAWSEFSFileSystem_kmsKey(t *testing.T) {
	rInt := acctest.RandInt()
	kmsKeyResourceName := "aws_kms_key.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigWithKmsKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aws_efs_file_system.foo-with-kms", "kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr("aws_efs_file_system.foo-with-kms", "encrypted", "true"),
				),
			},
		},
	})
}

func TestAccAWSEFSFileSystem_kmsConfigurationWithoutEncryption(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEFSFileSystemConfigWithKmsKeyNoEncryption(rInt),
				ExpectError: regexp.MustCompile(`encrypted must be set to true when kms_key_id is specified`),
			},
		},
	})
}

func TestAccAWSEFSFileSystem_ProvisionedThroughputInMibps(t *testing.T) {
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfig_ProvisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfig_ProvisionedThroughputInMibps(2.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "2"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name", "creation_token"},
			},
		},
	})
}

func TestAccAWSEFSFileSystem_ThroughputMode(t *testing.T) {
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfig_ProvisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfig_ThroughputMode(efs.ThroughputModeBursting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "0"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeBursting),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"reference_name", "creation_token"},
			},
		},
	})
}

func testAccCheckEfsFileSystemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).efsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_file_system" {
			continue
		}

		resp, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if efsErr, ok := err.(awserr.Error); ok && efsErr.Code() == "FileSystemNotFound" {
				// gone
				return nil
			}
			return fmt.Errorf("Error describing EFS in tests: %s", err)
		}
		if len(resp.FileSystems) > 0 {
			return fmt.Errorf("EFS file system %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckEfsFileSystem(resourceID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		_, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckEfsCreationToken(resourceID string, expectedToken string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		resp, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})

		fs := resp.FileSystems[0]
		if *fs.CreationToken != expectedToken {
			return fmt.Errorf("Creation Token mismatch.\nExpected: %s\nGiven: %v",
				expectedToken, *fs.CreationToken)
		}

		return err
	}
}

func testAccCheckEfsFileSystemTags(resourceID string, expectedTags map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		resp, err := conn.DescribeTags(&efs.DescribeTagsInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})

		if !reflect.DeepEqual(expectedTags, tagsToMapEFS(resp.Tags)) {
			return fmt.Errorf("Tags mismatch.\nExpected: %#v\nGiven: %#v",
				expectedTags, resp.Tags)
		}

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckEfsFileSystemPerformanceMode(resourceID string, expectedMode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		resp, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})

		fs := resp.FileSystems[0]
		if *fs.PerformanceMode != expectedMode {
			return fmt.Errorf("Performance Mode mismatch.\nExpected: %s\nGiven: %v",
				expectedMode, *fs.PerformanceMode)
		}

		if err != nil {
			return err
		}

		return nil
	}
}

const testAccAWSEFSFileSystemConfig = `
resource "aws_efs_file_system" "foo" {
	creation_token = "radeksimko"
}
`

func testAccAWSEFSFileSystemConfigPagedTags(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_efs_file_system" "foo" {
	tags = {
			Name = "foo-efs-%d"
			Another = "tag"
			Test = "yes"
			User = "root"
			Page = "1"
			Environment = "prod"
			CostCenter = "terraform"
			AcceptanceTest = "PagedTags"
			CreationToken = "radek"
			PerfMode = "max"
			Region = "us-west-2"
		}
	}
	`, rInt)
}

func testAccAWSEFSFileSystemConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
	resource "aws_efs_file_system" "foo-with-tags" {
	tags = {
			Name = "foo-efs-%d"
			Another = "tag"
		}
	}
	`, rInt)
}

const testAccAWSEFSFileSystemConfigWithPerformanceMode = `
resource "aws_efs_file_system" "foo-with-performance-mode" {
	creation_token = "supercalifragilisticexpialidocious"
	performance_mode = "maxIO"
}
`

func testAccAWSEFSFileSystemConfigWithKmsKey(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %d"
}

resource "aws_efs_file_system" "foo-with-kms" {
  encrypted = true
  kms_key_id = "${aws_kms_key.foo.arn}"
}
`, rInt)
}

func testAccAWSEFSFileSystemConfigWithKmsKeyNoEncryption(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description = "Terraform acc test %d"
}

resource "aws_efs_file_system" "foo-with-kms" {
  encrypted = false
  kms_key_id = "${aws_kms_key.foo.arn}"
}
`, rInt)
}

func testAccAWSEFSFileSystemConfig_ThroughputMode(throughputMode string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  throughput_mode = %q
}
`, throughputMode)
}

func testAccAWSEFSFileSystemConfig_ProvisionedThroughputInMibps(provisionedThroughputInMibps float64) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  provisioned_throughput_in_mibps = %f
  throughput_mode                 = "provisioned"
}
`, provisionedThroughputInMibps)
}
