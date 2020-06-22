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

func resourceAwsRedshiftSnapshotCopyGrant() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRedshiftSnapshotCopyGrantCreate,
		Read:   resourceAwsRedshiftSnapshotCopyGrantRead,
		Update: resourceAwsRedshiftSnapshotCopyGrantUpdate,
		Delete: resourceAwsRedshiftSnapshotCopyGrantDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_copy_grant_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsRedshiftSnapshotCopyGrantCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	grantName := d.Get("snapshot_copy_grant_name").(string)

	input := redshift.CreateSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	if v, ok := d.GetOk("kms_key_id"); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	input.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().RedshiftTags()

	log.Printf("[DEBUG]: Adding new Redshift SnapshotCopyGrant: %s", input)

	var out *redshift.CreateSnapshotCopyGrantOutput
	var err error

	out, err = conn.CreateSnapshotCopyGrant(&input)

	if err != nil {
		return fmt.Errorf("error creating Redshift Snapshot Copy Grant (%s): %s", grantName, err)
	}

	log.Printf("[DEBUG] Created new Redshift SnapshotCopyGrant: %s", *out.SnapshotCopyGrant.SnapshotCopyGrantName)
	d.SetId(grantName)

	err = resource.Retry(3*time.Minute, func() *resource.RetryError {
		var err error
		var grant *redshift.SnapshotCopyGrant
		grant, err = findAwsRedshiftSnapshotCopyGrant(conn, grantName)
		if isAWSErr(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault, "") || grant == nil {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = findAwsRedshiftSnapshotCopyGrant(conn, grantName)
		if err != nil {
			return err
		}
	}

	return resourceAwsRedshiftSnapshotCopyGrantRead(d, meta)
}

func resourceAwsRedshiftSnapshotCopyGrantRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	grantName := d.Id()
	log.Printf("[DEBUG] Looking for grant: %s", grantName)

	grant, err := findAwsRedshiftSnapshotCopyGrant(conn, grantName)
	if isAWSErr(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault, "") || grant == nil {
		log.Printf("[WARN] snapshot copy grant (%s) not found, removing from state", grantName)
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "redshift",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("snapshotcopygrant:%s", grantName),
	}.String()

	d.Set("arn", arn)

	d.Set("kms_key_id", grant.KmsKeyId)
	d.Set("snapshot_copy_grant_name", grant.SnapshotCopyGrantName)
	if err := d.Set("tags", keyvaluetags.RedshiftKeyValueTags(grant.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsRedshiftSnapshotCopyGrantUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.RedshiftUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Redshift Snapshot Copy Grant (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsRedshiftSnapshotCopyGrantRead(d, meta)
}

func resourceAwsRedshiftSnapshotCopyGrantDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).redshiftconn

	grantName := d.Id()

	deleteInput := redshift.DeleteSnapshotCopyGrantInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	log.Printf("[DEBUG] Deleting snapshot copy grant: %s", grantName)
	_, err := conn.DeleteSnapshotCopyGrant(&deleteInput)

	if err != nil {
		if isAWSErr(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault, "") {
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Checking if grant is deleted: %s", grantName)
	err = waitForAwsRedshiftSnapshotCopyGrantToBeDeleted(conn, grantName)

	return err
}

// Used by the tests as well
func waitForAwsRedshiftSnapshotCopyGrantToBeDeleted(conn *redshift.Redshift, grantName string) error {
	err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		var err error
		var grant *redshift.SnapshotCopyGrant
		grant, err = findAwsRedshiftSnapshotCopyGrant(conn, grantName)
		if isAWSErr(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault, "") || grant == nil {
			return nil
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(fmt.Errorf("[DEBUG] Grant still exists while expected to be deleted: %s", grantName))
	})
	if isResourceTimeoutError(err) {
		var grant *redshift.SnapshotCopyGrant
		grant, err = findAwsRedshiftSnapshotCopyGrant(conn, grantName)
		if isAWSErr(err, redshift.ErrCodeSnapshotCopyGrantNotFoundFault, "") || grant == nil {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("Error waiting for snapshot copy grant to be deleted: %s", err)
	}
	return nil
}

func findAwsRedshiftSnapshotCopyGrant(conn *redshift.Redshift, grantName string) (*redshift.SnapshotCopyGrant, error) {

	input := redshift.DescribeSnapshotCopyGrantsInput{
		SnapshotCopyGrantName: aws.String(grantName),
	}

	out, err := conn.DescribeSnapshotCopyGrants(&input)

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.SnapshotCopyGrants) == 0 {
		return nil, nil
	}

	return out.SnapshotCopyGrants[0], nil
}
