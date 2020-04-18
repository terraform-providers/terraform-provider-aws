package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsEbsVolume() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEbsVolumeCreate,
		Read:   resourceAwsEbsVolumeRead,
		Update: resourceAWSEbsVolumeUpdate,
		Delete: resourceAwsEbsVolumeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"iops": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get("type").(string) != ec2.VolumeTypeIo1 && new == "0"
				},
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsEbsVolumeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	request := &ec2.CreateVolumeInput{
		AvailabilityZone:  aws.String(d.Get("availability_zone").(string)),
		TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeVolume),
	}
	if value, ok := d.GetOk("encrypted"); ok {
		request.Encrypted = aws.Bool(value.(bool))
	}
	if value, ok := d.GetOk("kms_key_id"); ok {
		request.KmsKeyId = aws.String(value.(string))
	}
	if value, ok := d.GetOk("size"); ok {
		request.Size = aws.Int64(int64(value.(int)))
	}
	if value, ok := d.GetOk("snapshot_id"); ok {
		request.SnapshotId = aws.String(value.(string))
	}

	// IOPs are only valid, and required for, storage type io1. The current minimu
	// is 100. Instead of a hard validation we we only apply the IOPs to the
	// request if the type is io1, and log a warning otherwise. This allows users
	// to "disable" iops. See https://github.com/hashicorp/terraform/pull/4146
	var t string
	if value, ok := d.GetOk("type"); ok {
		t = value.(string)
		request.VolumeType = aws.String(t)
	}

	iops := d.Get("iops").(int)
	if t != "io1" && iops > 0 {
		log.Printf("[WARN] IOPs is only valid for storate type io1 for EBS Volumes")
	} else if t == "io1" {
		// We add the iops value without validating it's size, to allow AWS to
		// enforce a size requirement (currently 100)
		request.Iops = aws.Int64(int64(iops))
	}

	log.Printf(
		"[DEBUG] EBS Volume create opts: %s", request)
	result, err := conn.CreateVolume(request)
	if err != nil {
		return fmt.Errorf("Error creating EC2 volume: %s", err)
	}

	log.Println("[DEBUG] Waiting for Volume to become available")

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available"},
		Refresh:    volumeStateRefreshFunc(conn, *result.VolumeId),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for Volume (%s) to become available: %s",
			*result.VolumeId, err)
	}

	d.SetId(*result.VolumeId)

	return resourceAwsEbsVolumeRead(d, meta)
}

func resourceAWSEbsVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	requestUpdate := false
	params := &ec2.ModifyVolumeInput{
		VolumeId: aws.String(d.Id()),
	}

	if d.HasChange("size") {
		requestUpdate = true
		params.Size = aws.Int64(int64(d.Get("size").(int)))
	}

	if d.HasChange("type") {
		requestUpdate = true
		params.VolumeType = aws.String(d.Get("type").(string))
	}

	if d.HasChange("iops") {
		requestUpdate = true
		params.Iops = aws.Int64(int64(d.Get("iops").(int)))
	}

	if requestUpdate {
		result, err := conn.ModifyVolume(params)
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"creating", "modifying"},
			Target:     []string{"available", "in-use"},
			Refresh:    volumeStateRefreshFunc(conn, *result.VolumeModification.VolumeId),
			Timeout:    5 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf(
				"Error waiting for Volume (%s) to become available: %s",
				*result.VolumeModification.VolumeId, err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsEbsVolumeRead(d, meta)
}

// volumeStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// a the state of a Volume. Returns successfully when volume is available
func volumeStateRefreshFunc(conn *ec2.EC2, volumeID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeVolumes(&ec2.DescribeVolumesInput{
			VolumeIds: []*string{aws.String(volumeID)},
		})

		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok {
				// Set this to nil as if we didn't find anything.
				log.Printf("Error on Volume State Refresh: message: \"%s\", code:\"%s\"", ec2err.Message(), ec2err.Code())
				resp = nil
				return nil, "", err
			} else {
				log.Printf("Error on Volume State Refresh: %s", err)
				return nil, "", err
			}
		}

		v := resp.Volumes[0]
		return v, *v.State, nil
	}
}

func resourceAwsEbsVolumeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	request := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(d.Id())},
	}

	response, err := conn.DescribeVolumes(request)
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVolume.NotFound" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading EC2 volume %s: %s", d.Id(), err)
	}

	if response == nil || len(response.Volumes) == 0 || response.Volumes[0] == nil {
		return fmt.Errorf("error reading EC2 Volume (%s): empty response", d.Id())
	}

	volume := response.Volumes[0]

	arn := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("volume/%s", d.Id()),
		Service:   "ec2",
	}
	d.Set("arn", arn.String())
	d.Set("availability_zone", aws.StringValue(volume.AvailabilityZone))
	d.Set("encrypted", aws.BoolValue(volume.Encrypted))
	d.Set("iops", aws.Int64Value(volume.Iops))
	d.Set("kms_key_id", aws.StringValue(volume.KmsKeyId))
	d.Set("size", aws.Int64Value(volume.Size))
	d.Set("snapshot_id", aws.StringValue(volume.SnapshotId))

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(volume.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("type", aws.StringValue(volume.VolumeType))

	return nil
}

func resourceAwsEbsVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.Get("final_snapshot").(bool) {
		// Boosted heavily from r-aws_ebs_snapshot when I couldn't figure out
		// how to call that directly from here.

		log.Printf("[INFO] final_snaphot set to true; before destroying, making it so for volume: %s", d.Id())

		request := &ec2.CreateSnapshotInput{
			VolumeId:          aws.String(d.Id()),
			TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeSnapshot),
		}

		var res *ec2.Snapshot
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			res, err = conn.CreateSnapshot(request)

			if isAWSErr(err, "SnapshotCreationPerVolumeRateExceeded", "The maximum per volume CreateSnapshot request rate has been exceeded") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})
		if isResourceTimeoutError(err) {
			res, err = conn.CreateSnapshot(request)
		}
		if err != nil {
			return fmt.Errorf("error creating EC2 EBS Snapshot: %s", err)
		}

		err = func() error {
			log.Printf("Waiting for Snapshot %s to become available...", d.Id())
			input := &ec2.DescribeSnapshotsInput{
				SnapshotIds: []*string{aws.String(*res.SnapshotId)},
			}
			err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
				err := conn.WaitUntilSnapshotCompleted(input)
				if err == nil {
					return nil
				}
				if isAWSErr(err, "ResourceNotReady", "") {
					return resource.RetryableError(fmt.Errorf("EBS CreatingSnapshot - waiting for snapshot to become available"))
				}
				return resource.NonRetryableError(err)
			})
			if isResourceTimeoutError(err) {
				err = conn.WaitUntilSnapshotCompleted(input)
			}
			if err != nil {
				return fmt.Errorf("Error waiting for EBS snapshot to complete: %s", err)
			}
			return nil
		}()

		if err != nil {
			return fmt.Errorf("Error creating snapshot: %s", err)
		}

	}

	input := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(d.Id()),
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteVolume(input)

		if isAWSErr(err, "InvalidVolume.NotFound", "") {
			return nil
		}

		if isAWSErr(err, "VolumeInUse", "") {
			return resource.RetryableError(fmt.Errorf("EBS VolumeInUse - trying again while it detaches"))
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteVolume(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting EBS Volume (%s): %s", d.Id(), err)
	}

	describeInput := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(d.Id())},
	}

	var output *ec2.DescribeVolumesOutput
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error
		output, err = conn.DescribeVolumes(describeInput)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		for _, volume := range output.Volumes {
			if aws.StringValue(volume.VolumeId) == d.Id() {
				state := aws.StringValue(volume.State)

				if state == ec2.VolumeStateDeleting {
					return resource.RetryableError(fmt.Errorf("EBS Volume (%s) still deleting", d.Id()))
				}

				return resource.NonRetryableError(fmt.Errorf("EBS Volume (%s) in unexpected state after deletion: %s", d.Id(), state))
			}
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.DescribeVolumes(describeInput)
	}

	if isAWSErr(err, "InvalidVolume.NotFound", "") {
		return nil
	}

	for _, volume := range output.Volumes {
		if aws.StringValue(volume.VolumeId) == d.Id() {
			return fmt.Errorf("EBS Volume (%s) in unexpected state after deletion: %s", d.Id(), aws.StringValue(volume.State))
		}
	}

	return nil
}
