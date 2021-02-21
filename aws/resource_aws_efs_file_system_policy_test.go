package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEFSFileSystemPolicy_basic(t *testing.T) {
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicy(resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEFSFileSystemPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicy(resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func TestAccAWSEFSFileSystemPolicy_disappears(t *testing.T) {
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckEfsFileSystemPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEFSFileSystemPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicy(resourceName, &desc),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEfsFileSystemPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEfsFileSystemPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).efsconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_file_system_policy" {
			continue
		}

		resp, err := conn.DescribeFileSystemPolicy(&efs.DescribeFileSystemPolicyInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") ||
				isAWSErr(err, efs.ErrCodePolicyNotFound, "") {
				return nil
			}
			return fmt.Errorf("error describing EFS file system policy in tests: %s", err)
		}
		if resp != nil {
			return fmt.Errorf("EFS file system policy %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckEfsFileSystemPolicy(resourceID string, efsFsPolicy *efs.DescribeFileSystemPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		fs, err := conn.DescribeFileSystemPolicy(&efs.DescribeFileSystemPolicyInput{
			FileSystemId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*efsFsPolicy = *fs

		return nil
	}
}

func testAccAWSEFSFileSystemPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleSatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": [
                "elasticfilesystem:ClientMount",
                "elasticfilesystem:ClientWrite"
            ],
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName)
}

func testAccAWSEFSFileSystemPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleSatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": "elasticfilesystem:ClientMount",
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName)
}
