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
	resource.AddTestSweepers("aws_efs_file_system", &resource.Sweeper{
		Name: "aws_efs_file_system",
		F:    testSweepEfsFileSystems,
		Dependencies: []string{
			"aws_efs_mount_target",
			"aws_efs_access_point",
		},
	})
}

func testSweepEfsFileSystems(region string) error {
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

			log.Printf("[INFO] Deleting EFS File System: %s", id)

			_, err := conn.DeleteFileSystem(&efs.DeleteFileSystemInput{
				FileSystemId: filesystem.FileSystemId,
			})
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting EFS File System %q: %w", id, err))
				continue
			}

			err = waitForDeleteEfsFileSystem(conn, id, 10*time.Minute)
			if err != nil {
				errors = multierror.Append(fmt.Errorf("error waiting for EFS File System %q to delete: %w", id, err))
				continue
			}
		}
		return true
	})
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS File Systems: %w", err))
	}

	return errors
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

func TestAccAWSEFSFileSystem_basic(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"
	rName := acctest.RandomWithPrefix("tf-acc")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
					testAccMatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					resource.TestCheckResourceAttr(resourceName, "performance_mode", "generalPurpose"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeBursting),
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemPerformanceMode(resourceName, "generalPurpose"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"creation_token"},
			},
			{
				Config: testAccAWSEFSFileSystemConfigWithPerformanceMode,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem("aws_efs_file_system.test2", &desc),
					testAccCheckEfsCreationToken("aws_efs_file_system.test2", "supercalifragilisticexpialidocious"),
					testAccCheckEfsFileSystemPerformanceMode("aws_efs_file_system.test2", "maxIO"),
				),
			},
		},
	})
}

func TestAccAWSEFSFileSystem_tags(t *testing.T) {
	var desc efs.FileSystemDescription
	rName := acctest.RandomWithPrefix("tf-acc-tags")
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"creation_token"},
			},
			{
				Config: testAccAWSEFSFileSystemConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfigWithMaxTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "50"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Another", "tag"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag45", "TestTagValue"),
				),
			},
		},
	})
}

func TestAccAWSEFSFileSystem_pagedTags(t *testing.T) {
	var desc efs.FileSystemDescription
	rInt := acctest.RandInt()
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigPagedTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "10"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"creation_token"},
			},
		},
	})
}

func TestAccAWSEFSFileSystem_kmsKey(t *testing.T) {
	var desc efs.FileSystemDescription
	rInt := acctest.RandInt()
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigWithKmsKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "encrypted", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"creation_token"},
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
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfig_ProvisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfig_ProvisionedThroughputInMibps(2.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "2"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"creation_token"},
			},
		},
	})
}

func TestAccAWSEFSFileSystem_ThroughputMode(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfig_ProvisionedThroughputInMibps(1.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeProvisioned),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfig_ThroughputMode(efs.ThroughputModeBursting),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "provisioned_throughput_in_mibps", "0"),
					resource.TestCheckResourceAttr(resourceName, "throughput_mode", efs.ThroughputModeBursting),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"creation_token"},
			},
		},
	})
}

func TestAccAWSEFSFileSystem_lifecyclePolicy(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigWithLifecyclePolicy(
					"transition_to_ia",
					"invalid_value",
				),
				ExpectError: regexp.MustCompile(`got invalid_value`),
			},
			{
				Config: testAccAWSEFSFileSystemConfigWithLifecyclePolicy(
					"transition_to_ia",
					efs.TransitionToIARulesAfter30Days,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemLifecyclePolicy(resourceName, "badExpectation"),
				),
				ExpectError: regexp.MustCompile(`Expected: badExpectation`),
			},
			{
				Config: testAccAWSEFSFileSystemConfigWithLifecyclePolicy(
					"transition_to_ia",
					efs.TransitionToIARulesAfter30Days,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemLifecyclePolicy(resourceName, efs.TransitionToIARulesAfter30Days),
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

func TestAccAWSEFSFileSystem_lifecyclePolicy_update(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigWithLifecyclePolicy("transition_to_ia", efs.TransitionToIARulesAfter30Days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemLifecyclePolicy(resourceName, efs.TransitionToIARulesAfter30Days),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfigWithLifecyclePolicy("transition_to_ia", efs.TransitionToIARulesAfter90Days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemLifecyclePolicy(resourceName, efs.TransitionToIARulesAfter90Days),
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

func TestAccAWSEFSFileSystem_lifecyclePolicy_removal(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfigWithLifecyclePolicy("transition_to_ia", efs.TransitionToIARulesAfter14Days),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemLifecyclePolicy(resourceName, efs.TransitionToIARulesAfter14Days),
				),
			},
			{
				Config: testAccAWSEFSFileSystemConfigRemovedLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemLifecyclePolicy(resourceName, efs.TransitionToIARulesAfter14Days),
				),
				ExpectError: regexp.MustCompile(fmt.Sprintf(`Expected: %s`, efs.TransitionToIARulesAfter14Days)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEFSFileSystem_disappears(t *testing.T) {
	var desc efs.FileSystemDescription
	resourceName := "aws_efs_file_system.test"
	rName := acctest.RandomWithPrefix("tf-acc-disappears")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystem(resourceName, &desc),
					testAccCheckEfsFileSystemDisappears(&desc),
				),
				ExpectNonEmptyPlan: true,
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
			if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") {
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

func testAccCheckEfsFileSystem(resourceID string, fDesc *efs.FileSystemDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		fs, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(fs.FileSystems) == 0 {
			return fmt.Errorf("EFS File System not found")
		}

		*fDesc = *fs.FileSystems[0]

		return nil
	}
}

func testAccCheckEfsFileSystemDisappears(fDesc *efs.FileSystemDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).efsconn

		input := &efs.DeleteFileSystemInput{
			FileSystemId: fDesc.FileSystemId,
		}

		_, err := conn.DeleteFileSystem(input)

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

func testAccCheckEfsFileSystemLifecyclePolicy(resourceID string, expectedVal string) resource.TestCheckFunc {
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
		if err != nil {
			return fmt.Errorf("Error describing EFS file systems: %s", err.Error())
		}

		fs := resp.FileSystems[0]

		res, err := conn.DescribeLifecycleConfiguration(&efs.DescribeLifecycleConfigurationInput{
			FileSystemId: fs.FileSystemId,
		})
		if err != nil {
			return fmt.Errorf("Error describing lifecycle policy for EFS file system (%s): %s",
				aws.StringValue(fs.FileSystemId), err.Error())
		}
		lp := res.LifecyclePolicies

		newLP := make([]*map[string]interface{}, len(lp))

		for i := 0; i < len(lp); i++ {
			config := lp[i]
			data := make(map[string]interface{})
			newLP[i] = &data
			if *config.TransitionToIA == expectedVal {
				return nil
			}
		}
		return fmt.Errorf("Lifecycle Policy mismatch.\nExpected: %s\nFound: %+v", expectedVal, lp)
	}
}

func testAccAWSEFSFileSystemConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %q
}
`, rName)
}

func testAccAWSEFSFileSystemConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSEFSFileSystemConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSEFSFileSystemConfigPagedTags(rInt int) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name           = "test-efs-%[1]d"
    Another        = "tag"
    Test           = "yes"
    User           = "root"
    Page           = "1"
    Environment    = "prod"
    CostCenter     = "terraform"
    AcceptanceTest = "PagedTags"
    CreationToken  = "radek"
    PerfMode       = "max"
  }
}
`, rInt)
}

func testAccAWSEFSFileSystemConfigWithMaxTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name    = %q
    Another = "tag"
    Tag00   = "TestTagValue"
    Tag01   = "TestTagValue"
    Tag02   = "TestTagValue"
    Tag03   = "TestTagValue"
    Tag04   = "TestTagValue"
    Tag05   = "TestTagValue"
    Tag06   = "TestTagValue"
    Tag07   = "TestTagValue"
    Tag08   = "TestTagValue"
    Tag09   = "TestTagValue"
    Tag10   = "TestTagValue"
    Tag11   = "TestTagValue"
    Tag12   = "TestTagValue"
    Tag13   = "TestTagValue"
    Tag14   = "TestTagValue"
    Tag15   = "TestTagValue"
    Tag16   = "TestTagValue"
    Tag17   = "TestTagValue"
    Tag18   = "TestTagValue"
    Tag19   = "TestTagValue"
    Tag20   = "TestTagValue"
    Tag21   = "TestTagValue"
    Tag22   = "TestTagValue"
    Tag23   = "TestTagValue"
    Tag24   = "TestTagValue"
    Tag25   = "TestTagValue"
    Tag26   = "TestTagValue"
    Tag27   = "TestTagValue"
    Tag28   = "TestTagValue"
    Tag29   = "TestTagValue"
    Tag30   = "TestTagValue"
    Tag31   = "TestTagValue"
    Tag32   = "TestTagValue"
    Tag33   = "TestTagValue"
    Tag34   = "TestTagValue"
    Tag35   = "TestTagValue"
    Tag36   = "TestTagValue"
    Tag37   = "TestTagValue"
    Tag38   = "TestTagValue"
    Tag39   = "TestTagValue"
    Tag40   = "TestTagValue"
    Tag41   = "TestTagValue"
    Tag42   = "TestTagValue"
    Tag43   = "TestTagValue"
    Tag44   = "TestTagValue"
    Tag45   = "TestTagValue"
    Tag46   = "TestTagValue"
    Tag47   = "TestTagValue"
  }
}
`, rName)
}

const testAccAWSEFSFileSystemConfigWithPerformanceMode = `
resource "aws_efs_file_system" "test2" {
  creation_token   = "supercalifragilisticexpialidocious"
  performance_mode = "maxIO"
}
`

func testAccAWSEFSFileSystemConfigWithKmsKey(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test %d"
}

resource "aws_efs_file_system" "test" {
  encrypted  = true
  kms_key_id = aws_kms_key.test.arn
}
`, rInt)
}

func testAccAWSEFSFileSystemConfigWithKmsKeyNoEncryption(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = "Terraform acc test %d"
}

resource "aws_efs_file_system" "test" {
  encrypted  = false
  kms_key_id = aws_kms_key.test.arn
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

func testAccAWSEFSFileSystemConfigWithLifecyclePolicy(lpName string, lpVal string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  lifecycle_policy {
    %s = %q
  }
}
`, lpName, lpVal)
}

const testAccAWSEFSFileSystemConfigRemovedLifecyclePolicy = `
resource "aws_efs_file_system" "test" {}
`
