package aws

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

// the S3 API can be very inconsistent with the status of Versioning on a bucket
// so we require at least this many confirmations that the value is actually set
// before we move on
const s3BucketVersioningConfirmationsRequired = 10

func resourceAwsS3BucketVersioning() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketVersioningCreate,
		Read:   resourceAwsS3BucketVersioningRead,
		Update: resourceAwsS3BucketVersioningUpdate,
		Delete: resourceAwsS3BucketVersioningDelete,
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func s3VersioningStatusToBool(status *string) bool {
	if status == nil {
		return false
	}

	if *status == s3.BucketVersioningStatusEnabled {
		return true
	}

	return false
}

// updates s3 bucket versioning and waits for the change to be confirmed
func putS3BucketVersioning(s3conn *s3.S3, bucket string, enabled bool) error {
	vc := &s3.VersioningConfiguration{}
	if enabled {
		vc.SetStatus(s3.BucketVersioningStatusEnabled)
	} else {
		vc.SetStatus(s3.BucketVersioningStatusSuspended)
	}

	_, err := s3conn.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: vc,
	})
	if err != nil {
		return err
	}

	// wait up to 30 seconds for the change to be reflected
	updateConfirmations := 0
	for i := 0; i < 120; i++ {
		bucketVersioningStatus, err := s3conn.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: aws.String(bucket),
		})
		if err != nil {
			return fmt.Errorf("Error getting versioning config for %s: %s", bucket, err)
		}

		if bucketVersioningStatus != nil && s3VersioningStatusToBool(bucketVersioningStatus.Status) == enabled {
			updateConfirmations++
			if updateConfirmations >= s3BucketVersioningConfirmationsRequired {
				return nil
			}
		}

		time.Sleep(250 * time.Millisecond)
	}

	return fmt.Errorf("Timed out waiting for confirmation")
}

func resourceAwsS3BucketVersioningCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)
	versioningEnabled := d.Get("enabled").(bool)

	_, err := retryOnAwsCode("NoSuchBucket", func() (interface{}, error) {
		return nil, putS3BucketVersioning(conn, bucket, versioningEnabled)
	})
	if err != nil {
		return fmt.Errorf("Error putting bucket versioning for %s: %s", bucket, err)
	}

	d.SetId(resource.UniqueId())

	return nil
}

func resourceAwsS3BucketVersioningRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn

	bucket := d.Get("bucket").(string)

	vo, err := conn.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("Error getting bucket versioning for %s: %s", bucket, err)
	}

	var versioningEnabled bool
	if vo == nil {
		versioningEnabled = false
	} else {
		versioningEnabled = s3VersioningStatusToBool(vo.Status)
	}

	if err := d.Set("enabled", versioningEnabled); err != nil {
		return fmt.Errorf("[WARN] Error setting versioning status from S3 (%s / %t), error: %s", bucket, versioningEnabled, err)
	}

	return nil
}

func resourceAwsS3BucketVersioningUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn
	bucket := d.Get("bucket").(string)

	if d.HasChange("enabled") {
		_, versioningEnabled := d.GetChange("enabled")

		err := putS3BucketVersioning(conn, bucket, versioningEnabled.(bool))
		if err != nil {
			return fmt.Errorf("Error setting versioning = %t on %s: %s", versioningEnabled, bucket, err)
		}
	}

	return nil
}

func resourceAwsS3BucketVersioningDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).s3conn
	bucket := d.Get("bucket").(string)

	// disable versioning if it was enabled
	if d.Get("enabled").(bool) {
		err := putS3BucketVersioning(conn, bucket, false)
		if err != nil {
			return fmt.Errorf("Error deleting aws_s3_bucket_versioning: could not disable versioning on bucket %s: %s", bucket, err)
		}
	}

	d.SetId("")

	return nil
}
