package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsBatchComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBatchComputeEnvironmentCreate,
		Read:   resourceAwsBatchComputeEnvironmentRead,
		Update: resourceAwsBatchComputeEnvironmentUpdate,
		Delete: resourceAwsBatchComputeEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceAwsBatchComputeEnvironmentCustomizeDiff,
			SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"compute_environment_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name_prefix"},
				ValidateFunc:  validateBatchName,
			},
			"compute_environment_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name"},
				ValidateFunc:  validateBatchPrefix,
			},
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRAllocationStrategy_Values(), true),
						},
						"bid_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"desired_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"ec2_key_pair": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"image_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"instance_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
						"instance_type": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"launch_template": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ForceNew:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
									},
									"launch_template_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ForceNew:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"max_vcpus": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"min_vcpus": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"spot_iam_fleet_role": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tags": tagsSchemaForceNew(),
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							StateFunc: func(val interface{}) string {
								return strings.ToUpper(val.(string))
							},
							ValidateFunc: validation.StringInSlice(batch.CRType_Values(), true),
						},
					},
				},
			},
			"service_role": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateArn,
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(batch.CEState_Values(), true),
				Default:      batch.CEStateEnabled,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(batch.CEType_Values(), true),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ecs_cluster_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsBatchComputeEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	computeEnvironmentName := naming.Generate(d.Get("compute_environment_name").(string), d.Get("compute_environment_name_prefix").(string))
	computeEnvironmentType := d.Get("type").(string)

	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(computeEnvironmentName),
		ServiceRole:            aws.String(d.Get("service_role").(string)),
		Type:                   aws.String(computeEnvironmentType),
	}

	// TODO Check in CustomizeDiff that UNMANAGED compute environment has no compute_resources.
	// TODO This would be a breaking change.
	if computeEnvironmentType := strings.ToUpper(computeEnvironmentType); computeEnvironmentType == batch.CETypeManaged {
		if v, ok := d.GetOk("compute_resources"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ComputeResources = expandBatchComputeResource(v.([]interface{})[0].(map[string]interface{}))
		}
	}

	if v, ok := d.GetOk("state"); ok {
		input.State = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().BatchTags()
	}

	log.Printf("[DEBUG] Creating Batch Compute Environment: %s", input)
	output, err := conn.CreateComputeEnvironment(input)

	if err != nil {
		return fmt.Errorf("error creating Batch Compute Environment (%s): %w", computeEnvironmentName, err)
	}

	d.SetId(aws.StringValue(output.ComputeEnvironmentName))

	if _, err := waiter.ComputeEnvironmentCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Batch Compute Environment (%s) create: %w", d.Id(), err)
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func resourceAwsBatchComputeEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	computeEnvironment, err := finder.ComputeEnvironmentDetailByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Compute Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Batch Compute Environment (%s): %w", d.Id(), err)
	}

	computeEnvironmentType := aws.StringValue(computeEnvironment.Type)

	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("compute_environment_name", computeEnvironment.ComputeEnvironmentName)
	d.Set("compute_environment_name_prefix", naming.NamePrefixFromName(aws.StringValue(computeEnvironment.ComputeEnvironmentName)))
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("state", computeEnvironment.State)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)
	d.Set("type", computeEnvironmentType)

	// TODO See above on how to remove check on type.
	if computeEnvironmentType == batch.CETypeManaged {
		if computeEnvironment.ComputeResources != nil {
			if err := d.Set("compute_resources", []interface{}{flattenBatchComputeResource(computeEnvironment.ComputeResources)}); err != nil {
				return fmt.Errorf("error setting compute_resources: %w", err)
			}
		} else {
			d.Set("compute_resources", nil)
		}
	}

	tags := keyvaluetags.BatchKeyValueTags(computeEnvironment.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsBatchComputeEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if d.HasChange("service_role") {
			input.ServiceRole = aws.String(d.Get("service_role").(string))
		}

		if d.HasChange("state") {
			input.State = aws.String(d.Get("state").(string))
		}

		// TODO See above on how to remove check on type.
		if computeEnvironmentType := strings.ToUpper(d.Get("type").(string)); computeEnvironmentType == batch.CETypeManaged {
			// "At least one compute-resources attribute must be specified"
			computeResourceUpdate := &batch.ComputeResourceUpdate{
				MaxvCpus: aws.Int64(int64(d.Get("compute_resources.0.max_vcpus").(int))),
			}

			if d.HasChange("compute_resources.0.desired_vcpus") {
				computeResourceUpdate.DesiredvCpus = aws.Int64(int64(d.Get("compute_resources.0.desired_vcpus").(int)))
			}

			if d.HasChange("compute_resources.0.min_vcpus") {
				computeResourceUpdate.MinvCpus = aws.Int64(int64(d.Get("compute_resources.0.min_vcpus").(int)))
			}

			if d.HasChange("compute_resources.0.security_group_ids") {
				computeResourceUpdate.SecurityGroupIds = expandStringSet(d.Get("compute_resources.0.security_group_ids").(*schema.Set))
			}

			if d.HasChange("compute_resources.0.subnets") {
				computeResourceUpdate.Subnets = expandStringSet(d.Get("compute_resources.0.subnets").(*schema.Set))
			}

			input.ComputeResources = computeResourceUpdate
		}

		log.Printf("[DEBUG] Updating Batch Compute Environment: %s", input)
		if _, err := conn.UpdateComputeEnvironment(input); err != nil {
			return fmt.Errorf("error updating Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ComputeEnvironmentUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.BatchUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func resourceAwsBatchComputeEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	log.Printf("[DEBUG] Disabling Batch Compute Environment (%s)", d.Id())
	{
		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
			State:              aws.String(batch.CEStateDisabled),
		}

		if _, err := conn.UpdateComputeEnvironment(input); err != nil {
			return fmt.Errorf("error disabling Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ComputeEnvironmentDisabled(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) disable: %w", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Batch Compute Environment (%s)", d.Id())
	{
		input := &batch.DeleteComputeEnvironmentInput{
			ComputeEnvironment: aws.String(d.Id()),
		}

		if _, err := conn.DeleteComputeEnvironment(input); err != nil {
			return fmt.Errorf("error deleting Batch Compute Environment (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ComputeEnvironmentDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) delete: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsBatchComputeEnvironmentCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if diff.Id() != "" {
		// Update.

		computeResourceType := strings.ToUpper(diff.Get("compute_resources.0.type").(string))
		fargateComputeResources := false
		if computeResourceType == batch.CRTypeFargate || computeResourceType == batch.CRTypeFargateSpot {
			fargateComputeResources = true
		}

		if diff.HasChange("compute_resources.0.security_group_ids") && !fargateComputeResources {
			if err := diff.ForceNew("compute_resources.0.security_group_ids"); err != nil {
				return err
			}
		}

		if diff.HasChange("compute_resources.0.subnets") && !fargateComputeResources {
			if err := diff.ForceNew("compute_resources.0.subnets"); err != nil {
				return err
			}
		}
	}

	return nil
}

func expandBatchComputeResource(tfMap map[string]interface{}) *batch.ComputeResource {
	if tfMap == nil {
		return nil
	}

	var computeResourceType string

	if v, ok := tfMap["type"].(string); ok && v != "" {
		computeResourceType = v
	}

	apiObject := &batch.ComputeResource{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = aws.String(v)
	}

	if v, ok := tfMap["bid_percentage"].(int); ok && v != 0 {
		apiObject.BidPercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["desired_vcpus"].(int); ok && v != 0 {
		apiObject.DesiredvCpus = aws.Int64(int64(v))
	}

	if v, ok := tfMap["ec2_key_pair"].(string); ok && v != "" {
		apiObject.Ec2KeyPair = aws.String(v)
	}

	if v, ok := tfMap["image_id"].(string); ok && v != "" {
		apiObject.ImageId = aws.String(v)
	}

	if v, ok := tfMap["instance_role"].(string); ok && v != "" {
		apiObject.InstanceRole = aws.String(v)
	}

	if v, ok := tfMap["instance_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceTypes = expandStringSet(v)
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplate = expandBatchLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_vcpus"].(int); ok && v != 0 {
		apiObject.MaxvCpus = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_vcpus"].(int); ok && v != 0 {
		apiObject.MinvCpus = aws.Int64(int64(v))
	} else if computeResourceType := strings.ToUpper(computeResourceType); computeResourceType == batch.CRTypeEc2 || computeResourceType == batch.CRTypeSpot {
		apiObject.MinvCpus = aws.Int64(0)
	}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = expandStringSet(v)
	}

	if v, ok := tfMap["spot_iam_fleet_role"].(string); ok && v != "" {
		apiObject.SpotIamFleetRole = aws.String(v)
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Subnets = expandStringSet(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Tags = keyvaluetags.New(v).IgnoreAws().BatchTags()
	}

	if computeResourceType != "" {
		apiObject.Type = aws.String(computeResourceType)
	}

	return apiObject
}

func expandBatchLaunchTemplateSpecification(tfMap map[string]interface{}) *batch.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.LaunchTemplateSpecification{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func flattenBatchComputeResource(apiObject *batch.ComputeResource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != nil {
		tfMap["allocation_strategy"] = aws.StringValue(v)
	}

	if v := apiObject.BidPercentage; v != nil {
		tfMap["bid_percentage"] = aws.Int64Value(v)
	}

	if v := apiObject.DesiredvCpus; v != nil {
		tfMap["desired_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.Ec2KeyPair; v != nil {
		tfMap["ec2_key_pair"] = aws.StringValue(v)
	}

	if v := apiObject.ImageId; v != nil {
		tfMap["image_id"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceRole; v != nil {
		tfMap["instance_role"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceTypes; v != nil {
		tfMap["instance_type"] = aws.StringValueSlice(v)
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenBatchLaunchTemplateSpecification(v)}
	}

	if v := apiObject.MaxvCpus; v != nil {
		tfMap["max_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.MinvCpus; v != nil {
		tfMap["min_vcpus"] = aws.Int64Value(v)
	}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_group_ids"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SpotIamFleetRole; v != nil {
		tfMap["spot_iam_fleet_role"] = aws.StringValue(v)
	}

	if v := apiObject.Subnets; v != nil {
		tfMap["subnets"] = aws.StringValueSlice(v)
	}

	if v := apiObject.Tags; v != nil {
		tfMap["tags"] = keyvaluetags.BatchKeyValueTags(v).IgnoreAws().Map()
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenBatchLaunchTemplateSpecification(apiObject *batch.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.StringValue(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap["version"] = aws.StringValue(v)
	}

	return tfMap
}
