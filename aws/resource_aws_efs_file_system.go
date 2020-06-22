package aws

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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

			"creation_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},

			"reference_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Removed:  "Use `creation_token` argument instead",
			},

			"performance_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					efs.PerformanceModeGeneralPurpose,
					efs.PerformanceModeMaxIo,
				}, false),
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

			"tags": tagsSchema(),

			"throughput_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  efs.ThroughputModeBursting,
				ValidateFunc: validation.StringInSlice([]string{
					efs.ThroughputModeBursting,
					efs.ThroughputModeProvisioned,
				}, false),
			},

			"lifecycle_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transition_to_ia": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								efs.TransitionToIARulesAfter7Days,
								efs.TransitionToIARulesAfter14Days,
								efs.TransitionToIARulesAfter30Days,
								efs.TransitionToIARulesAfter60Days,
								efs.TransitionToIARulesAfter90Days,
							}, false),
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
		return fmt.Errorf("Error creating EFS file system: %s", err)
	}

	d.SetId(*fs.FileSystemId)
	log.Printf("[INFO] EFS file system ID: %s", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{efs.LifeCycleStateCreating},
		Target:     []string{efs.LifeCycleStateAvailable},
		Refresh:    resourceEfsFileSystemCreateUpdateRefreshFunc(d.Id(), conn),
		Timeout:    10 * time.Minute,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for EFS file system (%q) to create: %s", d.Id(), err)
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
			return fmt.Errorf("error updating EFS File System %q: %s", d.Id(), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{efs.LifeCycleStateUpdating},
			Target:     []string{efs.LifeCycleStateAvailable},
			Refresh:    resourceEfsFileSystemCreateUpdateRefreshFunc(d.Id(), conn),
			Timeout:    10 * time.Minute,
			Delay:      2 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("error waiting for EFS file system (%q) to update: %s", d.Id(), err)
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
			return fmt.Errorf("error updating EFS file system (%s) tags: %s", d.Id(), err)
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
		log.Printf("[WARN] EFS (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	fsARN := arn.ARN{
		AccountID: meta.(*AWSClient).accountid,
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Resource:  fmt.Sprintf("file-system/%s", aws.StringValue(fs.FileSystemId)),
		Service:   "elasticfilesystem",
	}.String()

	d.Set("arn", fsARN)
	d.Set("creation_token", fs.CreationToken)
	d.Set("encrypted", fs.Encrypted)
	d.Set("kms_key_id", fs.KmsKeyId)
	d.Set("performance_mode", fs.PerformanceMode)
	d.Set("provisioned_throughput_in_mibps", fs.ProvisionedThroughputInMibps)
	d.Set("throughput_mode", fs.ThroughputMode)

	if err := d.Set("tags", keyvaluetags.EfsKeyValueTags(fs.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("dns_name", meta.(*AWSClient).RegionalHostname(fmt.Sprintf("%s.efs", aws.StringValue(fs.FileSystemId))))

	res, err := conn.DescribeLifecycleConfiguration(&efs.DescribeLifecycleConfigurationInput{
		FileSystemId: fs.FileSystemId,
	})
	if err != nil {
		return fmt.Errorf("Error describing lifecycle configuration for EFS file system (%s): %s",
			aws.StringValue(fs.FileSystemId), err)
	}

	if err := d.Set("lifecycle_policy", flattenEfsFileSystemLifecyclePolicies(res.LifecyclePolicies)); err != nil {
		return fmt.Errorf("error setting lifecycle_policy: %s", err)
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
		return fmt.Errorf("Error delete file system: %s with err %s", d.Id(), err.Error())
	}

	err = waitForDeleteEfsFileSystem(conn, d.Id(), 10*time.Minute)
	if err != nil {
		return fmt.Errorf("Error waiting for EFS file system (%q) to delete: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] EFS file system %q deleted.", d.Id())

	return nil
}

func waitForDeleteEfsFileSystem(conn *efs.EFS, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"available", "deleting"},
		Target:  []string{},
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
				FileSystemId: aws.String(id),
			})
			if err != nil {
				if isAWSErr(err, efs.ErrCodeFileSystemNotFound, "") {
					return nil, "", nil
				}
				return nil, "error", err
			}

			if hasEmptyFileSystems(resp) {
				return nil, "", nil
			}

			fs := resp.FileSystems[0]
			log.Printf("[DEBUG] current status of %q: %q", *fs.FileSystemId, *fs.LifeCycleState)
			return fs, *fs.LifeCycleState, nil
		},
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func hasEmptyFileSystems(fs *efs.DescribeFileSystemsOutput) bool {
	if fs != nil && len(fs.FileSystems) > 0 {
		return false
	}
	return true
}

func resourceEfsFileSystemCreateUpdateRefreshFunc(id string, conn *efs.EFS) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeFileSystems(&efs.DescribeFileSystemsInput{
			FileSystemId: aws.String(id),
		})
		if err != nil {
			return nil, "error", err
		}

		if hasEmptyFileSystems(resp) {
			return nil, "not-found", fmt.Errorf("EFS file system %q could not be found.", id)
		}

		fs := resp.FileSystems[0]
		state := aws.StringValue(fs.LifeCycleState)
		log.Printf("[DEBUG] current status of %q: %q", id, state)
		return fs, state, nil
	}
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
