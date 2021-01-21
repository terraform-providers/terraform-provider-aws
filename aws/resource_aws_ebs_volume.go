package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

		CustomizeDiff: resourceAwsEbsVolumeCustomizeDiff,

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
			"iops": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"multi_attach_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"size", "snapshot_id"},
			},
			"snapshot_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"size", "snapshot_id"},
			},
			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tagsSchema(),
			"throughput": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(125, 1000),
			},
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
	if value, ok := d.GetOk("iops"); ok {
		request.Iops = aws.Int64(int64(value.(int)))
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
	if value, ok := d.GetOk("multi_attach_enabled"); ok {
		request.MultiAttachEnabled = aws.Bool(value.(bool))
	}
	if value, ok := d.GetOk("outpost_arn"); ok {
		request.OutpostArn = aws.String(value.(string))
	}
	if value, ok := d.GetOk("throughput"); ok {
		request.Throughput = aws.Int64(int64(value.(int)))
	}
	if value, ok := d.GetOk("type"); ok {
		request.VolumeType = aws.String(value.(string))
	}

	log.Printf("[DEBUG] EBS Volume create opts: %s", request)
	result, err := conn.CreateVolume(request)
	if err != nil {
		return fmt.Errorf("Error creating EC2 volume: %s", err)
	}

	log.Println("[DEBUG] Waiting for Volume to become available")

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeStateCreating},
		Target:     []string{ec2.VolumeStateAvailable},
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

	d.SetId(aws.StringValue(result.VolumeId))

	return resourceAwsEbsVolumeRead(d, meta)
}

func resourceAWSEbsVolumeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	if d.HasChangesExcept("tags") {
		params := &ec2.ModifyVolumeInput{
			VolumeId: aws.String(d.Id()),
		}

		if d.HasChange("size") {
			params.Size = aws.Int64(int64(d.Get("size").(int)))
		}

		if d.HasChange("type") {
			params.VolumeType = aws.String(d.Get("type").(string))
		}

		if d.HasChange("iops") {
			params.Iops = aws.Int64(int64(d.Get("iops").(int)))
		}

		// "If no throughput value is specified, the existing value is retained."
		// Not currently correct, so always specify any non-zero throughput value.
		if v := d.Get("throughput").(int); v > 0 {
			params.Throughput = aws.Int64(int64(v))
		}

		result, err := conn.ModifyVolume(params)
		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{ec2.VolumeStateCreating, ec2.VolumeModificationStateModifying},
			Target:     []string{ec2.VolumeStateAvailable, ec2.VolumeStateInUse},
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
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &ec2.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(d.Id())},
	}

	response, err := conn.DescribeVolumes(request)
	if err != nil {
		if isAWSErr(err, "InvalidVolume.NotFound", "") {
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
	d.Set("availability_zone", volume.AvailabilityZone)
	d.Set("encrypted", volume.Encrypted)
	d.Set("iops", volume.Iops)
	d.Set("kms_key_id", volume.KmsKeyId)
	d.Set("size", volume.Size)
	d.Set("snapshot_id", volume.SnapshotId)
	d.Set("outpost_arn", volume.OutpostArn)
	d.Set("multi_attach_enabled", volume.MultiAttachEnabled)
	d.Set("throughput", volume.Throughput)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(volume.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("type", volume.VolumeType)

	return nil
}

func resourceAwsEbsVolumeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

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

func resourceAwsEbsVolumeCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	iops := diff.Get("iops").(int)
	multiAttachEnabled := diff.Get("multi_attach_enabled").(bool)
	throughput := diff.Get("throughput").(int)
	volumeType := diff.Get("type").(string)

	if diff.Id() == "" {
		// Create.

		// Iops is required for io1 and io2 volumes.
		// The default for gp3 volumes is 3,000 IOPS.
		// This parameter is not supported for gp2, st1, sc1, or standard volumes.
		// Hard validation in place to return an error if IOPs are provided
		// for an unsupported storage type.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12667
		switch volumeType {
		case ec2.VolumeTypeIo1, ec2.VolumeTypeIo2:
			if iops == 0 {
				return fmt.Errorf("'iops' must be set when 'type' is '%s'", volumeType)
			}

		case ec2.VolumeTypeGp3:

		default:
			if iops != 0 {
				return fmt.Errorf("'iops' must not be set when 'type' is '%s'", volumeType)
			}
		}

		// MultiAttachEnabled is supported with io1 volumes only.
		if multiAttachEnabled && volumeType != ec2.VolumeTypeIo1 {
			return fmt.Errorf("'multi_attach_enabled' must not be set when 'type' is '%s'", volumeType)
		}

		// Throughput is valid only for gp3 volumes.
		if throughput > 0 && volumeType != ec2.VolumeTypeGp3 {
			return fmt.Errorf("'throughput' must not be set when 'type' is '%s'", volumeType)
		}
	} else {
		// Update.

		// Setting 'iops = 0' is a no-op if the volume type does not require Iops to be specified.
		if diff.HasChange("iops") && volumeType != ec2.VolumeTypeIo1 && volumeType != ec2.VolumeTypeIo2 && iops == 0 {
			return diff.Clear("iops")
		}
	}

	return nil
}
