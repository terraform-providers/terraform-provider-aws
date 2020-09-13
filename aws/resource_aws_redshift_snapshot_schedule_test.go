package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_redshift_snapshot_schedule", &resource.Sweeper{
		Name: "aws_redshift_snapshot_schedule",
		F:    testSweepRedshiftSnapshotSchedules,
	})
}

func testSweepRedshiftSnapshotSchedules(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).redshiftconn

	req := &redshift.DescribeSnapshotSchedulesInput{}

	resp, err := conn.DescribeSnapshotSchedules(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Regional Snapshot Schedules sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Redshift Regional Snapshot Schedules: %s", err)
	}

	if len(resp.SnapshotSchedules) == 0 {
		log.Print("[DEBUG] No AWS Redshift Regional Snapshot Schedules to sweep")
		return nil
	}

	for _, snapshotSchedules := range resp.SnapshotSchedules {
		identifier := aws.StringValue(snapshotSchedules.ScheduleIdentifier)

		hasPrefix := false
		prefixes := []string{"tf-snapshot-schedule-"}

		for _, prefix := range prefixes {
			if strings.HasPrefix(identifier, prefix) {
				hasPrefix = true
				break
			}
		}

		if !hasPrefix {
			log.Printf("[INFO] Skipping Delete Redshift Snapshot Schedule: %s", identifier)
			continue
		}

		_, err := conn.DeleteSnapshotSchedule(&redshift.DeleteSnapshotScheduleInput{
			ScheduleIdentifier: snapshotSchedules.ScheduleIdentifier,
		})
		if isAWSErr(err, redshift.ErrCodeSnapshotScheduleNotFoundFault, "") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error deleting Redshift Snapshot Schedule %s: %s", identifier, err)
		}
	}

	return nil
}

func TestAccAWSRedshiftSnapshotSchedule_basic(t *testing.T) {
	var v redshift.SnapshotSchedule

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfig(rName, "rate(12 hours)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "1"),
				),
			},
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfig(rName, "cron(30 12 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func TestAccAWSRedshiftSnapshotSchedule_withMultipleDefinition(t *testing.T) {
	var v redshift.SnapshotSchedule

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfigWithMultipleDefinition(rName, "cron(30 12 *)", "cron(15 6 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "2"),
				),
			},
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfigWithMultipleDefinition(rName, "cron(30 8 *)", "cron(15 10 *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "definitions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})

}

func TestAccAWSRedshiftSnapshotSchedule_withIdentifierPrefix(t *testing.T) {
	var v redshift.SnapshotSchedule

	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfigWithIdentifierPrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"identifier_prefix",
					"force_destroy",
				},
			},
		},
	})
}

func TestAccAWSRedshiftSnapshotSchedule_withDescription(t *testing.T) {
	var v redshift.SnapshotSchedule

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfigWithDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Test Schedule"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func TestAccAWSRedshiftSnapshotSchedule_withTags(t *testing.T) {
	var v redshift.SnapshotSchedule

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_snapshot_schedule.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfigWithTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfigWithTagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(resourceName, "tags.good", "bad"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func TestAccAWSRedshiftSnapshotSchedule_withForceDestroy(t *testing.T) {
	var snapshotSchedule redshift.SnapshotSchedule
	var cluster redshift.Cluster
	rInt := acctest.RandInt()
	rName := acctest.RandString(8)
	resourceName := "aws_redshift_snapshot_schedule.default"
	clusterResourceName := "aws_redshift_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotScheduleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotScheduleConfigWithForceDestroy(rInt, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleExists(resourceName, &snapshotSchedule),
					testAccCheckAWSRedshiftClusterExists(clusterResourceName, &cluster),
					testAccCheckAWSRedshiftSnapshotScheduleCreateSnapshotScheduleAssociation(&cluster, &snapshotSchedule),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy",
				},
			},
		},
	})
}

func testAccCheckAWSRedshiftSnapshotScheduleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_schedule" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(rs.Primary.ID),
		})

		if err == nil {
			if len(resp.SnapshotSchedules) != 0 {
				for _, s := range resp.SnapshotSchedules {
					if *s.ScheduleIdentifier == rs.Primary.ID {
						return fmt.Errorf("Redshift Cluster Snapshot Schedule %s still exists", rs.Primary.ID)
					}
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSRedshiftSnapshotScheduleExists(n string, v *redshift.SnapshotSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster Snapshot Schedule ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, s := range resp.SnapshotSchedules {
			if *s.ScheduleIdentifier == rs.Primary.ID {
				*v = *s
				return nil
			}
		}

		return fmt.Errorf("Redshift Snapshot Schedule (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSRedshiftSnapshotScheduleCreateSnapshotScheduleAssociation(cluster *redshift.Cluster, snapshotSchedule *redshift.SnapshotSchedule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).redshiftconn

		if _, err := conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
			ClusterIdentifier:    cluster.ClusterIdentifier,
			ScheduleIdentifier:   snapshotSchedule.ScheduleIdentifier,
			DisassociateSchedule: aws.Bool(false),
		}); err != nil {
			return fmt.Errorf("Error associate Redshift Cluster and Snapshot Schedule: %s", err)
		}

		if err := waitForRedshiftSnapshotScheduleAssociationActive(conn, 75*time.Minute, aws.StringValue(cluster.ClusterIdentifier), aws.StringValue(snapshotSchedule.ScheduleIdentifier)); err != nil {
			return err
		}

		return nil
	}
}

const testAccAWSRedshiftSnapshotScheduleConfigWithIdentifierPrefix = `
resource "aws_redshift_snapshot_schedule" "default" {
  identifier_prefix = "tf-snapshot-schedule-"
  definitions = [
    "rate(12 hours)",
  ]
}
`

func testAccAWSRedshiftSnapshotScheduleConfig(rName, definition string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = "tf-snapshot-schedule-%[1]s"
  definitions = [
    "%[2]s",
  ]
}
`, rName, definition)
}

func testAccAWSRedshiftSnapshotScheduleConfigWithMultipleDefinition(rName, definition1, definition2 string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = "tf-snapshot-schedule-%[1]s"
  definitions = [
    "%[2]s",
    "%[3]s",
  ]
}
`, rName, definition1, definition2)
}

func testAccAWSRedshiftSnapshotScheduleConfigWithDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier  = "tf-snapshot-schedule-%[1]s"
  description = "Test Schedule"
  definitions = [
    "rate(12 hours)",
  ]
}
`, rName)
}

func testAccAWSRedshiftSnapshotScheduleConfigWithTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = "tf-snapshot-schedule-%[1]s"
  definitions = [
    "rate(12 hours)",
  ]

  tags = {
    foo  = "bar"
    fizz = "buzz"
  }
}
`, rName)
}

func testAccAWSRedshiftSnapshotScheduleConfigWithTagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshift_snapshot_schedule" "default" {
  identifier = "tf-snapshot-schedule-%[1]s"
  definitions = [
    "rate(12 hours)",
  ]

  tags = {
    foo  = "bar2"
    good = "bad"
  }
}
`, rName)
}

func testAccAWSRedshiftSnapshotScheduleConfigWithForceDestroy(rInt int, rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_redshift_snapshot_schedule" "default" {
  identifier  = "tf-snapshot-schedule-%[2]s"
  description = "Test Schedule"
  definitions = [
    "rate(12 hours)",
  ]
  force_destroy = true
}
`, testAccAWSRedshiftClusterConfig_basic(rInt), rName)
}
