package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/efs/waiter"
)

func resourceAwsEfsFileSystem() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEfsFileSystemCreate,
		Read:   resourceAwsEfsFileSystemRead,
		Update: resourceAwsEfsFileSystemUpdate,
		Delete: resourceAwsEfsFileSystemDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"availability_zone_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"creation_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},

			"performance_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(efs.PerformanceMode_Values(), false),
			},

			"encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"kms_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"provisioned_throughput_in_mibps": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"number_of_mount_targets": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),

			"throughput_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      efs.ThroughputModeBursting,
				ValidateFunc: validation.StringInSlice(efs.ThroughputMode_Values(), false),
			},

			"lifecycle_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transition_to_ia": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(efs.TransitionToIARules_Values(), false),
						},
					},
				},
			},
			"size_in_bytes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"value_in_ia": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"value_in_standard": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsEfsFileSystemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	creationToken := ""
	if v, ok := d.GetOk("creation_token"); ok {
		creationToken = v.(string)
	} else {
		creationToken = resource.UniqueId()
	}
	throughputMode := d.Get("throughput_mode").(string)

	createOpts := &efs.CreateFileSystemInput{
		CreationToken:  aws.String(creationToken),
		ThroughputMode: aws.String(throughputMode),
		Tags:           keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EfsTags(),
	}

	if v, ok := d.GetOk("availability_zone_name"); ok {
		createOpts.AvailabilityZoneName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("performance_mode"); ok {
		createOpts.PerformanceMode = aws.String(v.(string))
	}

	if throughputMode == efs.ThroughputModeProvisioned {
		createOpts.ProvisionedThroughputInMibps = aws.Float64(d.Get("provisioned_throughput_in_mibps").(float64))
	}

	encrypted, hasEncrypted := d.GetOk("encrypted")
	kmsKeyId, hasKmsKeyId := d.GetOk("kms_key_id")

	if hasEncrypted {
		createOpts.Encrypted = aws.Bool(encrypted.(bool))
	}

	if hasKmsKeyId {
		createOpts.KmsKeyId = aws.String(kmsKeyId.(string))
	}

	if encrypted == false && hasKmsKeyId {
		return errors.New("encrypted must be set to true when kms_key_id is specified")
	}

	log.Printf("[DEBUG] EFS file system create options: %#v", *createOpts)
	fs, err := conn.CreateFileSystem(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating EFS file system: %w", err)
	}

	d.SetId(aws.StringValue(fs.FileSystemId))
	log.Printf("[INFO] EFS file system ID: %s", d.Id())

	if _, err := waiter.FileSystemAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for EFS file system (%s) to be available: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS file system %q created.", d.Id())

	_, hasLifecyclePolicy := d.GetOk("lifecycle_policy")
	if hasLifecyclePolicy {
		_, err := conn.PutLifecycleConfiguration(&efs.PutLifecycleConfigurationInput{
			FileSystemId:      aws.String(d.Id()),
			LifecyclePolicies: expandEfsFileSystemLifecyclePolicies(d.Get("lifecycle_policy").([]interface{})),
		})
		if err != nil {
			return fmt.Errorf("Error creating lifecycle policy for EFS file system %q: %s",
				d.Id(), err.Error())
		}
	}

	return resourceAwsEfsFileSystemRead(d, meta)
}

func resourceAwsEfsFileSystemUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	if d.HasChanges("provisioned_throughput_in_mibps", "throughput_mode") {
		throughputMode := d.Get("throughput_mode").(string)

		input := &efs.UpdateFileSystemInput{
			FileSystemId:   aws.String(d.Id()),
			ThroughputMode: aws.String(throughputMode),
		}

		if throughputMode == efs.ThroughputModeProvisioned {
			input.ProvisionedThroughputInMibps = aws.Float64(d.Get("provisioned_throughput_in_mibps").(float64))
		}

		_, err := conn.UpdateFileSystem(input)
		if err != nil {
			return fmt.Errorf("error updating EFS file system (%s): %w", d.Id(), err)
		}

		if _, err := waiter.FileSystemAvailable(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for EFS file system (%s) to be available: %w", d.Id(), err)
		}
	}

	if d.HasChange("lifecycle_policy") {
		input := &efs.PutLifecycleConfigurationInput{
			FileSystemId:      aws.String(d.Id()),
			LifecyclePolicies: expandEfsFileSystemLifecyclePolicies(d.Get("lifecycle_policy").([]interface{})),
		}

		// Prevent the following error during removal:
		// InvalidParameter: 1 validation error(s) found.
		// - missing required field, PutLifecycleConfigurationInput.LifecyclePolicies.
		if input.LifecyclePolicies == nil {
			input.LifecyclePolicies = []*efs.LifecyclePolicy{}
		}

		_, err := conn.PutLifecycleConfiguration(input)
		if err != nil {
			return fmt.Errorf("Error updating lifecycle policy for EFS file system %q: %s",
				d.Id(), err.Error())
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.EfsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EFS file system (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsEfsFileSystemRead(d, meta)
}

func resourceAwsEfsFileSystemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
		FileSystemId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") {
			log.Printf("[WARN] EFS file system (%s) could not be found.", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	if hasEmptyFileSystems(resp) {
		return fmt.Errorf("EFS file system %q could not be found.", d.Id())
	}

	var fs *efs.FileSystemDescription
	for _, f := range resp.FileSystems {
		if d.Id() == *f.FileSystemId {
			fs = f
			break
		}
	}
	if fs == nil {
		log.Printf("[WARN] EFS File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", fs.FileSystemArn)
	d.Set("availability_zone_id", fs.AvailabilityZoneId)
	d.Set("availability_zone_name", fs.AvailabilityZoneName)
	d.Set("creation_token", fs.CreationToken)
	d.Set("encrypted", fs.Encrypted)
	d.Set("kms_key_id", fs.KmsKeyId)
	d.Set("performance_mode", fs.PerformanceMode)
	d.Set("provisioned_throughput_in_mibps", fs.ProvisionedThroughputInMibps)
	d.Set("throughput_mode", fs.ThroughputMode)
	d.Set("owner_id", fs.OwnerId)
	d.Set("number_of_mount_targets", fs.NumberOfMountTargets)

	if err := d.Set("tags", keyvaluetags.EfsKeyValueTags(fs.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("size_in_bytes", flattenEfsFileSystemSizeInBytes(fs.SizeInBytes)); err != nil {
		return fmt.Errorf("error setting size_in_bytes: %w", err)
	}

	d.Set("dns_name", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(fs.FileSystemId))))

	res, err := conn.DescribeLifecycleConfiguration(&efs.DescribeLifecycleConfigurationInput{
		FileSystemId: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("Error describing lifecycle configuration for EFS file system (%s): %w", d.Id(), err)
	}

	if err := d.Set("lifecycle_policy", flattenEfsFileSystemLifecyclePolicies(res.LifecyclePolicies)); err != nil {
		return fmt.Errorf("error setting lifecycle_policy: %w", err)
	}

	return nil
}

func resourceAwsEfsFileSystemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).efsconn

	log.Printf("[DEBUG] Deleting EFS file system: %s", d.Id())
	_, err := conn.DeleteFileSystem(&efs.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") {
			return nil
		}
		return fmt.Errorf("Error delete file system: %s with err %s", d.Id(), err.Error())
	}

	if _, err := waiter.FileSystemDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") {
			return nil
		}
		return fmt.Errorf("error waiting for EFS file system (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func hasEmptyFileSystems(fs *efs.DescribeFileSystemsOutput) bool {
	if fs != nil && len(fs.FileSystems) > 0 {
		return false
	}
	return true
}

func flattenEfsFileSystemLifecyclePolicies(apiObjects []*efs.LifecyclePolicy) []interface{} {
	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfMap := make(map[string]interface{})

		if apiObject.TransitionToIA != nil {
			tfMap["transition_to_ia"] = aws.StringValue(apiObject.TransitionToIA)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandEfsFileSystemLifecyclePolicies(tfList []interface{}) []*efs.LifecyclePolicy {
	var apiObjects []*efs.LifecyclePolicy

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := &efs.LifecyclePolicy{}

		if v, ok := tfMap["transition_to_ia"].(string); ok && v != "" {
			apiObject.TransitionToIA = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEfsFileSystemSizeInBytes(sizeInBytes *efs.FileSystemSize) []interface{} {
	if sizeInBytes == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"value": aws.Int64Value(sizeInBytes.Value),
	}

	if sizeInBytes.ValueInIA != nil {
		m["value_in_ia"] = aws.Int64Value(sizeInBytes.ValueInIA)
	}

	if sizeInBytes.ValueInStandard != nil {
		m["value_in_standard"] = aws.Int64Value(sizeInBytes.ValueInStandard)
	}

	return []interface{}{m}
}
