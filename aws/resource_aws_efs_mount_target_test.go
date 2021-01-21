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
	resource.AddTestSweepers("aws_efs_mount_target", &resource.Sweeper{
		Name: "aws_efs_mount_target",
		F:    testSweepEfsMountTargets,
	})
}

func testSweepEfsMountTargets(region string) error {
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
			log.Printf("[INFO] Deleting Mount Targets for EFS File System: %s", id)

			var errors error
			input := &efs.DescribeMountTargetsInput{
				FileSystemId: filesystem.FileSystemId,
			}
			for {
				out, err := conn.DescribeMountTargets(input)
				if err != nil {
					errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS Mount Targets on File System %q: %w", id, err))
					break
				}

				if out == nil || len(out.MountTargets) == 0 {
					log.Printf("[INFO] No EFS Mount Targets to sweep on File System %q", id)
					break
				}

				for _, mounttarget := range out.MountTargets {
					id := aws.StringValue(mounttarget.MountTargetId)

					log.Printf("[INFO] Deleting EFS Mount Target: %s", id)
					_, err := conn.DeleteMountTarget(&efs.DeleteMountTargetInput{
						MountTargetId: mounttarget.MountTargetId,
					})
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error deleting EFS Mount Target %q: %w", id, err))
						continue
					}

					err = waitForDeleteEfsMountTarget(conn, id, 10*time.Minute)
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error waiting for EFS Mount Target %q to delete: %w", id, err))
						continue
					}
				}

				if out.NextMarker == nil {
					break
				}
				input.Marker = out.NextMarker
			}
		}
		return true
	})
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS File Systems: %w", err))
	}

	return errors
}

func TestAccAWSEFSMountTarget_basic(t *testing.T) {
	var mount efs.MountTargetDescription
	ct := fmt.Sprintf("createtoken-%d", acctest.RandInt())
	resourceName := "aws_efs_mount_target.test"
	resourceName2 := "aws_efs_mount_target.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSMountTargetConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_name"),
					testAccMatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					testAccMatchResourceAttrRegionalARN(resourceName, "file_system_arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
					resource.TestCheckResourceAttrSet(resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface_id"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEFSMountTargetConfigModified(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					testAccCheckEfsMountTarget(resourceName2, &mount),
					testAccMatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					testAccMatchResourceAttrRegionalHostname(resourceName2, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
				),
			},
		},
	})
}

func TestAccAWSEFSMountTarget_disappears(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"
	ct := fmt.Sprintf("createtoken-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSMountTargetConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEfsMountTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEFSMountTarget_IpAddress(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSMountTargetConfigIpAddress("10.0.0.100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.0.100"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13845
func TestAccAWSEFSMountTarget_IpAddress_EmptyString(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSMountTargetConfigIpAddress(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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

func TestResourceAWSEFSMountTarget_hasEmptyMountTargets(t *testing.T) {
	mto := &efs.DescribeMountTargetsOutput{
		MountTargets: []*efs.MountTargetDescription{},
	}

	actual := hasEmptyMountTargets(mto)
	if !actual {
		t.Fatalf("Expected return value to be true, got %t", actual)
	}

	// Add an empty mount target.
	mto.MountTargets = append(mto.MountTargets, &efs.MountTargetDescription{})

	actual = hasEmptyMountTargets(mto)
	if actual {
		t.Fatalf("Expected return value to be false, got %t", actual)
	}

}

func testAccCheckEfsMountTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).efsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_mount_target" {
			continue
		}

		resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			MountTargetId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, efs.ErrCodeMountTargetNotFound, "") {
				// gone
				return nil
			}
			return fmt.Errorf("Error describing EFS Mount in tests: %s", err)
		}
		if len(resp.MountTargets) > 0 {
			return fmt.Errorf("EFS Mount target %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckEfsMountTarget(resourceID string, mount *efs.MountTargetDescription) resource.TestCheckFunc {
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
		mt, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			MountTargetId: aws.String(fs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if aws.StringValue(mt.MountTargets[0].MountTargetId) != fs.Primary.ID {
			return fmt.Errorf("Mount target ID mismatch: %q != %q",
				*mt.MountTargets[0].MountTargetId, fs.Primary.ID)
		}

		*mount = *mt.MountTargets[0]

		return nil
	}
}

func testAccAWSEFSMountTargetConfig(ct string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  creation_token = "%s"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}
`, ct)
}

func testAccAWSEFSMountTargetConfigModified(ct string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  creation_token = "%s"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_efs_mount_target" "test2" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test2.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test2"
  }
}
`, ct)
}

func testAccAWSEFSMountTargetConfigIpAddress(ipAddress string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  ip_address     = %[1]q
  subnet_id      = aws_subnet.test.id
}
`, ipAddress)
}
