package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsRedshiftSnapshotScheduleAssociation() *schema.Resource {

	return &schema.Resource{
		Create: resourceAwsRedshiftSnapshotScheduleAssociationCreate,
		Read:   resourceAwsRedshiftSnapshotScheduleAssociationRead,
		Delete: resourceAwsRedshiftSnapshotScheduleAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				clusterIdentifier, scheduleIdentifier, err := resourceAwsRedshiftSnapshotScheduleAssociationParseId(d.Id())
				if err != nil {
					return nil, fmt.Errorf("Error parse Redshift Cluster Snapshot Schedule Association ID %s: %s", d.Id(), err)
				}

				d.Set("cluster_identifier", clusterIdentifier)
				d.Set("schedule_identifier", scheduleIdentifier)
				d.SetId(fmt.Sprintf("%s/%s", clusterIdentifier, scheduleIdentifier))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schedule_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsRedshiftSnapshotScheduleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	clusterIdentifier := d.Get("cluster_identifier").(string)
	scheduleIdentifier := d.Get("schedule_identifier").(string)

	_, err := conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(false),
	})

	if err != nil {
		return fmt.Errorf("Error associating Redshift Cluster (%s) and Snapshot Schedule (%s): %s", clusterIdentifier, scheduleIdentifier, err)
	}

	if err := waitForRedshiftSnapshotScheduleAssociationActive(conn, 75*time.Minute, clusterIdentifier, scheduleIdentifier); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", clusterIdentifier, scheduleIdentifier))

	return resourceAwsRedshiftSnapshotScheduleAssociationRead(d, meta)
}

func resourceAwsRedshiftSnapshotScheduleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	clusterIdentifier, scheduleIdentifier, err := resourceAwsRedshiftSnapshotScheduleAssociationParseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error parse Redshift Cluster Snapshot Schedule Association ID %s: %s", d.Id(), err)
	}

	descOpts := &redshift.DescribeSnapshotSchedulesInput{
		ClusterIdentifier:  aws.String(clusterIdentifier),
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	}

	resp, err := conn.DescribeSnapshotSchedules(descOpts)
	if err != nil {
		return fmt.Errorf("Error describing Redshift Cluster %s Snapshot Schedule %s: %s", clusterIdentifier, clusterIdentifier, err)
	}

	if resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) == 0 {
		return fmt.Errorf("Unable to find Redshift Cluster (%s) Snapshot Schedule (%s) Association", clusterIdentifier, scheduleIdentifier)
	}
	snapshotSchedule := resp.SnapshotSchedules[0]
	if snapshotSchedule.AssociatedClusters == nil || aws.Int64Value(snapshotSchedule.AssociatedClusterCount) == 0 {
		return fmt.Errorf("Unable to find Redshift Cluster (%s)", clusterIdentifier)
	}

	var associatedCluster *redshift.ClusterAssociatedToSchedule
	for _, cluster := range snapshotSchedule.AssociatedClusters {
		if *cluster.ClusterIdentifier == clusterIdentifier {
			associatedCluster = cluster
			break
		}
	}

	if associatedCluster == nil {
		return fmt.Errorf("Unable to find Redshift Cluster (%s)", clusterIdentifier)
	}

	d.Set("cluster_identifier", associatedCluster.ClusterIdentifier)
	d.Set("schedule_identifier", snapshotSchedule.ScheduleIdentifier)

	return nil
}

func resourceAwsRedshiftSnapshotScheduleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	clusterIdentifier, scheduleIdentifier, err := resourceAwsRedshiftSnapshotScheduleAssociationParseId(d.Id())
	if err != nil {
		return fmt.Errorf("Error parse Redshift Cluster Snapshot Schedule Association ID %s: %s", d.Id(), err)
	}

	_, err = conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
		ClusterIdentifier:    aws.String(clusterIdentifier),
		ScheduleIdentifier:   aws.String(scheduleIdentifier),
		DisassociateSchedule: aws.Bool(true),
	})

	if isAWSErr(err, redshift.ErrCodeClusterNotFoundFault, "") {
		log.Printf("[WARN] Redshift Snapshot Cluster (%s) not found, removing from state", clusterIdentifier)
		d.SetId("")
		return nil
	}
	if isAWSErr(err, redshift.ErrCodeSnapshotScheduleNotFoundFault, "") {
		log.Printf("[WARN] Redshift Snapshot Schedule (%s) not found, removing from state", scheduleIdentifier)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error disassociate Redshift Cluster (%s) and Snapshot Schedule (%s) Association: %s", clusterIdentifier, scheduleIdentifier, err)
	}

	if err := waitForRedshiftSnapshotScheduleAssociationDestroy(conn, 75*time.Minute, clusterIdentifier, scheduleIdentifier); err != nil {
		return err
	}

	return nil
}

func resourceAwsRedshiftSnapshotScheduleAssociationParseId(id string) (clusterIdentifier, scheduleIdentifier string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err = fmt.Errorf("aws_redshift_snapshot_schedule_association id must be of the form <ClusterIdentifier>/<ScheduleIdentifier>")
		return
	}

	clusterIdentifier = parts[0]
	scheduleIdentifier = parts[1]
	return
}

func resourceAwsRedshiftSnapshotScheduleAssociationStateRefreshFunc(clusterIdentifier, scheduleIdentifier string, conn *redshift.Redshift) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		log.Printf("[INFO] Reading Redshift Cluster (%s) Snapshot Schedule (%s) Association Information", clusterIdentifier, scheduleIdentifier)
		resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
			ClusterIdentifier:  aws.String(clusterIdentifier),
			ScheduleIdentifier: aws.String(scheduleIdentifier),
		})
		if isAWSErr(err, redshift.ErrCodeClusterNotFoundFault, "") {
			return 42, "destroyed", nil
		}
		if isAWSErr(err, redshift.ErrCodeSnapshotScheduleNotFoundFault, "") {
			return 42, "destroyed", nil
		}
		if err != nil {
			log.Printf("[WARN] Error on retrieving Redshift Cluster (%s) Snapshot Schedule (%s) Association when waiting: %s", clusterIdentifier, scheduleIdentifier, err)
			return nil, "", err
		}

		var rcas *redshift.ClusterAssociatedToSchedule

		for _, s := range resp.SnapshotSchedules {
			if aws.StringValue(s.ScheduleIdentifier) == scheduleIdentifier {
				for _, c := range s.AssociatedClusters {
					if aws.StringValue(c.ClusterIdentifier) == clusterIdentifier {
						rcas = c
					}
				}
			}
		}

		if rcas == nil {
			return 42, "destroyed", nil
		}

		if rcas.ScheduleAssociationState != nil {
			log.Printf("[DEBUG] Redshift Cluster (%s) Snapshot Schedule (%s) Association status: %s", clusterIdentifier, scheduleIdentifier, aws.StringValue(rcas.ScheduleAssociationState))
		}

		return rcas, aws.StringValue(rcas.ScheduleAssociationState), nil
	}
}

func waitForRedshiftSnapshotScheduleAssociationActive(conn *redshift.Redshift, timeout time.Duration, clusterIdentifier, scheduleIdentifier string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying},
		Target:     []string{redshift.ScheduleStateActive},
		Refresh:    resourceAwsRedshiftSnapshotScheduleAssociationStateRefreshFunc(clusterIdentifier, scheduleIdentifier, conn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for Redshift Cluster (%s) and  Snapshot Schedule (%s) Association state to be \"ACTIVE\": %s", clusterIdentifier, scheduleIdentifier, err)
	}

	return nil
}

func waitForRedshiftSnapshotScheduleAssociationDestroy(conn *redshift.Redshift, timeout time.Duration, clusterIdentifier, scheduleIdentifier string) error {

	stateConf := &resource.StateChangeConf{
		Pending:    []string{redshift.ScheduleStateModifying, redshift.ScheduleStateActive},
		Target:     []string{"destroyed"},
		Refresh:    resourceAwsRedshiftSnapshotScheduleAssociationStateRefreshFunc(clusterIdentifier, scheduleIdentifier, conn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for Redshift Cluster (%s) and  Snapshot Schedule (%s) Association state to be \"destroyed\": %s", clusterIdentifier, scheduleIdentifier, err)
	}

	return nil
}
