package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsEcsRunTask() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsRunTaskRead,

		Schema: map[string]*schema.Schema{
			"capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},

						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},

						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},

			"cluster": {
				Type:     schema.TypeString,
				Required: true,
			},

			"desired_count": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},

			"enable_ecs_managed_tags": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"group": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"launch_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.LaunchTypeEc2,
					ecs.LaunchTypeFargate,
				}, false),
			},

			"network_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"overrides": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_overrides": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"command": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									"cpu": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"environment": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},

									"memory": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"memory_reservation": {
										Type:     schema.TypeInt,
										Optional: true,
									},

									"name": {
										Type:     schema.TypeString,
										Required: true,
									},

									"resource_requirements": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},

						"cpu": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"execution_role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"inference_accelerator_overrides": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"device_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"device_type": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},

						"memory": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"task_role_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"placement_constraints": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 10,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								ecs.PlacementConstraintTypeDistinctInstance,
								ecs.PlacementConstraintTypeMemberOf,
							}, false),
						},
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"placement_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"field": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"platform_version": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"propagate_tags": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ecs.PropagateTagsService,
					ecs.PropagateTagsTaskDefinition,
				}, false),
			},

			"reference_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"started_by": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"tags": tagsSchema(),

			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsEcsRunTaskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	input := ecs.RunTaskInput{
		Cluster:        aws.String(d.Get("cluster").(string)),
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
		Tags:           keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EcsTags(),
	}

	if v, ok := d.GetOk("launch_type"); ok {
		input.LaunchType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("propagate_tags"); ok {
		input.PropagateTags = aws.String(v.(string))
	}

	if v, ok := d.GetOk("platform_version"); ok {
		input.PlatformVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("desired_count"); ok {
		input.Count = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("started_by"); ok {
		input.StartedBy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("group"); ok {
		input.Group = aws.String(v.(string))
	}

	if v, ok := d.GetOk("reference_id"); ok {
		input.ReferenceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("network_configuration"); ok {
		input.NetworkConfiguration = expandEcsNetworkConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("placement_constraints"); ok {
		pc, err := expandPlacementConstraints(v.(*schema.Set).List())
		if err != nil {
			return err
		}
		input.PlacementConstraints = pc
	}

	if v, ok := d.GetOk("capacity_provider_strategy"); ok {
		input.CapacityProviderStrategy = expandEcsCapacityProviderStrategy(v.(*schema.Set))
	}

	if v, ok := d.GetOk("placement_strategy"); ok {
		ps, err := expandPlacementStrategy(v.([]interface{}))
		if err != nil {
			return err
		}
		input.PlacementStrategy = ps
	}

	if v, ok := d.GetOk("overrides"); ok {
		input.Overrides = expandEcsTaskOverride(v.([]interface{}))
	}

	_, err := conn.RunTask(&input)

	if err != nil {
		return err
	}

	return nil
}

func expandEcsTaskOverride(to []interface{}) *ecs.TaskOverride {
	if len(to) == 0 {
		return nil
	}
	taskOverride := &ecs.TaskOverride{}
	raw := to[0].(map[string]interface{})

	if v, ok := raw["cpu"]; ok {
		taskOverride.Cpu = aws.String(v.(string))
	}

	if v, ok := raw["memory"]; ok {
		taskOverride.Memory = aws.String(v.(string))
	}

	if v, ok := raw["execution_role_arn"]; ok {
		taskOverride.ExecutionRoleArn = aws.String(v.(string))
	}

	if v, ok := raw["task_role_arn"]; ok {
		taskOverride.TaskRoleArn = aws.String(v.(string))
	}

	if v, ok := raw["inference_accelerator_overrides"]; ok {
		taskOverride.InferenceAcceleratorOverrides = expandEcsInferenceAcceleratorOverrides(v.([]interface{}))
	}

	if v, ok := raw["container_overrides"]; ok {
		taskOverride.ContainerOverrides = expandEcsContainerOverride(v.([]interface{}))
	}

	return taskOverride
}

func expandEcsInferenceAcceleratorOverrides(i []interface{}) []*ecs.InferenceAcceleratorOverride {
	if len(i) == 0 {
		return nil
	}
	rrs := make([]*ecs.InferenceAcceleratorOverride, 0)
	for _, item := range i {
		raw := item.(map[string]interface{})
		rr := &ecs.InferenceAcceleratorOverride{
			DeviceName: aws.String(raw["device_name"].(string)),
			DeviceType: aws.String(raw["device_type"].(string)),
		}
		rrs = append(rrs, rr)
	}
	return rrs
}

func expandEcsContainerOverride(o []interface{}) []*ecs.ContainerOverride {
	if len(o) == 0 {
		return nil
	}
	cos := make([]*ecs.ContainerOverride, 0)
	for _, item := range o {
		raw := item.(map[string]interface{})
		override := &ecs.ContainerOverride{
			Name: aws.String(raw["name"].(string)),
		}
		if v, ok := raw["command"]; ok {
			commandStrings := v.([]interface{})
			override.Command = expandStringList(commandStrings)
		}
		if v, ok := raw["cpu"]; ok {
			override.Cpu = aws.Int64(v.(int64))
		}
		if v, ok := raw["environment"]; ok {
			override.Environment = expandEcsTaskEnvironment(v.([]interface{}))
		}
		if v, ok := raw["memory"]; ok {
			override.Memory = aws.Int64(v.(int64))
		}
		if v, ok := raw["memory_reservation"]; ok {
			override.Memory = aws.Int64(v.(int64))
		}
		if v, ok := raw["resource_requirements"]; ok {
			override.ResourceRequirements = expandEcsResourceRequirements(v.([]interface{}))
		}
		cos = append(cos, override)
	}
	return cos
}

func expandEcsTaskEnvironment(e []interface{}) []*ecs.KeyValuePair {
	if len(e) == 0 {
		return nil
	}
	tes := make([]*ecs.KeyValuePair, 0)
	for _, item := range e {
		raw := item.(map[string]interface{})
		te := &ecs.KeyValuePair{
			Name:  aws.String(raw["name"].(string)),
			Value: aws.String(raw["value"].(string)),
		}
		tes = append(tes, te)
	}
	return tes
}

func expandEcsResourceRequirements(r []interface{}) []*ecs.ResourceRequirement {
	if len(r) == 0 {
		return nil
	}
	rrs := make([]*ecs.ResourceRequirement, 0)
	for _, item := range r {
		raw := item.(map[string]interface{})
		rr := &ecs.ResourceRequirement{
			Type:  aws.String(raw["type"].(string)),
			Value: aws.String(raw["value"].(string)),
		}
		rrs = append(rrs, rr)
	}
	return rrs
}
