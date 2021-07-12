package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/gamelift/waiter"
)

func resourceAwsGameliftGameServerGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGameliftGameServerGroupCreate,
		Read:   resourceAwsGameliftGameServerGroupRead,
		Update: resourceAwsGameliftGameServerGroupUpdate,
		Delete: resourceAwsGameliftGameServerGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_policy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"estimated_instance_warmup": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"target_tracking_configuration": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_value": {
										Type:         schema.TypeFloat,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.FloatAtLeast(0),
									},
								},
							},
						},
					},
				},
			},
			"balancing_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(gamelift.BalancingStrategy_Values(), false),
			},
			"game_server_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"game_server_protection_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(gamelift.GameServerProtectionPolicy_Values(), false),
			},
			"instance_definition": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 2,
				MaxItems: 20,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(gamelift.GameServerGroupInstanceType_Values(), false),
						},
						"weighted_capacity": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 3),
						},
					},
				},
			},
			"launch_template": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.name"},
							ValidateFunc:  validateLaunchTemplateId,
						},
						"name": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.id"},
							ValidateFunc:  validateLaunchTemplateName,
						},
						"version": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(1, 128),
						},
					},
				},
			},
			"max_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"min_size": {
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"tags": tagsSchema(),
			"vpc_subnets": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 20,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(15, 24),
				},
			},
		},
	}
}

func resourceAwsGameliftGameServerGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	gameServerGroupName := d.Get("game_server_group_name").(string)

	input := gamelift.CreateGameServerGroupInput{
		GameServerGroupName: aws.String(gameServerGroupName),
		InstanceDefinitions: expandGameliftInstanceDefinitions(d.Get("instance_definition").(*schema.Set).List()),
		LaunchTemplate:      expandGameliftLaunchTemplateSpecification(d.Get("launch_template").([]interface{})[0].(map[string]interface{})),
		MaxSize:             aws.Int64(int64(d.Get("max_size").(int))),
		MinSize:             aws.Int64(int64(d.Get("min_size").(int))),
		RoleArn:             aws.String(d.Get("role_arn").(string)),
		Tags:                keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GameliftTags(),
	}

	if v, ok := d.GetOk("auto_scaling_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoScalingPolicy = expandGameliftGameServerGroupAutoScalingPolicy(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("balancing_strategy"); ok {
		input.BalancingStrategy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("game_server_protection_policy"); ok {
		input.GameServerProtectionPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_subnets"); ok && v.(*schema.Set).Len() > 0 {
		input.VpcSubnets = expandStringSet(v.(*schema.Set))
	}

	_, err := conn.CreateGameServerGroup(&input)

	if err != nil {
		return fmt.Errorf("error creating GameLift Game Server Group: %w", err)
	}

	d.SetId(gameServerGroupName)

	if _, err := waiter.GameServerGroupActive(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Gamelift Game Server Group (%s) to be active: %w", d.Id(), err)
	}

	return resourceAwsGameliftGameServerGroupRead(d, meta)
}

func resourceAwsGameliftGameServerGroupRead(d *schema.ResourceData, meta interface{}) error {
	autoscalingconn := meta.(*AWSClient).autoscalingconn
	gameliftconn := meta.(*AWSClient).gameliftconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	gameServerGroupName := d.Id()

	describeGameServerGroupOutput, err := gameliftconn.DescribeGameServerGroup(&gamelift.DescribeGameServerGroupInput{
		GameServerGroupName: aws.String(gameServerGroupName),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
		log.Printf("[WARN] Gamelift Game Server Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Gamelift Game Server Group (%s): %w", d.Id(), err)
	}

	gameServerGroup := describeGameServerGroupOutput.GameServerGroup

	// To view properties of automatically created Auto Scaling group users should access it directly:
	// https://docs.aws.amazon.com/gamelift/latest/apireference/API_DescribeGameServerGroup.html
	autoScalingGroupName := strings.Split(aws.StringValue(gameServerGroup.AutoScalingGroupArn), "/")[1]
	autoScalingGroup, err := getAwsAutoscalingGroup(autoScalingGroupName, autoscalingconn)

	if err != nil {
		return err
	}

	if autoScalingGroup == nil {
		return fmt.Errorf("error describing Auto Scaling Group (%s): not found", autoScalingGroupName)
	}

	describePoliciesOutput, err := autoscalingconn.DescribePolicies(&autoscaling.DescribePoliciesInput{
		AutoScalingGroupName: aws.String(autoScalingGroupName),
		PolicyNames:          []*string{aws.String(gameServerGroupName)},
	})

	if err != nil {
		return fmt.Errorf("error describing Auto Scaling Group Policies (%s): %s", autoScalingGroupName, err)
	}

	arn := aws.StringValue(gameServerGroup.GameServerGroupArn)
	d.Set("arn", arn)
	d.Set("auto_scaling_group_arn", gameServerGroup.AutoScalingGroupArn)
	d.Set("balancing_strategy", gameServerGroup.BalancingStrategy)
	d.Set("game_server_group_name", gameServerGroupName)
	d.Set("game_server_protection_policy", gameServerGroup.GameServerProtectionPolicy)
	d.Set("max_size", autoScalingGroup.MaxSize)
	d.Set("min_size", autoScalingGroup.MinSize)
	d.Set("role_arn", gameServerGroup.RoleArn)

	// d.Set("vpc_subnets", ...) is absent because Gamelift doesn't return its value in API
	// and dynamically changes autoScalingGroup.VPCZoneIdentifier using its proprietary FleetIQ Spot Viability algorithm.

	if len(describePoliciesOutput.ScalingPolicies) == 1 {
		if err := d.Set("auto_scaling_policy", []interface{}{flattenGameliftGameServerGroupAutoScalingPolicy(describePoliciesOutput.ScalingPolicies[0])}); err != nil {
			return fmt.Errorf("error setting auto_scaling_policy: %w", err)
		}
	} else {
		d.Set("auto_scaling_policy", nil)
	}

	if err := d.Set("instance_definition", flattenGameliftInstanceDefinitions(gameServerGroup.InstanceDefinitions)); err != nil {
		return fmt.Errorf("error setting instance_definition: %s", err)
	}

	if err := d.Set("launch_template", flattenLaunchTemplateSpecification(autoScalingGroup.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification)); err != nil {
		return fmt.Errorf("error setting launch_template: %s", err)
	}

	tags, err := keyvaluetags.GameliftListTags(gameliftconn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Game Lift Game Server Group (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGameliftGameServerGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	gameServerGroupName := d.Id()

	if d.HasChanges("balancing_strategy", "game_server_protection_policy", "instance_definition", "role_arn") {
		input := gamelift.UpdateGameServerGroupInput{
			GameServerGroupName: aws.String(gameServerGroupName),
			InstanceDefinitions: expandGameliftInstanceDefinitions(d.Get("instance_definition").(*schema.Set).List()),
			RoleArn:             aws.String(d.Get("role_arn").(string)),
		}

		if v, ok := d.GetOk("balancing_strategy"); ok {
			input.BalancingStrategy = aws.String(v.(string))
		}

		if v, ok := d.GetOk("game_server_protection_policy"); ok {
			input.GameServerProtectionPolicy = aws.String(v.(string))
		}

		_, err := conn.UpdateGameServerGroup(&input)

		if err != nil {
			return fmt.Errorf("error updating GameLift Game Server Group (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GameliftUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Game Lift Fleet (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsGameliftGameServerGroupRead(d, meta)
}

func resourceAwsGameliftGameServerGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).gameliftconn

	gameServerGroupName := d.Id()

	_, err := conn.DeleteGameServerGroup(&gamelift.DeleteGameServerGroupInput{
		GameServerGroupName: aws.String(gameServerGroupName),
	})

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Gamelift Game Server Group (%s): %w", d.Id(), err)
	}

	_, err = waiter.GameServerGroupDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if tfawserr.ErrCodeEquals(err, gamelift.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for Gamelift Game Server Group (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func expandGameliftGameServerGroupAutoScalingPolicy(tfMap map[string]interface{}) *gamelift.GameServerGroupAutoScalingPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.GameServerGroupAutoScalingPolicy{
		TargetTrackingConfiguration: expandGameliftTargetTrackingConfiguration(tfMap["target_tracking_configuration"].([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := tfMap["estimated_instance_warmup"].(int); ok && v != 0 {
		apiObject.EstimatedInstanceWarmup = aws.Int64(int64(v))
	}

	return apiObject
}

func expandGameliftInstanceDefinition(tfMap map[string]interface{}) *gamelift.InstanceDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.InstanceDefinition{
		InstanceType: aws.String(tfMap["instance_type"].(string)),
	}

	if v, ok := tfMap["weighted_capacity"].(string); ok && v != "" {
		apiObject.WeightedCapacity = aws.String(v)
	}

	return apiObject
}

func expandGameliftInstanceDefinitions(tfList []interface{}) []*gamelift.InstanceDefinition {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*gamelift.InstanceDefinition

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandGameliftInstanceDefinition(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandGameliftLaunchTemplateSpecification(tfMap map[string]interface{}) *gamelift.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.LaunchTemplateSpecification{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandGameliftTargetTrackingConfiguration(tfMap map[string]interface{}) *gamelift.TargetTrackingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &gamelift.TargetTrackingConfiguration{
		TargetValue: aws.Float64(tfMap["target_value"].(float64)),
	}

	return apiObject
}

func flattenGameliftGameServerGroupAutoScalingPolicy(apiObject *autoscaling.ScalingPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_tracking_configuration": []interface{}{flattenGameliftTargetTrackingConfiguration(apiObject.TargetTrackingConfiguration)},
	}

	if v := apiObject.EstimatedInstanceWarmup; v != nil {
		tfMap["estimated_instance_warmup"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenGameliftInstanceDefinition(apiObject *gamelift.InstanceDefinition) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"instance_type": aws.StringValue(apiObject.InstanceType),
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenGameliftInstanceDefinitions(apiObjects []*gamelift.InstanceDefinition) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenGameliftInstanceDefinition(apiObject))
	}

	return tfList
}

func flattenGameliftTargetTrackingConfiguration(apiObject *autoscaling.TargetTrackingConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_value": aws.Float64Value(apiObject.TargetValue),
	}

	return tfMap
}
