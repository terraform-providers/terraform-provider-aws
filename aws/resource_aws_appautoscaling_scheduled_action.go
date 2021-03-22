package aws

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/experimental/nullable"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/applicationautoscaling/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsAppautoscalingScheduledAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppautoscalingScheduledActionPut,
		Read:   resourceAwsAppautoscalingScheduledActionRead,
		Update: resourceAwsAppautoscalingScheduledActionPut,
		Delete: resourceAwsAppautoscalingScheduledActionDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_namespace": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scalable_dimension": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scalable_target_action": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_capacity": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
							AtLeastOneOf: []string{
								"scalable_target_action.0.max_capacity",
								"scalable_target_action.0.min_capacity",
							},
						},
						"min_capacity": {
							Type:         nullable.TypeNullableInt,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
							AtLeastOneOf: []string{
								"scalable_target_action.0.max_capacity",
								"scalable_target_action.0.min_capacity",
							},
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			// The AWS API normalizes start_time and end_time to UTC. Uses
			// suppressEquivalentTime to allow any timezone to be used.
			"start_time": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: suppressEquivalentTime,
			},
			"end_time": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.IsRFC3339Time,
				DiffSuppressFunc: suppressEquivalentTime,
			},
			"timezone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "UTC",
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppautoscalingScheduledActionPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := &applicationautoscaling.PutScheduledActionInput{
		ScheduledActionName: aws.String(d.Get("name").(string)),
		ServiceNamespace:    aws.String(d.Get("service_namespace").(string)),
		ResourceId:          aws.String(d.Get("resource_id").(string)),
		ScalableDimension:   aws.String(d.Get("scalable_dimension").(string)),
	}

	needsPut := true
	if d.IsNewResource() {
		appautoscalingScheduledActionPopulateInputForCreate(input, d)
	} else {
		needsPut = appautoscalingScheduledActionPopulateInputForUpdate(input, d)
	}

	if needsPut {
		err := resource.Retry(5*time.Minute, func() *resource.RetryError {
			_, err := conn.PutScheduledAction(input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, applicationautoscaling.ErrCodeObjectNotFoundException) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.PutScheduledAction(input)
		}
		if err != nil {
			return fmt.Errorf("error putting Application Auto Scaling scheduled action: %w", err)
		}

		if d.IsNewResource() {
			d.SetId(d.Get("name").(string) + "-" + d.Get("service_namespace").(string) + "-" + d.Get("resource_id").(string))
		}
	}

	return resourceAwsAppautoscalingScheduledActionRead(d, meta)
}

func appautoscalingScheduledActionPopulateInputForCreate(input *applicationautoscaling.PutScheduledActionInput, d *schema.ResourceData) {
	input.Schedule = aws.String(d.Get("schedule").(string))
	input.ScalableTargetAction = expandScalableTargetAction(d.Get("scalable_target_action").([]interface{}))
	input.Timezone = aws.String(d.Get("timezone").(string))

	if v, ok := d.GetOk("start_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.StartTime = aws.Time(t)
	}
	if v, ok := d.GetOk("end_time"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		input.EndTime = aws.Time(t)
	}
}

func appautoscalingScheduledActionPopulateInputForUpdate(input *applicationautoscaling.PutScheduledActionInput, d *schema.ResourceData) bool {
	hasChange := false

	if d.HasChange("schedule") {
		input.Schedule = aws.String(d.Get("schedule").(string))
		hasChange = true
	}

	if d.HasChange("scalable_target_action") {
		input.ScalableTargetAction = expandScalableTargetAction(d.Get("scalable_target_action").([]interface{}))
		hasChange = true
	}

	if d.HasChange("timezone") {
		input.Timezone = aws.String(d.Get("timezone").(string))
		hasChange = true
	}

	if d.HasChange("start_time") {
		if v, ok := d.GetOk("start_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.StartTime = aws.Time(t)
			hasChange = true
		}
	}
	if d.HasChange("end_time") {
		if v, ok := d.GetOk("end_time"); ok {
			t, _ := time.Parse(time.RFC3339, v.(string))
			input.EndTime = aws.Time(t)
			hasChange = true
		}
	}

	return hasChange
}

func resourceAwsAppautoscalingScheduledActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	scheduledAction, err := finder.ScheduledAction(conn, d.Get("name").(string), d.Get("service_namespace").(string), d.Get("resource_id").(string))
	if tfresource.NotFound(err) {
		log.Printf("[WARN] Application Auto Scaling Scheduled Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error describing Application Auto Scaling Scheduled Action (%s): %w", d.Id(), err)
	}

	if err := d.Set("scalable_target_action", flattenScalableTargetAction(scheduledAction.ScalableTargetAction)); err != nil {
		return fmt.Errorf("error setting scalable_target_action: %w", err)
	}

	d.Set("schedule", scheduledAction.Schedule)
	if scheduledAction.StartTime != nil {
		d.Set("start_time", scheduledAction.StartTime.Format(time.RFC3339))
	}
	if scheduledAction.EndTime != nil {
		d.Set("end_time", scheduledAction.EndTime.Format(time.RFC3339))
	}
	d.Set("timezone", scheduledAction.Timezone)
	d.Set("arn", scheduledAction.ScheduledActionARN)

	return nil
}

func resourceAwsAppautoscalingScheduledActionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := &applicationautoscaling.DeleteScheduledActionInput{
		ScheduledActionName: aws.String(d.Get("name").(string)),
		ServiceNamespace:    aws.String(d.Get("service_namespace").(string)),
		ResourceId:          aws.String(d.Get("resource_id").(string)),
	}
	if v, ok := d.GetOk("scalable_dimension"); ok {
		input.ScalableDimension = aws.String(v.(string))
	}
	_, err := conn.DeleteScheduledAction(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, applicationautoscaling.ErrCodeObjectNotFoundException) {
			log.Printf("[WARN] Application Auto Scaling scheduled action (%s) not found, removing from state", d.Id())
			return nil
		}
		return err
	}

	return nil
}

func expandScalableTargetAction(l []interface{}) *applicationautoscaling.ScalableTargetAction {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	result := &applicationautoscaling.ScalableTargetAction{}

	if v, ok := m["max_capacity"]; ok {
		if v, null, _ := nullable.Int(v.(string)).Value(); !null {
			result.MaxCapacity = aws.Int64(v)
		}
	}
	if v, ok := m["min_capacity"]; ok {
		if v, null, _ := nullable.Int(v.(string)).Value(); !null {
			result.MinCapacity = aws.Int64(v)
		}
	}

	return result
}

func flattenScalableTargetAction(cfg *applicationautoscaling.ScalableTargetAction) []interface{} {
	if cfg == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if cfg.MaxCapacity != nil {
		m["max_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MaxCapacity), 10)
	}
	if cfg.MinCapacity != nil {
		m["min_capacity"] = strconv.FormatInt(aws.Int64Value(cfg.MinCapacity), 10)
	}

	return []interface{}{m}
}
