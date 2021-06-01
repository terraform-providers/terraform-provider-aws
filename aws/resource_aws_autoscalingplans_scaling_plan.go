package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/autoscalingplans/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/autoscalingplans/waiter"
)

func resourceAwsAutoScalingPlansScalingPlan() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAutoScalingPlansScalingPlanCreate,
		Read:   resourceAwsAutoScalingPlansScalingPlanRead,
		Update: resourceAwsAutoScalingPlansScalingPlanUpdate,
		Delete: resourceAwsAutoScalingPlansScalingPlanDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAutoScalingPlansScalingPlanImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[[:print:]]+$`), "must be printable"),
					validation.StringDoesNotContainAny("|:/"),
				),
			},

			"application_source": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudformation_stack_arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  validateArn,
							ConflictsWith: []string{"application_source.0.tag_filter"},
						},

						"tag_filter": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 0,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Required: true,
									},

									"values": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										MaxItems: 50,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},
								},
							},
							ConflictsWith: []string{"application_source.0.cloudformation_stack_arn"},
						},
					},
				},
			},

			"scaling_instruction": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"customized_load_metric_specification": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dimensions": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									"metric_name": {
										Type:     schema.TypeString,
										Required: true,
									},

									"namespace": {
										Type:     schema.TypeString,
										Required: true,
									},

									"statistic": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											autoscalingplans.MetricStatisticSum,
										}, false),
									},

									"unit": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},

						"disable_dynamic_scaling": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},

						"max_capacity": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"min_capacity": {
							Type:     schema.TypeInt,
							Required: true,
						},

						"predefined_load_metric_specification": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"predefined_load_metric_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(autoscalingplans.LoadMetricType_Values(), false),
									},

									"resource_label": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1023),
									},
								},
							},
						},

						"predictive_scaling_max_capacity_behavior": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(autoscalingplans.PredictiveScalingMaxCapacityBehavior_Values(), false),
						},

						"predictive_scaling_max_capacity_buffer": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 100),
						},

						"predictive_scaling_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(autoscalingplans.PredictiveScalingMode_Values(), false),
						},

						"resource_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1600),
						},

						"scalable_dimension": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(autoscalingplans.ScalableDimension_Values(), false),
						},

						"scaling_policy_update_behavior": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      autoscalingplans.ScalingPolicyUpdateBehaviorKeepExternalPolicies,
							ValidateFunc: validation.StringInSlice(autoscalingplans.ScalingPolicyUpdateBehavior_Values(), false),
						},

						"scheduled_action_buffer_time": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},

						"service_namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(autoscalingplans.ServiceNamespace_Values(), false),
						},

						"target_tracking_configuration": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							MaxItems: 10,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"customized_scaling_metric_specification": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"dimensions": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												"metric_name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"namespace": {
													Type:     schema.TypeString,
													Required: true,
												},

												"statistic": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(autoscalingplans.MetricStatistic_Values(), false),
												},

												"unit": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},

									"disable_scale_in": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},

									"estimated_instance_warmup": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"predefined_scaling_metric_specification": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"predefined_scaling_metric_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(autoscalingplans.ScalingMetricType_Values(), false),
												},

												"resource_label": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 1023),
												},
											},
										},
									},

									"scale_in_cooldown": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"scale_out_cooldown": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"target_value": {
										Type:         schema.TypeFloat,
										Required:     true,
										ValidateFunc: validation.FloatBetween(8.515920e-109, 1.174271e+108),
									},
								},
							},
						},
					},
				},
			},

			"scaling_plan_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsAutoScalingPlansScalingPlanCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingplansconn

	scalingPlanName := d.Get("name").(string)

	input := &autoscalingplans.CreateScalingPlanInput{
		ApplicationSource:   expandAutoScalingPlansApplicationSource(d.Get("application_source").([]interface{})),
		ScalingInstructions: expandAutoScalingPlansScalingInstructions(d.Get("scaling_instruction").(*schema.Set)),
		ScalingPlanName:     aws.String(scalingPlanName),
	}

	log.Printf("[DEBUG] Creating Auto Scaling Scaling Plan: %s", input)
	output, err := conn.CreateScalingPlan(input)
	if err != nil {
		return fmt.Errorf("error creating Auto Scaling Scaling Plan: %w", err)
	}

	scalingPlanVersion := int(aws.Int64Value(output.ScalingPlanVersion))
	d.SetId(autoScalingPlansScalingPlanId(scalingPlanName, scalingPlanVersion))
	d.Set("scaling_plan_version", scalingPlanVersion)

	_, err = waiter.ScalingPlanCreated(conn, scalingPlanName, scalingPlanVersion)
	if err != nil {
		return fmt.Errorf("error waiting for Auto Scaling Scaling Plan (%s) to be created: %w", d.Id(), err)
	}

	return resourceAwsAutoScalingPlansScalingPlanRead(d, meta)
}

func resourceAwsAutoScalingPlansScalingPlanRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingplansconn

	scalingPlan, err := finder.ScalingPlan(conn, d.Get("name").(string), d.Get("scaling_plan_version").(int))
	if err != nil {
		return fmt.Errorf("error reading Auto Scaling Scaling Plan (%s): %w", d.Id(), err)
	}
	if scalingPlan == nil {
		log.Printf("[WARN] Auto Scaling Scaling Plan (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	err = d.Set("application_source", flattenAutoScalingPlansApplicationSource(scalingPlan.ApplicationSource))
	if err != nil {
		return fmt.Errorf("error setting application_source: %w", err)
	}
	d.Set("name", scalingPlan.ScalingPlanName)
	err = d.Set("scaling_instruction", flattenAutoScalingPlansScalingInstructions(scalingPlan.ScalingInstructions))
	if err != nil {
		return fmt.Errorf("error setting application_source: %w", err)
	}
	d.Set("scaling_plan_version", int(aws.Int64Value(scalingPlan.ScalingPlanVersion)))

	return nil
}

func resourceAwsAutoScalingPlansScalingPlanUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingplansconn

	scalingPlanName := d.Get("name").(string)
	scalingPlanVersion := d.Get("scaling_plan_version").(int)

	input := &autoscalingplans.UpdateScalingPlanInput{
		ApplicationSource:   expandAutoScalingPlansApplicationSource(d.Get("application_source").([]interface{})),
		ScalingInstructions: expandAutoScalingPlansScalingInstructions(d.Get("scaling_instruction").(*schema.Set)),
		ScalingPlanName:     aws.String(scalingPlanName),
		ScalingPlanVersion:  aws.Int64(int64(scalingPlanVersion)),
	}

	log.Printf("[DEBUG] Updating Auto Scaling Scaling Plan: %s", input)
	_, err := conn.UpdateScalingPlan(input)
	if err != nil {
		return fmt.Errorf("error updating Auto Scaling Scaling Plan (%s): %w", d.Id(), err)
	}

	_, err = waiter.ScalingPlanUpdated(conn, scalingPlanName, scalingPlanVersion)
	if err != nil {
		return fmt.Errorf("error waiting for Auto Scaling Scaling Plan (%s) to be updated: %w", d.Id(), err)
	}

	return resourceAwsAutoScalingPlansScalingPlanRead(d, meta)
}

func resourceAwsAutoScalingPlansScalingPlanDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).autoscalingplansconn

	scalingPlanName := d.Get("name").(string)
	scalingPlanVersion := d.Get("scaling_plan_version").(int)

	log.Printf("[DEBUG] Deleting Auto Scaling Scaling Plan: %s", d.Id())
	_, err := conn.DeleteScalingPlan(&autoscalingplans.DeleteScalingPlanInput{
		ScalingPlanName:    aws.String(scalingPlanName),
		ScalingPlanVersion: aws.Int64(int64(scalingPlanVersion)),
	})
	if isAWSErr(err, autoscalingplans.ErrCodeObjectNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Auto Scaling Scaling Plan (%s): %w", d.Id(), err)
	}

	_, err = waiter.ScalingPlanDeleted(conn, scalingPlanName, scalingPlanVersion)
	if err != nil {
		return fmt.Errorf("error waiting for Auto Scaling Scaling Plan (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func resourceAwsAutoScalingPlansScalingPlanImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	scalingPlanName := d.Id()
	scalingPlanVersion := 1

	d.SetId(autoScalingPlansScalingPlanId(scalingPlanName, scalingPlanVersion))
	d.Set("name", scalingPlanName)
	d.Set("scaling_plan_version", scalingPlanVersion)

	return []*schema.ResourceData{d}, nil
}

// Terraform resource ID.
func autoScalingPlansScalingPlanId(scalingPlanName string, scalingPlanVersion int) string {
	return fmt.Sprintf("%s/%d", scalingPlanName, scalingPlanVersion)
}

//
// ApplicationSource functions.
//

func expandAutoScalingPlansApplicationSource(vApplicationSource []interface{}) *autoscalingplans.ApplicationSource {
	if len(vApplicationSource) == 0 || vApplicationSource[0] == nil {
		return nil
	}
	mApplicationSource := vApplicationSource[0].(map[string]interface{})

	applicationSource := &autoscalingplans.ApplicationSource{}

	if v, ok := mApplicationSource["cloudformation_stack_arn"].(string); ok && v != "" {
		applicationSource.CloudFormationStackARN = aws.String(v)
	}

	if vTagFilters, ok := mApplicationSource["tag_filter"].(*schema.Set); ok && vTagFilters.Len() > 0 {
		tagFilters := []*autoscalingplans.TagFilter{}

		for _, vTagFilter := range vTagFilters.List() {
			tagFilter := &autoscalingplans.TagFilter{}

			mTagFilter := vTagFilter.(map[string]interface{})

			if v, ok := mTagFilter["key"].(string); ok && v != "" {
				tagFilter.Key = aws.String(v)
			}

			if vValues, ok := mTagFilter["values"].(*schema.Set); ok && vValues.Len() > 0 {
				tagFilter.Values = expandStringSet(vValues)
			}

			tagFilters = append(tagFilters, tagFilter)
		}

		applicationSource.TagFilters = tagFilters
	}

	return applicationSource
}

func flattenAutoScalingPlansApplicationSource(applicationSource *autoscalingplans.ApplicationSource) []interface{} {
	if applicationSource == nil {
		return []interface{}{}
	}

	mApplicationSource := map[string]interface{}{
		"cloudformation_stack_arn": aws.StringValue(applicationSource.CloudFormationStackARN),
	}

	if tagFilters := applicationSource.TagFilters; tagFilters != nil {
		vTagFilters := []interface{}{}

		for _, tagFilter := range tagFilters {
			mTagFilter := map[string]interface{}{
				"key":    aws.StringValue(tagFilter.Key),
				"values": flattenStringSet(tagFilter.Values),
			}

			vTagFilters = append(vTagFilters, mTagFilter)
		}

		mApplicationSource["tag_filter"] = vTagFilters
	}

	return []interface{}{mApplicationSource}
}

//
// ScalingInstruction functions.
//

func expandAutoScalingPlansScalingInstructions(vScalingInstructions *schema.Set) []*autoscalingplans.ScalingInstruction {
	scalingInstructions := []*autoscalingplans.ScalingInstruction{}

	for _, vScalingInstruction := range vScalingInstructions.List() {
		mScalingInstruction := vScalingInstruction.(map[string]interface{})

		scalingInstruction := &autoscalingplans.ScalingInstruction{}

		if v, ok := mScalingInstruction["disable_dynamic_scaling"].(bool); ok {
			scalingInstruction.DisableDynamicScaling = aws.Bool(v)
		}
		if v, ok := mScalingInstruction["max_capacity"].(int); ok {
			scalingInstruction.MaxCapacity = aws.Int64(int64(v))
		}
		if v, ok := mScalingInstruction["min_capacity"].(int); ok {
			scalingInstruction.MinCapacity = aws.Int64(int64(v))
		}
		if v, ok := mScalingInstruction["predictive_scaling_max_capacity_behavior"].(string); ok && v != "" {
			scalingInstruction.PredictiveScalingMaxCapacityBehavior = aws.String(v)
		}
		if v, ok := mScalingInstruction["predictive_scaling_max_capacity_buffer"].(int); ok && v > 0 {
			scalingInstruction.PredictiveScalingMaxCapacityBuffer = aws.Int64(int64(v))
		}
		if v, ok := mScalingInstruction["predictive_scaling_mode"].(string); ok && v != "" {
			scalingInstruction.PredictiveScalingMode = aws.String(v)
		}
		if v, ok := mScalingInstruction["resource_id"].(string); ok && v != "" {
			scalingInstruction.ResourceId = aws.String(v)
		}
		if v, ok := mScalingInstruction["scalable_dimension"].(string); ok && v != "" {
			scalingInstruction.ScalableDimension = aws.String(v)
		}
		if v, ok := mScalingInstruction["scaling_policy_update_behavior"].(string); ok && v != "" {
			scalingInstruction.ScalingPolicyUpdateBehavior = aws.String(v)
		}
		if v, ok := mScalingInstruction["scheduled_action_buffer_time"].(int); ok && v > 0 {
			scalingInstruction.ScheduledActionBufferTime = aws.Int64(int64(v))
		}
		if v, ok := mScalingInstruction["service_namespace"].(string); ok && v != "" {
			scalingInstruction.ServiceNamespace = aws.String(v)
		}

		if vCustomizedLoadMetricSpecification, ok := mScalingInstruction["customized_load_metric_specification"].([]interface{}); ok && len(vCustomizedLoadMetricSpecification) > 0 && vCustomizedLoadMetricSpecification[0] != nil {
			mCustomizedLoadMetricSpecification := vCustomizedLoadMetricSpecification[0].(map[string]interface{})

			customizedLoadMetricSpecification := &autoscalingplans.CustomizedLoadMetricSpecification{}

			if v, ok := mCustomizedLoadMetricSpecification["dimensions"].(map[string]interface{}); ok {
				dimensions := []*autoscalingplans.MetricDimension{}

				for key, value := range v {
					dimension := &autoscalingplans.MetricDimension{}

					dimension.Name = aws.String(key)
					dimension.Value = aws.String(value.(string))

					dimensions = append(dimensions, dimension)
				}

				customizedLoadMetricSpecification.Dimensions = dimensions
			}
			if v, ok := mCustomizedLoadMetricSpecification["metric_name"].(string); ok && v != "" {
				customizedLoadMetricSpecification.MetricName = aws.String(v)
			}
			if v, ok := mCustomizedLoadMetricSpecification["namespace"].(string); ok && v != "" {
				customizedLoadMetricSpecification.Namespace = aws.String(v)
			}
			if v, ok := mCustomizedLoadMetricSpecification["statistic"].(string); ok && v != "" {
				customizedLoadMetricSpecification.Statistic = aws.String(v)
			}
			if v, ok := mCustomizedLoadMetricSpecification["unit"].(string); ok && v != "" {
				customizedLoadMetricSpecification.Unit = aws.String(v)
			}

			scalingInstruction.CustomizedLoadMetricSpecification = customizedLoadMetricSpecification
		}

		if vPredefinedLoadMetricSpecification, ok := mScalingInstruction["predefined_load_metric_specification"].([]interface{}); ok && len(vPredefinedLoadMetricSpecification) > 0 && vPredefinedLoadMetricSpecification[0] != nil {
			mPredefinedLoadMetricSpecification := vPredefinedLoadMetricSpecification[0].(map[string]interface{})

			predefinedLoadMetricSpecification := &autoscalingplans.PredefinedLoadMetricSpecification{}

			if v, ok := mPredefinedLoadMetricSpecification["predefined_load_metric_type"].(string); ok && v != "" {
				predefinedLoadMetricSpecification.PredefinedLoadMetricType = aws.String(v)
			}
			if v, ok := mPredefinedLoadMetricSpecification["resource_label"].(string); ok && v != "" {
				predefinedLoadMetricSpecification.ResourceLabel = aws.String(v)
			}

			scalingInstruction.PredefinedLoadMetricSpecification = predefinedLoadMetricSpecification
		}

		if vTargetTrackingConfigurations, ok := mScalingInstruction["target_tracking_configuration"].(*schema.Set); ok && vTargetTrackingConfigurations.Len() > 0 {
			targetTrackingConfigurations := []*autoscalingplans.TargetTrackingConfiguration{}

			for _, vTargetTrackingConfiguration := range vTargetTrackingConfigurations.List() {
				targetTrackingConfiguration := &autoscalingplans.TargetTrackingConfiguration{}

				mTargetTrackingConfiguration := vTargetTrackingConfiguration.(map[string]interface{})

				if v, ok := mTargetTrackingConfiguration["disable_scale_in"].(bool); ok {
					targetTrackingConfiguration.DisableScaleIn = aws.Bool(v)
				}
				if v, ok := mTargetTrackingConfiguration["estimated_instance_warmup"].(int); ok && v > 0 {
					targetTrackingConfiguration.EstimatedInstanceWarmup = aws.Int64(int64(v))
				}
				if v, ok := mTargetTrackingConfiguration["scale_in_cooldown"].(int); ok && v > 0 {
					targetTrackingConfiguration.ScaleInCooldown = aws.Int64(int64(v))
				}
				if v, ok := mTargetTrackingConfiguration["scale_out_cooldown"].(int); ok && v > 0 {
					targetTrackingConfiguration.ScaleOutCooldown = aws.Int64(int64(v))
				}
				if v, ok := mTargetTrackingConfiguration["target_value"].(float64); ok && v > 0.0 {
					targetTrackingConfiguration.TargetValue = aws.Float64(v)
				}

				if vCustomizedScalingMetricSpecification, ok := mTargetTrackingConfiguration["customized_scaling_metric_specification"].([]interface{}); ok && len(vCustomizedScalingMetricSpecification) > 0 && vCustomizedScalingMetricSpecification[0] != nil {
					mCustomizedScalingMetricSpecification := vCustomizedScalingMetricSpecification[0].(map[string]interface{})

					customizedScalingMetricSpecification := &autoscalingplans.CustomizedScalingMetricSpecification{}

					if v, ok := mCustomizedScalingMetricSpecification["dimensions"].(map[string]interface{}); ok {
						dimensions := []*autoscalingplans.MetricDimension{}

						for key, value := range v {
							dimension := &autoscalingplans.MetricDimension{}

							dimension.Name = aws.String(key)
							dimension.Value = aws.String(value.(string))

							dimensions = append(dimensions, dimension)
						}

						customizedScalingMetricSpecification.Dimensions = dimensions
					}
					if v, ok := mCustomizedScalingMetricSpecification["metric_name"].(string); ok && v != "" {
						customizedScalingMetricSpecification.MetricName = aws.String(v)
					}
					if v, ok := mCustomizedScalingMetricSpecification["namespace"].(string); ok && v != "" {
						customizedScalingMetricSpecification.Namespace = aws.String(v)
					}
					if v, ok := mCustomizedScalingMetricSpecification["statistic"].(string); ok && v != "" {
						customizedScalingMetricSpecification.Statistic = aws.String(v)
					}
					if v, ok := mCustomizedScalingMetricSpecification["unit"].(string); ok && v != "" {
						customizedScalingMetricSpecification.Unit = aws.String(v)
					}

					targetTrackingConfiguration.CustomizedScalingMetricSpecification = customizedScalingMetricSpecification
				}

				if vPredefinedScalingMetricSpecification, ok := mTargetTrackingConfiguration["predefined_scaling_metric_specification"].([]interface{}); ok && len(vPredefinedScalingMetricSpecification) > 0 && vPredefinedScalingMetricSpecification[0] != nil {
					mPredefinedScalingMetricSpecification := vPredefinedScalingMetricSpecification[0].(map[string]interface{})

					predefinedScalingMetricSpecification := &autoscalingplans.PredefinedScalingMetricSpecification{}

					if v, ok := mPredefinedScalingMetricSpecification["predefined_scaling_metric_type"].(string); ok && v != "" {
						predefinedScalingMetricSpecification.PredefinedScalingMetricType = aws.String(v)
					}
					if v, ok := mPredefinedScalingMetricSpecification["resource_label"].(string); ok && v != "" {
						predefinedScalingMetricSpecification.ResourceLabel = aws.String(v)
					}

					targetTrackingConfiguration.PredefinedScalingMetricSpecification = predefinedScalingMetricSpecification
				}

				targetTrackingConfigurations = append(targetTrackingConfigurations, targetTrackingConfiguration)
			}

			scalingInstruction.TargetTrackingConfigurations = targetTrackingConfigurations
		}

		scalingInstructions = append(scalingInstructions, scalingInstruction)
	}

	return scalingInstructions
}

func flattenAutoScalingPlansScalingInstructions(scalingInstructions []*autoscalingplans.ScalingInstruction) []interface{} {
	vScalingInstructions := []interface{}{}

	for _, scalingInstruction := range scalingInstructions {
		mScalingInstruction := map[string]interface{}{
			"disable_dynamic_scaling":                  aws.BoolValue(scalingInstruction.DisableDynamicScaling),
			"max_capacity":                             int(aws.Int64Value(scalingInstruction.MaxCapacity)),
			"min_capacity":                             int(aws.Int64Value(scalingInstruction.MinCapacity)),
			"predictive_scaling_max_capacity_behavior": aws.StringValue(scalingInstruction.PredictiveScalingMaxCapacityBehavior),
			"predictive_scaling_max_capacity_buffer":   int(aws.Int64Value(scalingInstruction.PredictiveScalingMaxCapacityBuffer)),
			"predictive_scaling_mode":                  aws.StringValue(scalingInstruction.PredictiveScalingMode),
			"resource_id":                              aws.StringValue(scalingInstruction.ResourceId),
			"scalable_dimension":                       aws.StringValue(scalingInstruction.ScalableDimension),
			"scaling_policy_update_behavior":           aws.StringValue(scalingInstruction.ScalingPolicyUpdateBehavior),
			"scheduled_action_buffer_time":             int(aws.Int64Value(scalingInstruction.ScheduledActionBufferTime)),
			"service_namespace":                        aws.StringValue(scalingInstruction.ServiceNamespace),
		}

		if customizedLoadMetricSpecification := scalingInstruction.CustomizedLoadMetricSpecification; customizedLoadMetricSpecification != nil {
			mDimensions := map[string]interface{}{}
			for _, dimension := range customizedLoadMetricSpecification.Dimensions {
				mDimensions[aws.StringValue(dimension.Name)] = aws.StringValue(dimension.Value)
			}

			mScalingInstruction["customized_load_metric_specification"] = []interface{}{
				map[string]interface{}{
					"dimensions":  mDimensions,
					"metric_name": aws.StringValue(customizedLoadMetricSpecification.MetricName),
					"namespace":   aws.StringValue(customizedLoadMetricSpecification.Namespace),
					"statistic":   aws.StringValue(customizedLoadMetricSpecification.Statistic),
					"unit":        aws.StringValue(customizedLoadMetricSpecification.Unit),
				},
			}
		}

		if predefinedLoadMetricSpecification := scalingInstruction.PredefinedLoadMetricSpecification; predefinedLoadMetricSpecification != nil {
			mScalingInstruction["predefined_load_metric_specification"] = []interface{}{
				map[string]interface{}{
					"predefined_load_metric_type": aws.StringValue(predefinedLoadMetricSpecification.PredefinedLoadMetricType),
					"resource_label":              aws.StringValue(predefinedLoadMetricSpecification.ResourceLabel),
				},
			}
		}

		if targetTrackingConfigurations := scalingInstruction.TargetTrackingConfigurations; targetTrackingConfigurations != nil {
			vTargetTrackingConfigurations := []interface{}{}

			for _, targetTrackingConfiguration := range targetTrackingConfigurations {
				mTargetTrackingConfiguration := map[string]interface{}{
					"disable_scale_in":          aws.BoolValue(targetTrackingConfiguration.DisableScaleIn),
					"estimated_instance_warmup": int(aws.Int64Value(targetTrackingConfiguration.EstimatedInstanceWarmup)),
					"scale_in_cooldown":         int(aws.Int64Value(targetTrackingConfiguration.ScaleInCooldown)),
					"scale_out_cooldown":        int(aws.Int64Value(targetTrackingConfiguration.ScaleOutCooldown)),
					"target_value":              aws.Float64Value(targetTrackingConfiguration.TargetValue),
				}

				if customizedScalingMetricSpecification := targetTrackingConfiguration.CustomizedScalingMetricSpecification; customizedScalingMetricSpecification != nil {
					mDimensions := map[string]interface{}{}
					for _, dimension := range customizedScalingMetricSpecification.Dimensions {
						mDimensions[aws.StringValue(dimension.Name)] = aws.StringValue(dimension.Value)
					}

					mTargetTrackingConfiguration["customized_scaling_metric_specification"] = []interface{}{
						map[string]interface{}{
							"dimensions":  mDimensions,
							"metric_name": aws.StringValue(customizedScalingMetricSpecification.MetricName),
							"namespace":   aws.StringValue(customizedScalingMetricSpecification.Namespace),
							"statistic":   aws.StringValue(customizedScalingMetricSpecification.Statistic),
							"unit":        aws.StringValue(customizedScalingMetricSpecification.Unit),
						},
					}
				}

				if predefinedScalingMetricSpecification := targetTrackingConfiguration.PredefinedScalingMetricSpecification; predefinedScalingMetricSpecification != nil {
					mTargetTrackingConfiguration["predefined_scaling_metric_specification"] = []interface{}{
						map[string]interface{}{
							"predefined_scaling_metric_type": aws.StringValue(predefinedScalingMetricSpecification.PredefinedScalingMetricType),
							"resource_label":                 aws.StringValue(predefinedScalingMetricSpecification.ResourceLabel),
						},
					}
				}

				vTargetTrackingConfigurations = append(vTargetTrackingConfigurations, mTargetTrackingConfiguration)
			}

			mScalingInstruction["target_tracking_configuration"] = vTargetTrackingConfigurations
		}

		vScalingInstructions = append(vScalingInstructions, mScalingInstruction)
	}

	return vScalingInstructions
}
