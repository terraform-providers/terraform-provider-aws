package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsBatchComputeEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBatchComputeEnvironmentCreate,
		Read:   resourceAwsBatchComputeEnvironmentRead,
		Update: resourceAwsBatchComputeEnvironmentUpdate,
		Delete: resourceAwsBatchComputeEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("compute_environment_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},
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
				ForceNew:      true,
				ConflictsWith: []string{"compute_environment_name"},
				ValidateFunc:  validateBatchPrefix,
			},
			"compute_resources": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								batch.CRAllocationStrategyBestFit,
								batch.CRAllocationStrategyBestFitProgressive,
								batch.CRAllocationStrategySpotCapacityOptimized}, true),
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
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
						"instance_type": {
							Type:     schema.TypeSet,
							Required: true,
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
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_name"},
										ForceNew:      true,
									},
									"launch_template_name": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"compute_resources.0.launch_template.0.launch_template_id"},
										ForceNew:      true,
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
							Required: true,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
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
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"tags": tagsSchemaForceNew(),
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice([]string{batch.CRTypeEc2, batch.CRTypeSpot}, true),
						},
					},
				},
			},
			"service_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"state": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{batch.CEStateEnabled, batch.CEStateDisabled}, true),
				Default:      batch.CEStateEnabled,
			},
			"tags": tagsSchema(),
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{batch.CETypeManaged, batch.CETypeUnmanaged}, true),
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

	// Build the compute environment name.
	var computeEnvironmentName string
	if v, ok := d.GetOk("compute_environment_name"); ok {
		computeEnvironmentName = v.(string)
	} else if v, ok := d.GetOk("compute_environment_name_prefix"); ok {
		computeEnvironmentName = resource.PrefixedUniqueId(v.(string))
	} else {
		computeEnvironmentName = resource.UniqueId()
	}
	d.Set("compute_environment_name", computeEnvironmentName)

	serviceRole := d.Get("service_role").(string)
	computeEnvironmentType := d.Get("type").(string)

	input := &batch.CreateComputeEnvironmentInput{
		ComputeEnvironmentName: aws.String(computeEnvironmentName),
		ServiceRole:            aws.String(serviceRole),
		Type:                   aws.String(computeEnvironmentType),
	}

	if v, ok := d.GetOk("state"); ok {
		input.State = aws.String(v.(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().BatchTags()
	}

	if computeEnvironmentType == batch.CETypeManaged {
		computeResources := d.Get("compute_resources").([]interface{})
		if len(computeResources) == 0 {
			return fmt.Errorf("One compute environment is expected, but no compute environments are set")
		}
		computeResource := computeResources[0].(map[string]interface{})

		instanceRole := computeResource["instance_role"].(string)
		maxvCpus := int64(computeResource["max_vcpus"].(int))
		minvCpus := int64(computeResource["min_vcpus"].(int))
		computeResourceType := computeResource["type"].(string)

		var instanceTypes []*string
		for _, v := range computeResource["instance_type"].(*schema.Set).List() {
			instanceTypes = append(instanceTypes, aws.String(v.(string)))
		}

		var securityGroupIds []*string
		for _, v := range computeResource["security_group_ids"].(*schema.Set).List() {
			securityGroupIds = append(securityGroupIds, aws.String(v.(string)))
		}

		var subnets []*string
		for _, v := range computeResource["subnets"].(*schema.Set).List() {
			subnets = append(subnets, aws.String(v.(string)))
		}

		input.ComputeResources = &batch.ComputeResource{
			InstanceRole:     aws.String(instanceRole),
			InstanceTypes:    instanceTypes,
			MaxvCpus:         aws.Int64(maxvCpus),
			MinvCpus:         aws.Int64(minvCpus),
			SecurityGroupIds: securityGroupIds,
			Subnets:          subnets,
			Type:             aws.String(computeResourceType),
		}

		if v, ok := computeResource["allocation_strategy"]; ok {
			input.ComputeResources.AllocationStrategy = aws.String(v.(string))
		}
		if v, ok := computeResource["bid_percentage"]; ok {
			input.ComputeResources.BidPercentage = aws.Int64(int64(v.(int)))
		}
		if v, ok := computeResource["desired_vcpus"]; ok && v.(int) > 0 {
			input.ComputeResources.DesiredvCpus = aws.Int64(int64(v.(int)))
		}
		if v, ok := computeResource["ec2_key_pair"]; ok {
			input.ComputeResources.Ec2KeyPair = aws.String(v.(string))
		}
		if v, ok := computeResource["image_id"]; ok {
			input.ComputeResources.ImageId = aws.String(v.(string))
		}
		if v, ok := computeResource["spot_iam_fleet_role"]; ok {
			input.ComputeResources.SpotIamFleetRole = aws.String(v.(string))
		}
		if v, ok := computeResource["tags"]; ok {
			input.ComputeResources.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().BatchTags()
		}

		if raw, ok := computeResource["launch_template"]; ok && len(raw.([]interface{})) > 0 {
			input.ComputeResources.LaunchTemplate = &batch.LaunchTemplateSpecification{}
			launchTemplate := raw.([]interface{})[0].(map[string]interface{})
			if v, ok := launchTemplate["launch_template_id"]; ok {
				input.ComputeResources.LaunchTemplate.LaunchTemplateId = aws.String(v.(string))
			}
			if v, ok := launchTemplate["launch_template_name"]; ok {
				input.ComputeResources.LaunchTemplate.LaunchTemplateName = aws.String(v.(string))
			}
			if v, ok := launchTemplate["version"]; ok {
				input.ComputeResources.LaunchTemplate.Version = aws.String(v.(string))
			}
		}
	}

	log.Printf("[DEBUG] Create compute environment %s.\n", input)

	if _, err := conn.CreateComputeEnvironment(input); err != nil {
		return err
	}

	d.SetId(computeEnvironmentName)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{batch.CEStatusCreating},
		Target:     []string{batch.CEStatusValid},
		Refresh:    resourceAwsBatchComputeEnvironmentStatusRefreshFunc(computeEnvironmentName, conn),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return err
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func resourceAwsBatchComputeEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	computeEnvironmentName := d.Get("compute_environment_name").(string)

	input := &batch.DescribeComputeEnvironmentsInput{
		ComputeEnvironments: []*string{
			aws.String(computeEnvironmentName),
		},
	}

	log.Printf("[DEBUG] Read compute environment %s.\n", input)

	result, err := conn.DescribeComputeEnvironments(input)

	if err != nil {
		return fmt.Errorf("error reading Batch Compute Environment (%s): %w", d.Id(), err)
	}

	if len(result.ComputeEnvironments) == 0 {
		log.Printf("[WARN] Batch Compute Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	computeEnvironment := result.ComputeEnvironments[0]

	d.Set("service_role", computeEnvironment.ServiceRole)
	d.Set("state", computeEnvironment.State)

	if err := d.Set("tags", keyvaluetags.BatchKeyValueTags(computeEnvironment.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("type", computeEnvironment.Type)

	if aws.StringValue(computeEnvironment.Type) == batch.CETypeManaged {
		if err := d.Set("compute_resources", flattenBatchComputeResources(computeEnvironment.ComputeResources)); err != nil {
			return fmt.Errorf("error setting compute_resources: %s", err)
		}
	}

	d.Set("arn", computeEnvironment.ComputeEnvironmentArn)
	d.Set("ecs_cluster_arn", computeEnvironment.EcsClusterArn)
	d.Set("status", computeEnvironment.Status)
	d.Set("status_reason", computeEnvironment.StatusReason)

	return nil
}

func flattenBatchComputeResources(computeResource *batch.ComputeResource) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	m := make(map[string]interface{})

	m["allocation_strategy"] = aws.StringValue(computeResource.AllocationStrategy)
	m["bid_percentage"] = int(aws.Int64Value(computeResource.BidPercentage))
	m["desired_vcpus"] = int(aws.Int64Value(computeResource.DesiredvCpus))
	m["ec2_key_pair"] = aws.StringValue(computeResource.Ec2KeyPair)
	m["image_id"] = aws.StringValue(computeResource.ImageId)
	m["instance_role"] = aws.StringValue(computeResource.InstanceRole)
	m["instance_type"] = schema.NewSet(schema.HashString, flattenStringList(computeResource.InstanceTypes))
	m["max_vcpus"] = int(aws.Int64Value(computeResource.MaxvCpus))
	m["min_vcpus"] = int(aws.Int64Value(computeResource.MinvCpus))
	m["security_group_ids"] = schema.NewSet(schema.HashString, flattenStringList(computeResource.SecurityGroupIds))
	m["spot_iam_fleet_role"] = aws.StringValue(computeResource.SpotIamFleetRole)
	m["subnets"] = schema.NewSet(schema.HashString, flattenStringList(computeResource.Subnets))
	m["tags"] = keyvaluetags.BatchKeyValueTags(computeResource.Tags).IgnoreAws().Map()
	m["type"] = aws.StringValue(computeResource.Type)

	if launchTemplate := computeResource.LaunchTemplate; launchTemplate != nil {
		lt := make(map[string]interface{})
		lt["launch_template_id"] = aws.StringValue(launchTemplate.LaunchTemplateId)
		lt["launch_template_name"] = aws.StringValue(launchTemplate.LaunchTemplateName)
		lt["version"] = aws.StringValue(launchTemplate.Version)
		m["launch_template"] = []map[string]interface{}{lt}
	}

	result = append(result, m)
	return result
}

func resourceAwsBatchComputeEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	computeEnvironmentName := d.Get("compute_environment_name").(string)

	log.Printf("[DEBUG] Disabling Batch Compute Environment: %s", computeEnvironmentName)
	err := disableBatchComputeEnvironment(computeEnvironmentName, d.Timeout(schema.TimeoutDelete), conn)
	if err != nil {
		return fmt.Errorf("error disabling Batch Compute Environment (%s): %s", computeEnvironmentName, err)
	}

	log.Printf("[DEBUG] Deleting Batch Compute Environment: %s", computeEnvironmentName)
	err = deleteBatchComputeEnvironment(computeEnvironmentName, d.Timeout(schema.TimeoutDelete), conn)
	if err != nil {
		return fmt.Errorf("error deleting Batch Compute Environment (%s): %s", computeEnvironmentName, err)
	}

	return nil
}

func resourceAwsBatchComputeEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	if d.HasChanges("compute_resources", "service_role", "state") {
		computeEnvironmentName := d.Get("compute_environment_name").(string)

		input := &batch.UpdateComputeEnvironmentInput{
			ComputeEnvironment: aws.String(computeEnvironmentName),
			ComputeResources:   &batch.ComputeResourceUpdate{},
		}

		if d.HasChange("service_role") {
			input.ServiceRole = aws.String(d.Get("service_role").(string))
		}
		if d.HasChange("state") {
			input.State = aws.String(d.Get("state").(string))
		}

		if d.HasChange("compute_resources") {
			computeResources := d.Get("compute_resources").([]interface{})
			if len(computeResources) == 0 {
				return fmt.Errorf("One compute environment is expected, but no compute environments are set")
			}
			computeResource := computeResources[0].(map[string]interface{})

			if d.HasChange("compute_resources.0.desired_vcpus") {
				input.ComputeResources.DesiredvCpus = aws.Int64(int64(computeResource["desired_vcpus"].(int)))
			}

			input.ComputeResources.MaxvCpus = aws.Int64(int64(computeResource["max_vcpus"].(int)))
			input.ComputeResources.MinvCpus = aws.Int64(int64(computeResource["min_vcpus"].(int)))
		}

		log.Printf("[DEBUG] Update compute environment %s.\n", input)

		if _, err := conn.UpdateComputeEnvironment(input); err != nil {
			return fmt.Errorf("error updating Batch Compute Environment (%s): %w", d.Id(), err)
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{batch.CEStatusUpdating},
			Target:     []string{batch.CEStatusValid},
			Refresh:    resourceAwsBatchComputeEnvironmentStatusRefreshFunc(computeEnvironmentName, conn),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			MinTimeout: 5 * time.Second,
		}

		if _, err := stateConf.WaitForState(); err != nil {
			return fmt.Errorf("error waiting for Batch Compute Environment (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.BatchUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsBatchComputeEnvironmentRead(d, meta)
}

func resourceAwsBatchComputeEnvironmentStatusRefreshFunc(computeEnvironmentName string, conn *batch.Batch) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
			ComputeEnvironments: []*string{
				aws.String(computeEnvironmentName),
			},
		})
		if err != nil {
			return nil, "failed", err
		}

		if len(result.ComputeEnvironments) == 0 {
			return nil, "failed", fmt.Errorf("One compute environment is expected, but AWS return no compute environment")
		}

		computeEnvironment := result.ComputeEnvironments[0]
		return result, *(computeEnvironment.Status), nil
	}
}

func resourceAwsBatchComputeEnvironmentDeleteRefreshFunc(computeEnvironmentName string, conn *batch.Batch) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeComputeEnvironments(&batch.DescribeComputeEnvironmentsInput{
			ComputeEnvironments: []*string{
				aws.String(computeEnvironmentName),
			},
		})
		if err != nil {
			return nil, "failed", err
		}

		if len(result.ComputeEnvironments) == 0 {
			return result, batch.CEStatusDeleted, nil
		}

		computeEnvironment := result.ComputeEnvironments[0]
		return result, *(computeEnvironment.Status), nil
	}
}

func deleteBatchComputeEnvironment(computeEnvironment string, timeout time.Duration, conn *batch.Batch) error {
	input := &batch.DeleteComputeEnvironmentInput{
		ComputeEnvironment: aws.String(computeEnvironment),
	}

	if _, err := conn.DeleteComputeEnvironment(input); err != nil {
		return err
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending:    []string{batch.CEStatusDeleting},
		Target:     []string{batch.CEStatusDeleted},
		Refresh:    resourceAwsBatchComputeEnvironmentDeleteRefreshFunc(computeEnvironment, conn),
		Timeout:    timeout,
		MinTimeout: 5 * time.Second,
	}
	_, err := stateChangeConf.WaitForState()
	return err
}

func disableBatchComputeEnvironment(computeEnvironment string, timeout time.Duration, conn *batch.Batch) error {
	input := &batch.UpdateComputeEnvironmentInput{
		ComputeEnvironment: aws.String(computeEnvironment),
		State:              aws.String(batch.CEStateDisabled),
	}

	if _, err := conn.UpdateComputeEnvironment(input); err != nil {
		return err
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending:    []string{batch.CEStatusUpdating},
		Target:     []string{batch.CEStatusValid},
		Refresh:    resourceAwsBatchComputeEnvironmentStatusRefreshFunc(computeEnvironment, conn),
		Timeout:    timeout,
		MinTimeout: 5 * time.Second,
	}
	_, err := stateChangeConf.WaitForState()
	return err
}
