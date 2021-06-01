package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatch/finder"
)

func resourceAwsCloudWatchCompositeAlarm() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsCloudWatchCompositeAlarmCreate,
		ReadContext:   resourceAwsCloudWatchCompositeAlarmRead,
		UpdateContext: resourceAwsCloudWatchCompositeAlarmUpdate,
		DeleteContext: resourceAwsCloudWatchCompositeAlarmDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"actions_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"alarm_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"alarm_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"alarm_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"alarm_rule": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 10240),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"insufficient_data_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"ok_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      schema.HashString,
				MaxItems: 5,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateArn,
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsCloudWatchCompositeAlarmCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchconn
	name := d.Get("alarm_name").(string)

	input := expandAwsCloudWatchPutCompositeAlarmInput(d)

	_, err := conn.PutCompositeAlarmWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("error creating CloudWatch Composite Alarm (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceAwsCloudWatchCompositeAlarmRead(ctx, d, meta)
}

func resourceAwsCloudWatchCompositeAlarmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	name := d.Id()

	alarm, err := finder.CompositeAlarmByName(ctx, conn, name)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
		log.Printf("[WARN] CloudWatch Composite Alarm %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading CloudWatch Composite Alarm (%s): %s", name, err)
	}

	if alarm == nil {
		if d.IsNewResource() {
			return diag.Errorf("error reading CloudWatch Composite Alarm (%s): not found", name)
		}

		log.Printf("[WARN] CloudWatch Composite Alarm %s not found, removing from state", name)
		d.SetId("")
		return nil
	}

	d.Set("actions_enabled", alarm.ActionsEnabled)

	if err := d.Set("alarm_actions", flattenStringSet(alarm.AlarmActions)); err != nil {
		return diag.Errorf("error setting alarm_actions: %s", err)
	}

	d.Set("alarm_description", alarm.AlarmDescription)
	d.Set("alarm_name", alarm.AlarmName)
	d.Set("alarm_rule", alarm.AlarmRule)
	d.Set("arn", alarm.AlarmArn)

	if err := d.Set("insufficient_data_actions", flattenStringSet(alarm.InsufficientDataActions)); err != nil {
		return diag.Errorf("error setting insufficient_data_actions: %s", err)
	}

	if err := d.Set("ok_actions", flattenStringSet(alarm.OKActions)); err != nil {
		return diag.Errorf("error setting ok_actions: %s", err)
	}

	tags, err := keyvaluetags.CloudwatchListTags(conn, aws.StringValue(alarm.AlarmArn))
	if err != nil {
		return diag.Errorf("error listing tags of alarm: %s", err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsCloudWatchCompositeAlarmUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchconn
	name := d.Id()

	input := expandAwsCloudWatchPutCompositeAlarmInput(d)

	_, err := conn.PutCompositeAlarmWithContext(ctx, &input)
	if err != nil {
		return diag.Errorf("error updating CloudWatch Composite Alarm (%s): %s", name, err)
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.CloudwatchUpdateTags(conn, arn, o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsCloudWatchCompositeAlarmRead(ctx, d, meta)
}

func resourceAwsCloudWatchCompositeAlarmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).cloudwatchconn
	name := d.Id()

	input := cloudwatch.DeleteAlarmsInput{
		AlarmNames: aws.StringSlice([]string{name}),
	}

	_, err := conn.DeleteAlarmsWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
			return nil
		}
		return diag.Errorf("error deleting CloudWatch Composite Alarm (%s): %s", name, err)
	}

	return nil
}

func expandAwsCloudWatchPutCompositeAlarmInput(d *schema.ResourceData) cloudwatch.PutCompositeAlarmInput {
	out := cloudwatch.PutCompositeAlarmInput{
		ActionsEnabled: aws.Bool(d.Get("actions_enabled").(bool)),
	}

	if v, ok := d.GetOk("alarm_actions"); ok {
		out.AlarmActions = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		out.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("alarm_name"); ok {
		out.AlarmName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("alarm_rule"); ok {
		out.AlarmRule = aws.String(v.(string))
	}

	if v, ok := d.GetOk("insufficient_data_actions"); ok {
		out.InsufficientDataActions = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("ok_actions"); ok {
		out.OKActions = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tags"); ok {
		out.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().CloudwatchTags()
	}

	return out
}
