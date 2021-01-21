package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func resourceAwsSagemakerNotebookInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerNotebookInstanceCreate,
		Read:   resourceAwsSagemakerNotebookInstanceRead,
		Update: resourceAwsSagemakerNotebookInstanceUpdate,
		Delete: resourceAwsSagemakerNotebookInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange("volume_size", func(_ context.Context, old, new, meta interface{}) bool {
				return new.(int) < old.(int)
			}),
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},

			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},

			"instance_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.InstanceType_Values(), false),
			},
			"additional_code_repositories": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 3,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"volume_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
			},

			"subnet_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"security_groups": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"lifecycle_config_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"root_access": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      sagemaker.RootAccessEnabled,
				ValidateFunc: validation.StringInSlice(sagemaker.RootAccess_Values(), false),
			},

			"direct_internet_access": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      sagemaker.DirectInternetAccessEnabled,
				ValidateFunc: validation.StringInSlice(sagemaker.DirectInternetAccess_Values(), false),
			},

			"default_code_repository": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_interface_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerNotebookInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("name").(string)

	createOpts := &sagemaker.CreateNotebookInstanceInput{
		SecurityGroupIds:     expandStringSet(d.Get("security_groups").(*schema.Set)),
		NotebookInstanceName: aws.String(name),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		InstanceType:         aws.String(d.Get("instance_type").(string)),
	}

	if v, ok := d.GetOk("root_access"); ok {
		createOpts.RootAccess = aws.String(v.(string))
	}

	if v, ok := d.GetOk("direct_internet_access"); ok {
		createOpts.DirectInternetAccess = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_code_repository"); ok {
		createOpts.DefaultCodeRepository = aws.String(v.(string))
	}

	if s, ok := d.GetOk("subnet_id"); ok {
		createOpts.SubnetId = aws.String(s.(string))
	}

	if v, ok := d.GetOk("volume_size"); ok {
		createOpts.VolumeSizeInGB = aws.Int64(int64(v.(int)))
	}

	if k, ok := d.GetOk("kms_key_id"); ok {
		createOpts.KmsKeyId = aws.String(k.(string))
	}

	if l, ok := d.GetOk("lifecycle_config_name"); ok {
		createOpts.LifecycleConfigName = aws.String(l.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		createOpts.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	if v, ok := d.GetOk("additional_code_repositories"); ok && v.(*schema.Set).Len() > 0 {
		createOpts.AdditionalCodeRepositories = expandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] sagemaker notebook instance create config: %#v", *createOpts)
	_, err := conn.CreateNotebookInstance(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Sagemaker Notebook Instance: %s", err)
	}

	d.SetId(name)
	log.Printf("[INFO] sagemaker notebook instance ID: %s", d.Id())

	if _, err := waiter.NotebookInstanceInService(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsSagemakerNotebookInstanceRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}
	notebookInstance, err := conn.DescribeNotebookInstance(describeNotebookInput)
	if err != nil {
		if isAWSErr(err, "ValidationException", "RecordNotFound") {
			d.SetId("")
			log.Printf("[WARN] Unable to find sageMaker notebook instance (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error finding sagemaker notebook instance (%s): %s", d.Id(), err)

	}

	if err := d.Set("security_groups", flattenStringList(notebookInstance.SecurityGroups)); err != nil {
		return fmt.Errorf("error setting security groups for sagemaker notebook instance (%s): %s", d.Id(), err)
	}
	if err := d.Set("name", notebookInstance.NotebookInstanceName); err != nil {
		return fmt.Errorf("error setting name for sagemaker notebook instance (%s): %s", d.Id(), err)
	}
	if err := d.Set("role_arn", notebookInstance.RoleArn); err != nil {
		return fmt.Errorf("error setting role_arn for sagemaker notebook instance (%s): %s", d.Id(), err)
	}
	if err := d.Set("instance_type", notebookInstance.InstanceType); err != nil {
		return fmt.Errorf("error setting instance_type for sagemaker notebook instance (%s): %s", d.Id(), err)
	}
	if err := d.Set("subnet_id", notebookInstance.SubnetId); err != nil {
		return fmt.Errorf("error setting subnet_id for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("kms_key_id", notebookInstance.KmsKeyId); err != nil {
		return fmt.Errorf("error setting kms_key_id for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("volume_size", notebookInstance.VolumeSizeInGB); err != nil {
		return fmt.Errorf("error setting volume_size for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("lifecycle_config_name", notebookInstance.NotebookInstanceLifecycleConfigName); err != nil {
		return fmt.Errorf("error setting lifecycle_config_name for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("arn", notebookInstance.NotebookInstanceArn); err != nil {
		return fmt.Errorf("error setting arn for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("root_access", notebookInstance.RootAccess); err != nil {
		return fmt.Errorf("error setting root_access for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("direct_internet_access", notebookInstance.DirectInternetAccess); err != nil {
		return fmt.Errorf("error setting direct_internet_access for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("default_code_repository", notebookInstance.DefaultCodeRepository); err != nil {
		return fmt.Errorf("error setting default_code_repository for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("url", notebookInstance.Url); err != nil {
		return fmt.Errorf("error setting url for sagemaker notebook instance (%s): %w", d.Id(), err)
	}

	if err := d.Set("network_interface_id", notebookInstance.NetworkInterfaceId); err != nil {
		return fmt.Errorf("error setting network_interface_id for sagemaker notebook instance (%s): %w", d.Id(), err)
	}

	if err := d.Set("additional_code_repositories", flattenStringSet(notebookInstance.AdditionalCodeRepositories)); err != nil {
		return fmt.Errorf("error setting additional_code_repositories for sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, aws.StringValue(notebookInstance.NotebookInstanceArn))

	if err != nil {
		return fmt.Errorf("error listing tags for Sagemaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSagemakerNotebookInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Sagemaker Notebook Instance (%s) tags: %s", d.Id(), err)
		}
	}

	hasChanged := false
	// Update
	updateOpts := &sagemaker.UpdateNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("role_arn") {
		updateOpts.RoleArn = aws.String(d.Get("role_arn").(string))
		hasChanged = true
	}

	if d.HasChange("instance_type") {
		updateOpts.InstanceType = aws.String(d.Get("instance_type").(string))
		hasChanged = true
	}

	if d.HasChange("volume_size") {
		updateOpts.VolumeSizeInGB = aws.Int64(int64(d.Get("volume_size").(int)))
		hasChanged = true
	}

	if d.HasChange("lifecycle_config_name") {
		if v, ok := d.GetOk("lifecycle_config_name"); ok {
			updateOpts.LifecycleConfigName = aws.String(v.(string))
		} else {
			updateOpts.DisassociateLifecycleConfig = aws.Bool(true)
		}
		hasChanged = true
	}

	if d.HasChange("default_code_repository") {
		if v, ok := d.GetOk("default_code_repository"); ok {
			updateOpts.DefaultCodeRepository = aws.String(v.(string))
		} else {
			updateOpts.DisassociateDefaultCodeRepository = aws.Bool(true)
		}
		hasChanged = true
	}

	if d.HasChange("root_access") {
		updateOpts.RootAccess = aws.String(d.Get("root_access").(string))
		hasChanged = true
	}

	if d.HasChange("additional_code_repositories") {
		if v, ok := d.GetOk("additional_code_repositories"); ok {
			updateOpts.AdditionalCodeRepositories = expandStringSet(v.(*schema.Set))
		} else {
			updateOpts.DisassociateAdditionalCodeRepositories = aws.Bool(true)
		}
		hasChanged = true
	}

	if hasChanged {

		// Stop notebook
		_, previousStatus, _ := sagemakerNotebookInstanceStateRefreshFunc(conn, d.Id())()
		if previousStatus != sagemaker.NotebookInstanceStatusStopped {
			if err := stopSagemakerNotebookInstance(conn, d.Id()); err != nil {
				return fmt.Errorf("error stopping sagemaker notebook instance prior to updating: %s", err)
			}
		}

		if _, err := conn.UpdateNotebookInstance(updateOpts); err != nil {
			return fmt.Errorf("error updating sagemaker notebook instance: %s", err)
		}

		if _, err := waiter.NotebookInstanceStopped(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to stop: %w", d.Id(), err)
		}

		// Restart if needed
		if previousStatus == sagemaker.NotebookInstanceStatusInService {
			startOpts := &sagemaker.StartNotebookInstanceInput{
				NotebookInstanceName: aws.String(d.Id()),
			}
			stateConf := &resource.StateChangeConf{
				Pending: []string{
					sagemaker.NotebookInstanceStatusStopped,
				},
				Target:  []string{sagemaker.NotebookInstanceStatusInService, sagemaker.NotebookInstanceStatusPending},
				Refresh: sagemakerNotebookInstanceStateRefreshFunc(conn, d.Id()),
				Timeout: 30 * time.Second,
			}
			// StartNotebookInstance sometimes doesn't take so we'll check for a state change and if
			// it doesn't change we'll send another request
			err := resource.Retry(5*time.Minute, func() *resource.RetryError {
				_, err := conn.StartNotebookInstance(startOpts)
				if err != nil {
					return resource.NonRetryableError(fmt.Errorf("error starting sagemaker notebook instance (%s): %s", d.Id(), err))
				}

				_, err = stateConf.WaitForState()
				if err != nil {
					return resource.RetryableError(fmt.Errorf("error waiting for sagemaker notebook instance (%s) to start: %s", d.Id(), err))
				}

				return nil
			})
			if isResourceTimeoutError(err) {
				_, err = conn.StartNotebookInstance(startOpts)
				if err != nil {
					return fmt.Errorf("error starting sagemaker notebook instance (%s): %s", d.Id(), err)
				}

				_, err = stateConf.WaitForState()
			}
			if err != nil {
				return fmt.Errorf("Error waiting for sagemaker notebook instance to start: %s", err)
			}

			if _, err := waiter.NotebookInstanceInService(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to to start after update: %w", d.Id(), err)
			}
		}
	}

	return resourceAwsSagemakerNotebookInstanceRead(d, meta)
}

func resourceAwsSagemakerNotebookInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}
	notebook, err := conn.DescribeNotebookInstance(describeNotebookInput)
	if err != nil {
		if isAWSErr(err, "ValidationException", "RecordNotFound") {
			return nil
		}
		return fmt.Errorf("unable to find sagemaker notebook instance to delete (%s): %s", d.Id(), err)
	}

	if aws.StringValue(notebook.NotebookInstanceStatus) != sagemaker.NotebookInstanceStatusFailed &&
		aws.StringValue(notebook.NotebookInstanceStatus) != sagemaker.NotebookInstanceStatusStopped {
		if err := stopSagemakerNotebookInstance(conn, d.Id()); err != nil {
			return err
		}
	}

	deleteOpts := &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteNotebookInstance(deleteOpts); err != nil {
		return fmt.Errorf("error trying to delete sagemaker notebook instance (%s): %s", d.Id(), err)
	}

	if _, err := waiter.NotebookInstanceDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, "ValidationException", "RecordNotFound") {
			return nil
		}
		return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func stopSagemakerNotebookInstance(conn *sagemaker.SageMaker, id string) error {
	describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}
	notebook, err := conn.DescribeNotebookInstance(describeNotebookInput)
	if err != nil {
		if isAWSErr(err, "ValidationException", "RecordNotFound") {
			return nil
		}
		return fmt.Errorf("unable to find sagemaker notebook instance (%s): %s", id, err)
	}
	if aws.StringValue(notebook.NotebookInstanceStatus) == sagemaker.NotebookInstanceStatusStopped {
		return nil
	}

	stopOpts := &sagemaker.StopNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}

	if _, err := conn.StopNotebookInstance(stopOpts); err != nil {
		return fmt.Errorf("Error stopping sagemaker notebook instance: %s", err)
	}

	if _, err := waiter.NotebookInstanceStopped(conn, id); err != nil {
		return fmt.Errorf("error waiting for sagemaker notebook instance (%s) to stop: %w", id, err)
	}

	return nil
}

func sagemakerNotebookInstanceStateRefreshFunc(conn *sagemaker.SageMaker, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeNotebookInput := &sagemaker.DescribeNotebookInstanceInput{
			NotebookInstanceName: aws.String(name),
		}
		notebook, err := conn.DescribeNotebookInstance(describeNotebookInput)
		if err != nil {
			if isAWSErr(err, "ValidationException", "RecordNotFound") {
				return 1, "", nil
			}
			return nil, "", err
		}

		if notebook == nil {
			return nil, "", nil
		}

		return notebook, *notebook.NotebookInstanceStatus, nil
	}
}
