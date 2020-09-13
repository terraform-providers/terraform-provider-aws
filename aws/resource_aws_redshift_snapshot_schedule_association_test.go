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

func TestAccAWSRedshiftSnapshotScheduleAssociation_basic(t *testing.T) {
	rInt := acctest.RandInt()
	rName := acctest.RandString(8)
	resourceName := "aws_redshift_snapshot_schedule_association.default"
	snapshotScheduleResourceName := "aws_redshift_snapshot_schedule.default"
	clusterResourceName := "aws_redshift_cluster.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftSnapshotScheduleAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftSnapshotScheduleAssociationConfig(rInt, rName, "rate(12 hours)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftSnapshotScheduleAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "schedule_identifier", snapshotScheduleResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_identifier", clusterResourceName, "id"),
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

func testAccCheckAWSRedshiftSnapshotScheduleAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_snapshot_schedule_association" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		clusterIdentifier, scheduleIdentifier, err := resourceAwsRedshiftSnapshotScheduleAssociationParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(scheduleIdentifier),
			ClusterIdentifier:  aws.String(clusterIdentifier),
		})

		if err != nil {
			return err
		}

		if resp != nil && len(resp.SnapshotSchedules) > 0 {
			return fmt.Errorf("Redshift Cluster (%s) Snapshot Schedule (%s) Association still exist", clusterIdentifier, scheduleIdentifier)
		}

		return err
	}

	return nil
}

func testAccCheckAWSRedshiftSnapshotScheduleAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster Snapshot Schedule Association ID is set")
		}

		clusterIdentifier, scheduleIdentifier, err := resourceAwsRedshiftSnapshotScheduleAssociationParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ScheduleIdentifier: aws.String(scheduleIdentifier),
			ClusterIdentifier:  aws.String(clusterIdentifier),
		})

		if err != nil {
			return err
		}

		if len(resp.SnapshotSchedules) != 0 {
			return nil
		}

		return fmt.Errorf("Redshift Cluster (%s) Snapshot Schedule (%s) Association not found", clusterIdentifier, scheduleIdentifier)
	}
}

func testAccAWSRedshiftSnapshotScheduleAssociationConfig(rInt int, rName, definition string) string {
	return fmt.Sprintf(`
%s
%s

resource "aws_redshift_snapshot_schedule_association" "default" {
  schedule_identifier = aws_redshift_snapshot_schedule.default.id
  cluster_identifier  = aws_redshift_cluster.default.id
}
`, testAccAWSRedshiftClusterConfig_basic(rInt), testAccAWSRedshiftSnapshotScheduleConfig(rName, definition))
}
