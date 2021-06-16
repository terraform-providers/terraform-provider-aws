package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfkinesisanalyticsv2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsKinesisAnalyticsV2ApplicationSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisAnalyticsV2ApplicationSnapshotCreate,
		Read:   resourceAwsKinesisAnalyticsV2ApplicationSnapshotRead,
		Delete: resourceAwsKinesisAnalyticsV2ApplicationSnapshotDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},

			"application_version_id": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"snapshot_creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"snapshot_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`), "must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
		},
	}
}

func resourceAwsKinesisAnalyticsV2ApplicationSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn
	applicationName := d.Get("application_name").(string)
	snapshotName := d.Get("snapshot_name").(string)

	input := &kinesisanalyticsv2.CreateApplicationSnapshotInput{
		ApplicationName: aws.String(applicationName),
		SnapshotName:    aws.String(snapshotName),
	}

	log.Printf("[DEBUG] Creating Kinesis Analytics v2 Application Snapshot: %s", input)

	_, err := conn.CreateApplicationSnapshot(input)

	if err != nil {
		return fmt.Errorf("error creating Kinesis Analytics v2 Application Snapshot (%s/%s): %w", applicationName, snapshotName, err)
	}

	d.SetId(tfkinesisanalyticsv2.ApplicationSnapshotCreateID(applicationName, snapshotName))

	_, err = waiter.SnapshotCreated(conn, applicationName, snapshotName)

	if err != nil {
		return fmt.Errorf("error waiting for Kinesis Analytics v2 Application Snapshot (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsKinesisAnalyticsV2ApplicationSnapshotRead(d, meta)
}

func resourceAwsKinesisAnalyticsV2ApplicationSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn

	applicationName, snapshotName, err := tfkinesisanalyticsv2.ApplicationSnapshotParseID(d.Id())

	if err != nil {
		return err
	}

	snapshot, err := finder.SnapshotDetailsByApplicationAndSnapshotNames(conn, applicationName, snapshotName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kinesis Analytics v2 Application Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Kinesis Analytics v2 Application Snapshot (%s): %w", d.Id(), err)
	}

	d.Set("application_name", applicationName)
	d.Set("application_version_id", snapshot.ApplicationVersionId)
	d.Set("snapshot_creation_timestamp", aws.TimeValue(snapshot.SnapshotCreationTimestamp).Format(time.RFC3339))
	d.Set("snapshot_name", snapshotName)

	return nil
}

func resourceAwsKinesisAnalyticsV2ApplicationSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisanalyticsv2conn

	applicationName, snapshotName, err := tfkinesisanalyticsv2.ApplicationSnapshotParseID(d.Id())

	if err != nil {
		return err
	}

	snapshotCreationTimestamp, err := time.Parse(time.RFC3339, d.Get("snapshot_creation_timestamp").(string))
	if err != nil {
		return fmt.Errorf("error parsing snapshot_creation_timestamp: %w", err)
	}

	log.Printf("[DEBUG] Deleting Kinesis Analytics v2 Application Snapshot (%s)", d.Id())
	_, err = conn.DeleteApplicationSnapshot(&kinesisanalyticsv2.DeleteApplicationSnapshotInput{
		ApplicationName:           aws.String(applicationName),
		SnapshotCreationTimestamp: aws.Time(snapshotCreationTimestamp),
		SnapshotName:              aws.String(snapshotName),
	})

	if tfawserr.ErrCodeEquals(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException) {
		return nil
	}

	if tfawserr.ErrMessageContains(err, kinesisanalyticsv2.ErrCodeInvalidArgumentException, "does not exist") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Kinesis Analytics v2 Application Snapshot (%s): %w", d.Id(), err)
	}

	_, err = waiter.SnapshotDeleted(conn, applicationName, snapshotName)

	if err != nil {
		return fmt.Errorf("error waiting for Kinesis Analytics v2 Application Snapshot (%s) deletion: %w", d.Id(), err)
	}

	return nil
}
