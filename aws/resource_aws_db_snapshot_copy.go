package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsDbSnapshotCopy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDbSnapshotCopyCreate,
		Read:   resourceAwsDbSnapshotCopyRead,
		Delete: resourceAwsDbSnapshotCopyDelete,

		Schema: map[string]*schema.Schema{
			"copy_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"destination_region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"presigned_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"option_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_db_snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_region": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
			"target_db_snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDbSnapshotCopyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RdsTags()

	request := &rds.CopyDBSnapshotInput{
		SourceRegion:               aws.String(d.Get("source_region").(string)),
		SourceDBSnapshotIdentifier: aws.String(d.Get("source_db_snapshot_identifier").(string)),
		TargetDBSnapshotIdentifier: aws.String(d.Get("target_db_snapshot_identifier").(string)),
		Tags:                       tags,
	}
	if v, ok := d.GetOk("copy_tags"); ok {
		request.CopyTags = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("kms_key_id"); ok {
		request.KmsKeyId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("option_group_name"); ok {
		request.OptionGroupName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("destination_region"); ok {
		request.DestinationRegion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("presigned_url"); ok {
		request.PreSignedUrl = aws.String(v.(string))
	}

	res, err := conn.CopyDBSnapshot(request)
	if err != nil {
		return err
	}

	d.SetId(*res.DBSnapshot.DBSnapshotIdentifier)

	err = resourceAwsDbSnapshotCopyWaitForAvailable(d.Id(), conn)
	if err != nil {
		return err
	}

	return resourceAwsDbSnapshotCopyRead(d, meta)
}

func resourceAwsDbSnapshotCopyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	req := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: aws.String(d.Id()),
	}
	res, err := conn.DescribeDBSnapshots(req)
	if isAWSErr(err, "InvalidDBSnapshot.NotFound", "") {
		log.Printf("Snapshot %q Not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	snapshot := res.DBSnapshots[0]

	arn := aws.StringValue(snapshot.DBSnapshotArn)
	d.Set("engine", snapshot.Engine)
	d.Set("engine_version", snapshot.EngineVersion)
	d.Set("encrypted", snapshot.Encrypted)
	d.Set("snapshot_create_type", snapshot.SnapshotCreateTime)
	d.Set("snapshot_identifier", snapshot.DBSnapshotIdentifier)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("storage_type", snapshot.StorageType)

	tags, err := keyvaluetags.RdsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for RDS DB Snapshot (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDbSnapshotCopyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	input := &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(d.Id()),
	}
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDBSnapshot(input)
		if err == nil {
			return nil
		}

		if isAWSErr(err, "SnapshotInUse", "") {
			return resource.RetryableError(fmt.Errorf("RDS SnapshotInUse - trying again while it detaches"))
		}

		if isAWSErr(err, "InvalidSnapshot.NotFound", "") {
			return nil
		}

		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteDBSnapshot(input)
		if isAWSErr(err, "InvalidDBSnapshot.NotFound", "") {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error deleting RDS snapshot copy: %s", err)
	}
	return nil
}

func resourceAwsDbSnapshotCopyWaitForAvailable(id string, conn *rds.RDS) error {
	log.Printf("Waiting for Snapshot %s to become available...", id)

	req := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: aws.String(id),
	}
	err := conn.WaitUntilDBSnapshotAvailable(req)
	return err
}
