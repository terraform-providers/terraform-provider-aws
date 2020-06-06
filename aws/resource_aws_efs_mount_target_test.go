package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/efs"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
					testAccMatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					testAccMatchResourceAttrRegionalARN(resourceName, "file_system_arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
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
					testAccCheckEfsMountTarget(
						resourceName,
						&mount,
					),
					testAccAWSEFSMountTargetDisappears(&mount),
				),
				ExpectNonEmptyPlan: true,
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
			if efsErr, ok := err.(awserr.Error); ok && efsErr.Code() == "MountTargetNotFound" {
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

		if *mt.MountTargets[0].MountTargetId != fs.Primary.ID {
			return fmt.Errorf("Mount target ID mismatch: %q != %q",
				*mt.MountTargets[0].MountTargetId, fs.Primary.ID)
		}

		*mount = *mt.MountTargets[0]

		return nil
	}
}

func testAccAWSEFSMountTargetDisappears(mount *efs.MountTargetDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).efsconn

		_, err := conn.DeleteMountTarget(&efs.DeleteMountTargetInput{
			MountTargetId: mount.MountTargetId,
		})

		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "MountTargetNotFound" {
				return nil
			}
			return err
		}

		return resource.Retry(3*time.Minute, func() *resource.RetryError {
			resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
				MountTargetId: mount.MountTargetId,
			})
			if err != nil {
				if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "MountTargetNotFound" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error reading EFS mount target: %s", err))
			}
			if resp.MountTargets == nil || len(resp.MountTargets) < 1 {
				return nil
			}
			if *resp.MountTargets[0].LifeCycleState == "deleted" {
				return nil
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting for EFS mount target: %s", *mount.MountTargetId))
		})
	}

}

func testAccAWSEFSMountTargetConfig(ct string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "foo" {
  creation_token = "%s"
}

resource "aws_efs_mount_target" "test" {
  file_system_id = "${aws_efs_file_system.foo.id}"
  subnet_id      = "${aws_subnet.test.id}"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-efs-mount-target"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.foo.id}"
  availability_zone = "us-west-2a"
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}
`, ct)
}

func testAccAWSEFSMountTargetConfigModified(ct string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "foo" {
  creation_token = "%s"
}

resource "aws_efs_mount_target" "test" {
  file_system_id = "${aws_efs_file_system.foo.id}"
  subnet_id      = "${aws_subnet.test.id}"
}

resource "aws_efs_mount_target" "test2" {
  file_system_id = "${aws_efs_file_system.foo.id}"
  subnet_id      = "${aws_subnet.test2.id}"
}

resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-efs-mount-target-modified"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.foo.id}"
  availability_zone = "us-west-2a"
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = "${aws_vpc.foo.id}"
  availability_zone = "us-west-2b"
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test2"
  }
}
`, ct)
}
