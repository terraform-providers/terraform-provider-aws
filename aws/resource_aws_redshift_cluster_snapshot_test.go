package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("aws_redshift_cluster_snapshot", &resource.Sweeper{
		Name: "aws_redshift_cluster_snapshot",
		F:    testSweepRedshiftClusterSnapshots,
		Dependencies: []string{
			"aws_redshift_cluster",
		},
	})
}

func testSweepRedshiftClusterSnapshots(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).redshiftconn

	err = conn.DescribeClusterSnapshotsPages(&redshift.DescribeClusterSnapshotsInput{}, func(resp *redshift.DescribeClusterSnapshotsOutput, isLast bool) bool {
		if len(resp.Snapshots) == 0 {
			log.Print("[DEBUG] No Redshift cluster snapshots to sweep")
			return false
		}

		for _, s := range resp.Snapshots {
			id := aws.StringValue(s.SnapshotIdentifier)

			if !strings.EqualFold(aws.StringValue(s.SnapshotType), "manual") || !strings.EqualFold(aws.StringValue(s.Status), "available") {
				log.Printf("[INFO] Skipping Redshift cluster snapshot: %s", id)
				continue
			}

			log.Printf("[INFO] Deleting Redshift cluster snapshot: %s", id)
			_, err := conn.DeleteClusterSnapshot(&redshift.DeleteClusterSnapshotInput{
				SnapshotIdentifier: s.SnapshotIdentifier,
			})
			if err != nil {
				log.Printf("[ERROR] Failed deleting Redshift cluster snapshot (%s): %s", id, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Cluster Snapshot sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Redshift cluster snapshots: %w", err)
	}
	return nil
}
