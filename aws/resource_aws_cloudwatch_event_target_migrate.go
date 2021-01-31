package aws

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfevents "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatchevents"
)

func resourceAwsCloudWatchEventTargetV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"batch_target": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"array_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"job_attempts": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"job_definition": {
							Type:     schema.TypeString,
							Required: true,
						},
						"job_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"ecs_target": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"launch_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"assign_public_ip": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"security_groups": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"subnets": {
										Type:     schema.TypeSet,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"platform_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"task_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"task_definition_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"input": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"input_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"input_transformer": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_paths": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"input_template": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"kinesis_target": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"partition_key_path": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"run_command_targets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"sqs_target": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message_group_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"target_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsCloudWatchEventTargetStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}

	if _, ok := rawState["event_bus_name"]; !ok {
		rawState["event_bus_name"] = tfevents.DefaultEventBusName
	}

	return rawState, nil
}
