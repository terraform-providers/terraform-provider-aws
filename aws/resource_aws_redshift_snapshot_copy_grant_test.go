package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSRedshiftSnapshotCopyGrant_basic(t *testing.T) {
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotCopyGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotCopyGrant_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "snapshot_copy_grant_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
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

func TestAccAWSRedshiftSnapshotCopyGrant_Update(t *testing.T) {
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotCopyGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotCopyGrantWithTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Env", "Test"),
				),
			},
			{
				Config: testAccAWSRedshiftSnapshotCopyGrant_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSRedshiftSnapshotCopyGrantWithTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotCopyGrantExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Env", "Test"),
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

func TestAccAWSRedshiftSnapshotCopyGrant_disappears(t *testing.T) {
	resourceName := "aws_redshift_snapshot_copy_grant.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotCopyGrantDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotCopyGrant_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotCopyGrantExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRedshiftSnapshotCopyGrant(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSRedshiftSnapshotCopyGrantDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).redshiftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_copy_grant" {
			continue
		}

		err := waitForAwsRedshiftSnapshotCopyGrantToBeDeleted(conn, rs.Primary.ID)
		return err
	}

	return nil
}

func testAccCheckAWSRedshiftSnapshotCopyGrantExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Snapshot Copy Grant ID (SnapshotCopyGrantName) is not set")
		}

		// retrieve the client from the test provider
		conn := testAccProvider.Meta().(*AWSClient).redshiftconn

		input := redshift.DescribeSnapshotCopyGrantsInput{
			MaxRecords:            aws.Int64(int64(100)),
			SnapshotCopyGrantName: aws.String(rs.Primary.ID),
		}

		response, err := conn.DescribeSnapshotCopyGrants(&input)

		if err != nil {
			return err
		}

		// we expect only a single snapshot copy grant by this ID. If we find zero, or many,
		// then we consider this an error
		if len(response.SnapshotCopyGrants) != 1 ||
			*response.SnapshotCopyGrants[0].SnapshotCopyGrantName != rs.Primary.ID {
			return fmt.Errorf("Snapshot copy grant not found")
		}

		return nil
	}
}

func testAccAWSRedshiftSnapshotCopyGrant_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_copy_grant" "test" {
  snapshot_copy_grant_name = %[1]q
}
`, rName)
}

func testAccAWSRedshiftSnapshotCopyGrantWithTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_copy_grant" "test" {
  snapshot_copy_grant_name = %[1]q

  tags = {
    Name = %[1]q
    Env  = "Test"
  }
}
`, rName)
}
