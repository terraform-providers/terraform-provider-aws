package aws

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	AWSAMIRetryTimeout       = 40 * time.Minute
	AWSAMIDeleteRetryTimeout = 90 * time.Minute
	AWSAMIRetryDelay         = 5 * time.Second
	AWSAMIRetryMinTimeout    = 3 * time.Second
)

func resourceAwsAmi() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmiCreate,
		// The Read, Update and Delete operations are shared with aws_ami_copy
		// and aws_ami_from_instance, since they differ only in how the image
		// is created.
		Read:   resourceAwsAmiRead,
		Update: resourceAwsAmiUpdate,
		Delete: resourceAwsAmiDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(AWSAMIRetryTimeout),
			Update: schema.DefaultTimeout(AWSAMIRetryTimeout),
			Delete: schema.DefaultTimeout(AWSAMIDeleteRetryTimeout),
		},

		Schema: map[string]*schema.Schema{
			"image_location": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"architecture": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.ArchitectureTypeX8664,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.ArchitectureTypeX8664,
					ec2.ArchitectureValuesI386,
					ec2.ArchitectureValuesArm64,
				}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			// The following block device attributes intentionally mimick the
			// corresponding attributes on aws_instance, since they have the
			// same meaning.
			// However, we don't use root_block_device here because the constraint
			// on which root device attributes can be overridden for an instance to
			// not apply when registering an AMI.
			"ebs_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
							ForceNew: true,
						},

						"device_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"encrypted": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},

						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},

						"snapshot_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},

						"throughput": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"volume_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"volume_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      ec2.VolumeTypeStandard,
							ValidateFunc: validation.StringInSlice(ec2.VolumeType_Values(), false),
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["snapshot_id"].(string)))
					return hashcode.String(buf.String())
				},
			},
			"ena_support": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"virtual_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["virtual_name"].(string)))
					return hashcode.String(buf.String())
				},
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			// Not a public attribute; used to let the aws_ami_copy and aws_ami_from_instance
			// resources record that they implicitly created new EBS snapshots that we should
			// now manage. Not set by aws_ami, since the snapshots used there are presumed to
			// be independently managed.
			"manage_ebs_snapshots": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ramdisk_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"root_device_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"root_snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sriov_net_support": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "simple",
			},
			"tags": tagsSchema(),
			"virtualization_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.VirtualizationTypeParavirtual,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.VirtualizationTypeParavirtual,
					ec2.VirtualizationTypeHvm,
				}, false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAmiCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).ec2conn

	req := &ec2.RegisterImageInput{
		Name:               aws.String(d.Get("name").(string)),
		Description:        aws.String(d.Get("description").(string)),
		Architecture:       aws.String(d.Get("architecture").(string)),
		ImageLocation:      aws.String(d.Get("image_location").(string)),
		RootDeviceName:     aws.String(d.Get("root_device_name").(string)),
		SriovNetSupport:    aws.String(d.Get("sriov_net_support").(string)),
		VirtualizationType: aws.String(d.Get("virtualization_type").(string)),
		EnaSupport:         aws.Bool(d.Get("ena_support").(bool)),
	}

	if kernelId := d.Get("kernel_id").(string); kernelId != "" {
		req.KernelId = aws.String(kernelId)
	}
	if ramdiskId := d.Get("ramdisk_id").(string); ramdiskId != "" {
		req.RamdiskId = aws.String(ramdiskId)
	}

	if v, ok := d.GetOk("ebs_block_device"); ok && v.(*schema.Set).Len() > 0 {
		for _, tfMapRaw := range v.(*schema.Set).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			var encrypted bool

			if v, ok := tfMap["encrypted"].(bool); ok {
				encrypted = v
			}

			var snapshot string

			if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
				snapshot = v
			}

			if snapshot != "" && encrypted {
				return errors.New("can't set both 'snapshot_id' and 'encrypted'")
			}
		}

		req.BlockDeviceMappings = expandEc2BlockDeviceMappingsForAmiEbsBlockDevice(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ephemeral_block_device"); ok && v.(*schema.Set).Len() > 0 {
		req.BlockDeviceMappings = append(req.BlockDeviceMappings, expandEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(v.(*schema.Set).List())...)
	}

	res, err := client.RegisterImage(req)
	if err != nil {
		return err
	}

	id := *res.ImageId
	d.SetId(id)

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		if err := keyvaluetags.Ec2CreateTags(client, id, v); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	_, err = resourceAwsAmiWaitForAvailable(d.Timeout(schema.TimeoutCreate), id, client)
	if err != nil {
		return err
	}

	return resourceAwsAmiRead(d, meta)
}

func resourceAwsAmiRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	id := d.Id()

	req := &ec2.DescribeImagesInput{
		ImageIds: []*string{aws.String(id)},
	}

	var res *ec2.DescribeImagesOutput
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		res, err = client.DescribeImages(req)
		if err != nil {
			if isAWSErr(err, "InvalidAMIID.NotFound", "") {
				if d.IsNewResource() {
					return resource.RetryableError(err)
				}
				log.Printf("[WARN] AMI (%s) not found, removing from state", d.Id())
				d.SetId("")
				return nil
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		res, err = client.DescribeImages(req)
	}
	if err != nil {
		return fmt.Errorf("Unable to find AMI after retries: %s", err)
	}

	if len(res.Images) != 1 {
		log.Printf("[WARN] AMI (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	image := res.Images[0]
	state := *image.State

	if state == ec2.ImageStatePending {
		// This could happen if a user manually adds an image we didn't create
		// to the state. We'll wait for the image to become available
		// before we continue. We should never take this branch in normal
		// circumstances since we would've waited for availability during
		// the "Create" step.
		image, err = resourceAwsAmiWaitForAvailable(d.Timeout(schema.TimeoutCreate), id, client)
		if err != nil {
			return err
		}
		state = *image.State
	}

	if state == ec2.ImageStateDeregistered {
		log.Printf("[WARN] AMI (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if state != ec2.ImageStateAvailable {
		return fmt.Errorf("AMI has become %s", state)
	}

	d.Set("name", image.Name)
	d.Set("description", image.Description)
	d.Set("image_location", image.ImageLocation)
	d.Set("architecture", image.Architecture)
	d.Set("kernel_id", image.KernelId)
	d.Set("ramdisk_id", image.RamdiskId)
	d.Set("root_device_name", image.RootDeviceName)
	d.Set("root_snapshot_id", amiRootSnapshotId(image))
	d.Set("sriov_net_support", image.SriovNetSupport)
	d.Set("virtualization_type", image.VirtualizationType)
	d.Set("ena_support", image.EnaSupport)

	imageArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("image/%s", d.Id()),
		Service:   "ec2",
	}.String()

	d.Set("arn", imageArn)

	if err := d.Set("ebs_block_device", flattenEc2BlockDeviceMappingsForAmiEbsBlockDevice(image.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("error setting ebs_block_device: %w", err)
	}

	if err := d.Set("ephemeral_block_device", flattenEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(image.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("error setting ephemeral_block_device: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(image.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsAmiUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).ec2conn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.Ec2UpdateTags(client, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating AMI (%s) tags: %s", d.Id(), err)
		}
	}

	if d.Get("description").(string) != "" {
		_, err := client.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
			ImageId: aws.String(d.Id()),
			Description: &ec2.AttributeValue{
				Value: aws.String(d.Get("description").(string)),
			},
		})
		if err != nil {
			return err
		}
	}

	return resourceAwsAmiRead(d, meta)
}

func resourceAwsAmiDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).ec2conn

	req := &ec2.DeregisterImageInput{
		ImageId: aws.String(d.Id()),
	}

	_, err := client.DeregisterImage(req)
	if err != nil {
		return err
	}

	// If we're managing the EBS snapshots then we need to delete those too.
	if d.Get("manage_ebs_snapshots").(bool) {
		errs := map[string]error{}
		ebsBlockDevsSet := d.Get("ebs_block_device").(*schema.Set)
		req := &ec2.DeleteSnapshotInput{}
		for _, ebsBlockDevI := range ebsBlockDevsSet.List() {
			ebsBlockDev := ebsBlockDevI.(map[string]interface{})
			snapshotId := ebsBlockDev["snapshot_id"].(string)
			if snapshotId != "" {
				req.SnapshotId = aws.String(snapshotId)
				_, err := client.DeleteSnapshot(req)
				if err != nil {
					errs[snapshotId] = err
				}
			}
		}

		if len(errs) > 0 {
			errParts := []string{"Errors while deleting associated EBS snapshots:"}
			for snapshotId, err := range errs {
				errParts = append(errParts, fmt.Sprintf("%s: %s", snapshotId, err))
			}
			errParts = append(errParts, "These are no longer managed by Terraform and must be deleted manually.")
			return errors.New(strings.Join(errParts, "\n"))
		}
	}

	// Verify that the image is actually removed, if not we need to wait for it to be removed
	if err := resourceAwsAmiWaitForDestroy(d.Timeout(schema.TimeoutDelete), d.Id(), client); err != nil {
		return err
	}

	return nil
}

func AMIStateRefreshFunc(client *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		emptyResp := &ec2.DescribeImagesOutput{}

		resp, err := client.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{aws.String(id)}})
		if err != nil {
			if isAWSErr(err, "InvalidAMIID.NotFound", "") {
				return emptyResp, "destroyed", nil
			} else if resp != nil && len(resp.Images) == 0 {
				return emptyResp, "destroyed", nil
			} else {
				return emptyResp, "", fmt.Errorf("Error on refresh: %+v", err)
			}
		}

		if resp == nil || resp.Images == nil || len(resp.Images) == 0 {
			return emptyResp, "destroyed", nil
		}

		// AMI is valid, so return it's state
		return resp.Images[0], *resp.Images[0].State, nil
	}
}

func resourceAwsAmiWaitForDestroy(timeout time.Duration, id string, client *ec2.EC2) error {
	log.Printf("Waiting for AMI %s to be deleted...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.ImageStateAvailable, ec2.ImageStatePending, ec2.ImageStateFailed},
		Target:     []string{"destroyed"},
		Refresh:    AMIStateRefreshFunc(client, id),
		Timeout:    timeout,
		Delay:      AWSAMIRetryDelay,
		MinTimeout: AWSAMIRetryMinTimeout,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for AMI (%s) to be deleted: %v", id, err)
	}

	return nil
}

func resourceAwsAmiWaitForAvailable(timeout time.Duration, id string, client *ec2.EC2) (*ec2.Image, error) {
	log.Printf("Waiting for AMI %s to become available...", id)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.ImageStatePending},
		Target:     []string{ec2.ImageStateAvailable},
		Refresh:    AMIStateRefreshFunc(client, id),
		Timeout:    timeout,
		Delay:      AWSAMIRetryDelay,
		MinTimeout: AWSAMIRetryMinTimeout,
	}

	info, err := stateConf.WaitForState()
	if err != nil {
		return nil, fmt.Errorf("Error waiting for AMI (%s) to be ready: %v", id, err)
	}
	return info.(*ec2.Image), nil
}

func expandEc2BlockDeviceMappingForAmiEbsBlockDevice(tfMap map[string]interface{}) *ec2.BlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.BlockDeviceMapping{
		Ebs: &ec2.EbsBlockDevice{},
	}

	if v, ok := tfMap["delete_on_termination"].(bool); ok {
		apiObject.Ebs.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Ebs.Iops = aws.Int64(int64(v))
	}

	// "Parameter encrypted is invalid. You cannot specify the encrypted flag if specifying a snapshot id in a block device mapping."
	if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
		apiObject.Ebs.SnapshotId = aws.String(v)
	} else if v, ok := tfMap["encrypted"].(bool); ok {
		apiObject.Ebs.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap["throughput"].(int); ok && v != 0 {
		apiObject.Ebs.Throughput = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_size"].(int); ok && v != 0 {
		apiObject.Ebs.VolumeSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_type"].(string); ok && v != "" {
		apiObject.Ebs.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandEc2BlockDeviceMappingsForAmiEbsBlockDevice(tfList []interface{}) []*ec2.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2BlockDeviceMappingForAmiEbsBlockDevice(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEc2BlockDeviceMappingForAmiEbsBlockDevice(apiObject *ec2.BlockDeviceMapping) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	if apiObject.Ebs == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Ebs.DeleteOnTermination; v != nil {
		tfMap["delete_on_termination"] = aws.BoolValue(v)
	}

	if v := apiObject.DeviceName; v != nil {
		tfMap["device_name"] = aws.StringValue(v)
	}

	if v := apiObject.Ebs.Encrypted; v != nil {
		tfMap["encrypted"] = aws.BoolValue(v)
	}

	if v := apiObject.Ebs.Iops; v != nil {
		tfMap["iops"] = aws.Int64Value(v)
	}

	if v := apiObject.Ebs.SnapshotId; v != nil {
		tfMap["snapshot_id"] = aws.StringValue(v)
	}

	if v := apiObject.Ebs.Throughput; v != nil {
		tfMap["throughput"] = aws.Int64Value(v)
	}

	if v := apiObject.Ebs.VolumeSize; v != nil {
		tfMap["volume_size"] = aws.Int64Value(v)
	}

	if v := apiObject.Ebs.VolumeType; v != nil {
		tfMap["volume_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2BlockDeviceMappingsForAmiEbsBlockDevice(apiObjects []*ec2.BlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if apiObject.Ebs == nil {
			continue
		}

		tfList = append(tfList, flattenEc2BlockDeviceMappingForAmiEbsBlockDevice(apiObject))
	}

	return tfList
}

func expandEc2BlockDeviceMappingForAmiEphemeralBlockDevice(tfMap map[string]interface{}) *ec2.BlockDeviceMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.BlockDeviceMapping{}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["virtual_name"].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(tfList []interface{}) []*ec2.BlockDeviceMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.BlockDeviceMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEc2BlockDeviceMappingForAmiEphemeralBlockDevice(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEc2BlockDeviceMappingForAmiEphemeralBlockDevice(apiObject *ec2.BlockDeviceMapping) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap["device_name"] = aws.StringValue(v)
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap["virtual_name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEc2BlockDeviceMappingsForAmiEphemeralBlockDevice(apiObjects []*ec2.BlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if apiObject.Ebs != nil {
			continue
		}

		tfList = append(tfList, flattenEc2BlockDeviceMappingForAmiEphemeralBlockDevice(apiObject))
	}

	return tfList
}
