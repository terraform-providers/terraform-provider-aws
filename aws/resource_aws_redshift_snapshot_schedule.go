package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsRedshiftSnapshotSchedule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRedshiftSnapshotScheduleCreate,
		Read:   resourceAwsRedshiftSnapshotScheduleRead,
		Update: resourceAwsRedshiftSnapshotScheduleUpdate,
		Delete: resourceAwsRedshiftSnapshotScheduleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"identifier_prefix"},
			},
			"identifier_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"definitions": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": tagsSchema(),
		},
	}

}

func resourceAwsRedshiftSnapshotScheduleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RedshiftTags()
	var identifier string
	if v, ok := d.GetOk("identifier"); ok {
		identifier = v.(string)
	} else {
		if v, ok := d.GetOk("identifier_prefix"); ok {
			identifier = resource.PrefixedUniqueId(v.(string))
		} else {
			identifier = resource.UniqueId()
		}
	}
	createOpts := &redshift.CreateSnapshotScheduleInput{
		ScheduleIdentifier:  aws.String(identifier),
		ScheduleDefinitions: expandStringSet(d.Get("definitions").(*schema.Set)),
		Tags:                tags,
	}
	if attr, ok := d.GetOk("description"); ok {
		createOpts.ScheduleDescription = aws.String(attr.(string))
	}

	resp, err := conn.CreateSnapshotSchedule(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Redshift Snapshot Schedule: %s", err)
	}

	d.SetId(aws.StringValue(resp.ScheduleIdentifier))

	return resourceAwsRedshiftSnapshotScheduleRead(d, meta)
}

func resourceAwsRedshiftSnapshotScheduleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	descOpts := &redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(d.Id()),
	}

	resp, err := conn.DescribeSnapshotSchedules(descOpts)
	if err != nil {
		return fmt.Errorf("Error describing Redshift Cluster Snapshot Schedule %s: %s", d.Id(), err)
	}

	if resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) != 1 {
		log.Printf("[WARN] Unable to find Redshift Cluster Snapshot Schedule (%s)", d.Id())
		d.SetId("")
		return nil
	}
	snapshotSchedule := resp.SnapshotSchedules[0]

	d.Set("identifier", snapshotSchedule.ScheduleIdentifier)
	d.Set("description", snapshotSchedule.ScheduleDescription)
	if err := d.Set("definitions", flattenStringList(snapshotSchedule.ScheduleDefinitions)); err != nil {
		return fmt.Errorf("Error setting definitions: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.RedshiftKeyValueTags(snapshotSchedule.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "redshift",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("snapshotschedule:%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceAwsRedshiftSnapshotScheduleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RedshiftUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Snapshot Schedule (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	if d.HasChange("definitions") {
		modifyOpts := &redshift.ModifySnapshotScheduleInput{
			ScheduleIdentifier:  aws.String(d.Id()),
			ScheduleDefinitions: expandStringList(d.Get("definitions").(*schema.Set).List()),
		}
		_, err := conn.ModifySnapshotSchedule(modifyOpts)
		if isAWSErr(err, redshift.ErrCodeSnapshotScheduleNotFoundFault, "") {
			log.Printf("[WARN] Redshift Snapshot Schedule (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error modifying Redshift Snapshot Schedule %s: %s", d.Id(), err)
		}
	}

	return resourceAwsRedshiftSnapshotScheduleRead(d, meta)
}

func resourceAwsRedshiftSnapshotScheduleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	if d.Get("force_destroy").(bool) {
		if err := resourceAwsRedshiftSnapshotScheduleDeleteAllAssociatedClusters(conn, d.Id()); err != nil {
			return err
		}
	}

	_, err := conn.DeleteSnapshotSchedule(&redshift.DeleteSnapshotScheduleInput{
		ScheduleIdentifier: aws.String(d.Id()),
	})
	if isAWSErr(err, redshift.ErrCodeSnapshotScheduleNotFoundFault, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting Redshift Snapshot Schedule %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRedshiftSnapshotScheduleDeleteAllAssociatedClusters(conn *redshift.Redshift, scheduleIdentifier string) error {

	resp, err := conn.DescribeSnapshotSchedules(&redshift.DescribeSnapshotSchedulesInput{
		ScheduleIdentifier: aws.String(scheduleIdentifier),
	})
	if isAWSErr(err, redshift.ErrCodeSnapshotScheduleNotFoundFault, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error describing Redshift Cluster Snapshot Schedule %s: %s", scheduleIdentifier, err)
	}
	if resp.SnapshotSchedules == nil || len(resp.SnapshotSchedules) != 1 {
		log.Printf("[WARN] Unable to find Redshift Cluster Snapshot Schedule (%s)", scheduleIdentifier)
		return nil
	}

	snapshotSchedule := resp.SnapshotSchedules[0]

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		_, err = conn.ModifyClusterSnapshotSchedule(&redshift.ModifyClusterSnapshotScheduleInput{
			ClusterIdentifier:    associatedCluster.ClusterIdentifier,
			ScheduleIdentifier:   aws.String(scheduleIdentifier),
			DisassociateSchedule: aws.Bool(true),
		})

		if isAWSErr(err, redshift.ErrCodeClusterNotFoundFault, "") {
			log.Printf("[WARN] Redshift Snapshot Cluster (%s) not found, removing from state", aws.StringValue(associatedCluster.ClusterIdentifier))
			continue
		}
		if isAWSErr(err, redshift.ErrCodeSnapshotScheduleNotFoundFault, "") {
			log.Printf("[WARN] Redshift Snapshot Schedule (%s) not found, removing from state", scheduleIdentifier)
			continue
		}
		if err != nil {
			return fmt.Errorf("Error disassociate Redshift Cluster (%s) and Snapshot Schedule (%s) Association: %s", aws.StringValue(associatedCluster.ClusterIdentifier), scheduleIdentifier, err)
		}
	}

	for _, associatedCluster := range snapshotSchedule.AssociatedClusters {
		if err := waitForRedshiftSnapshotScheduleAssociationDestroy(conn, 75*time.Minute, aws.StringValue(associatedCluster.ClusterIdentifier), scheduleIdentifier); err != nil {
			return err
		}
	}

	return nil
}
