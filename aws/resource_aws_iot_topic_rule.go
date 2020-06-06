package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsIotTopicRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotTopicRuleCreate,
		Read:   resourceAwsIotTopicRuleRead,
		Update: resourceAwsIotTopicRuleUpdate,
		Delete: resourceAwsIotTopicRuleDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_alarm": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarm_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"state_reason": {
							Type:     schema.TypeString,
							Required: true,
						},
						"state_value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateIoTTopicRuleCloudWatchAlarmStateValue,
						},
					},
				},
			},
			"cloudwatch_metric": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"metric_namespace": {
							Type:     schema.TypeString,
							Required: true,
						},
						"metric_timestamp": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateIoTTopicRuleCloudWatchMetricTimestamp,
						},
						"metric_unit": {
							Type:     schema.TypeString,
							Required: true,
						},
						"metric_value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dynamodb": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hash_key_field": {
							Type:     schema.TypeString,
							Required: true,
						},
						"hash_key_value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"hash_key_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"operation": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"DELETE",
								"INSERT",
								"UPDATE",
							}, false),
						},
						"payload_field": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"range_key_field": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"range_key_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"range_key_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"dynamodbv2": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"put_item": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"table_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"elasticsearch": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateIoTTopicRuleElasticSearchEndpoint,
						},
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"index": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"firehose": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delivery_stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"separator": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateIoTTopicRuleFirehoseSeparator,
						},
					},
				},
			},
			"iot_analytics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"iot_events": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"message_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"kinesis": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"partition_key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"lambda": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateIoTTopicRuleName,
			},
			"republish": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"qos": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 1),
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"topic": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"s3": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"step_functions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"execution_name_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"state_machine_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"sns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message_format": {
							Type:     schema.TypeString,
							Default:  iot.MessageFormatRaw,
							Optional: true,
						},
						"target_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"sql": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sql_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sqs": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
						"use_base64": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsIotTopicRuleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	ruleName := d.Get("name").(string)

	input := &iot.CreateTopicRuleInput{
		RuleName:         aws.String(ruleName),
		Tags:             aws.String(keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().UrlEncode()),
		TopicRulePayload: expandIotTopicRulePayload(d),
	}

	_, err := conn.CreateTopicRule(input)

	if err != nil {
		return fmt.Errorf("error creating IoT Topic Rule (%s): %w", ruleName, err)
	}

	d.SetId(ruleName)

	return resourceAwsIotTopicRuleRead(d, meta)
}

func resourceAwsIotTopicRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &iot.GetTopicRuleInput{
		RuleName: aws.String(d.Id()),
	}

	out, err := conn.GetTopicRule(input)

	if err != nil {
		return fmt.Errorf("error getting IoT Topic Rule (%s): %w", d.Id(), err)
	}

	d.Set("arn", out.RuleArn)
	d.Set("name", out.Rule.RuleName)
	d.Set("description", out.Rule.Description)
	d.Set("enabled", !aws.BoolValue(out.Rule.RuleDisabled))
	d.Set("sql", out.Rule.Sql)
	d.Set("sql_version", out.Rule.AwsIotSqlVersion)

	tags, err := keyvaluetags.IotListTags(conn, aws.StringValue(out.RuleArn))

	if err != nil {
		return fmt.Errorf("error listing tags for IoT Topic Rule (%s): %w", aws.StringValue(out.RuleArn), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("cloudwatch_alarm", flattenIotCloudWatchAlarmActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting cloudwatch_alarm: %w", err)
	}

	if err := d.Set("cloudwatch_metric", flattenIotCloudwatchMetricActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting cloudwatch_metric: %w", err)
	}

	if err := d.Set("dynamodb", flattenIotDynamoDbActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting dynamodb: %w", err)
	}

	if err := d.Set("dynamodbv2", flattenIotDynamoDbv2Actions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting dynamodbv2: %w", err)
	}

	if err := d.Set("elasticsearch", flattenIotElasticsearchActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting elasticsearch: %w", err)
	}

	if err := d.Set("firehose", flattenIotFirehoseActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting firehose: %w", err)
	}

	if err := d.Set("iot_analytics", flattenIotIotAnalyticsActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting iot_analytics: %w", err)
	}

	if err := d.Set("iot_events", flattenIotIotEventsActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting iot_events: %w", err)
	}

	if err := d.Set("kinesis", flattenIotKinesisActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting kinesis: %w", err)
	}

	if err := d.Set("lambda", flattenIotLambdaActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting lambda: %w", err)
	}

	if err := d.Set("republish", flattenIotRepublishActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting republish: %w", err)
	}

	if err := d.Set("s3", flattenIotS3Actions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting s3: %w", err)
	}

	if err := d.Set("sns", flattenIotSnsActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting sns: %w", err)
	}

	if err := d.Set("sqs", flattenIotSqsActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting sqs: %w", err)
	}

	if err := d.Set("step_functions", flattenIotStepFunctionsActions(out.Rule.Actions)); err != nil {
		return fmt.Errorf("error setting step_functions: %w", err)
	}

	return nil
}

func resourceAwsIotTopicRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	if d.HasChanges(
		"cloudwatch_alarm",
		"cloudwatch_metric",
		"description",
		"dynamodb",
		"dynamodbv2",
		"elasticsearch",
		"enabled",
		"firehose",
		"iot_analytics",
		"iot_events",
		"kinesis",
		"lambda",
		"republish",
		"s3",
		"step_functions",
		"sns",
		"sql",
		"sql_version",
		"sqs",
	) {
		input := &iot.ReplaceTopicRuleInput{
			RuleName:         aws.String(d.Get("name").(string)),
			TopicRulePayload: expandIotTopicRulePayload(d),
		}

		_, err := conn.ReplaceTopicRule(input)

		if err != nil {
			return fmt.Errorf("error updating IoT Topic Rule (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.IotUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsIotTopicRuleRead(d, meta)
}

func resourceAwsIotTopicRuleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	input := &iot.DeleteTopicRuleInput{
		RuleName: aws.String(d.Id()),
	}

	_, err := conn.DeleteTopicRule(input)

	if err != nil {
		return fmt.Errorf("error deleting IoT Topic Rule (%s): %w", d.Id(), err)
	}

	return nil
}

func expandIotPutItemInput(tfList []interface{}) *iot.PutItemInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.PutItemInput{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["table_name"].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	return apiObject
}

func expandIotCloudwatchAlarmAction(tfList []interface{}) *iot.CloudwatchAlarmAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.CloudwatchAlarmAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["alarm_name"].(string); ok && v != "" {
		apiObject.AlarmName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["state_reason"].(string); ok && v != "" {
		apiObject.StateReason = aws.String(v)
	}

	if v, ok := tfMap["state_value"].(string); ok && v != "" {
		apiObject.StateValue = aws.String(v)
	}

	return apiObject
}

func expandIotCloudwatchMetricAction(tfList []interface{}) *iot.CloudwatchMetricAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.CloudwatchMetricAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["metric_name"].(string); ok && v != "" {
		apiObject.MetricName = aws.String(v)
	}

	if v, ok := tfMap["metric_namespace"].(string); ok && v != "" {
		apiObject.MetricNamespace = aws.String(v)
	}

	if v, ok := tfMap["metric_timestamp"].(string); ok && v != "" {
		apiObject.MetricTimestamp = aws.String(v)
	}

	if v, ok := tfMap["metric_unit"].(string); ok && v != "" {
		apiObject.MetricUnit = aws.String(v)
	}

	if v, ok := tfMap["metric_value"].(string); ok && v != "" {
		apiObject.MetricValue = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandIotDynamoDBAction(tfList []interface{}) *iot.DynamoDBAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.DynamoDBAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["hash_key_field"].(string); ok && v != "" {
		apiObject.HashKeyField = aws.String(v)
	}

	if v, ok := tfMap["hash_key_type"].(string); ok && v != "" {
		apiObject.HashKeyType = aws.String(v)
	}

	if v, ok := tfMap["hash_key_value"].(string); ok && v != "" {
		apiObject.HashKeyValue = aws.String(v)
	}

	if v, ok := tfMap["operation"].(string); ok && v != "" {
		apiObject.Operation = aws.String(v)
	}

	if v, ok := tfMap["payload_field"].(string); ok && v != "" {
		apiObject.PayloadField = aws.String(v)
	}

	if v, ok := tfMap["range_key_field"].(string); ok && v != "" {
		apiObject.RangeKeyField = aws.String(v)
	}

	if v, ok := tfMap["range_key_type"].(string); ok && v != "" {
		apiObject.RangeKeyType = aws.String(v)
	}

	if v, ok := tfMap["range_key_value"].(string); ok && v != "" {
		apiObject.RangeKeyValue = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["table_name"].(string); ok && v != "" {
		apiObject.TableName = aws.String(v)
	}

	return apiObject
}

func expandIotDynamoDBv2Action(tfList []interface{}) *iot.DynamoDBv2Action {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.DynamoDBv2Action{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["put_item"].([]interface{}); ok {
		apiObject.PutItem = expandIotPutItemInput(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandIotElasticsearchAction(tfList []interface{}) *iot.ElasticsearchAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.ElasticsearchAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["endpoint"].(string); ok && v != "" {
		apiObject.Endpoint = aws.String(v)
	}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.Id = aws.String(v)
	}

	if v, ok := tfMap["index"].(string); ok && v != "" {
		apiObject.Index = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandIotFirehoseAction(tfList []interface{}) *iot.FirehoseAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.FirehoseAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["delivery_stream_name"].(string); ok && v != "" {
		apiObject.DeliveryStreamName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["separator"].(string); ok && v != "" {
		apiObject.Separator = aws.String(v)
	}

	return apiObject
}

func expandIotIotAnalyticsAction(tfList []interface{}) *iot.IotAnalyticsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.IotAnalyticsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["channel_name"].(string); ok && v != "" {
		apiObject.ChannelName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandIotIotEventsAction(tfList []interface{}) *iot.IotEventsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.IotEventsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["input_name"].(string); ok && v != "" {
		apiObject.InputName = aws.String(v)
	}

	if v, ok := tfMap["message_id"].(string); ok && v != "" {
		apiObject.MessageId = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandIotKinesisAction(tfList []interface{}) *iot.KinesisAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.KinesisAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["partition_key"].(string); ok && v != "" {
		apiObject.PartitionKey = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["stream_name"].(string); ok && v != "" {
		apiObject.StreamName = aws.String(v)
	}

	return apiObject
}

func expandIotLambdaAction(tfList []interface{}) *iot.LambdaAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.LambdaAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["function_arn"].(string); ok && v != "" {
		apiObject.FunctionArn = aws.String(v)
	}

	return apiObject
}

func expandIotRepublishAction(tfList []interface{}) *iot.RepublishAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.RepublishAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["qos"].(int); ok {
		apiObject.Qos = aws.Int64(int64(v))
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["topic"].(string); ok && v != "" {
		apiObject.Topic = aws.String(v)
	}

	return apiObject
}

func expandIotS3Action(tfList []interface{}) *iot.S3Action {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.S3Action{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandIotSnsAction(tfList []interface{}) *iot.SnsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.SnsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["message_format"].(string); ok && v != "" {
		apiObject.MessageFormat = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["target_arn"].(string); ok && v != "" {
		apiObject.TargetArn = aws.String(v)
	}

	return apiObject
}

func expandIotSqsAction(tfList []interface{}) *iot.SqsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.SqsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["queue_url"].(string); ok && v != "" {
		apiObject.QueueUrl = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["use_base64"].(bool); ok {
		apiObject.UseBase64 = aws.Bool(v)
	}

	return apiObject
}

func expandIotStepFunctionsAction(tfList []interface{}) *iot.StepFunctionsAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &iot.StepFunctionsAction{}
	tfMap := tfList[0].(map[string]interface{})

	if v, ok := tfMap["execution_name_prefix"].(string); ok && v != "" {
		apiObject.ExecutionNamePrefix = aws.String(v)
	}

	if v, ok := tfMap["state_machine_name"].(string); ok && v != "" {
		apiObject.StateMachineName = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleArn = aws.String(v)
	}

	return apiObject
}

func expandIotTopicRulePayload(d *schema.ResourceData) *iot.TopicRulePayload {
	var actions []*iot.Action

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("cloudwatch_alarm").(*schema.Set).List() {
		action := expandIotCloudwatchAlarmAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{CloudwatchAlarm: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("cloudwatch_metric").(*schema.Set).List() {
		action := expandIotCloudwatchMetricAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{CloudwatchMetric: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("dynamodb").(*schema.Set).List() {
		action := expandIotDynamoDBAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{DynamoDB: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("dynamodbv2").(*schema.Set).List() {
		action := expandIotDynamoDBv2Action([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{DynamoDBv2: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("elasticsearch").(*schema.Set).List() {
		action := expandIotElasticsearchAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Elasticsearch: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("firehose").(*schema.Set).List() {
		action := expandIotFirehoseAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Firehose: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("iot_analytics").(*schema.Set).List() {
		action := expandIotIotAnalyticsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{IotAnalytics: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("iot_events").(*schema.Set).List() {
		action := expandIotIotEventsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{IotEvents: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("kinesis").(*schema.Set).List() {
		action := expandIotKinesisAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Kinesis: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("lambda").(*schema.Set).List() {
		action := expandIotLambdaAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Lambda: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("republish").(*schema.Set).List() {
		action := expandIotRepublishAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Republish: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("s3").(*schema.Set).List() {
		action := expandIotS3Action([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{S3: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("sns").(*schema.Set).List() {
		action := expandIotSnsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Sns: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("sqs").(*schema.Set).List() {
		action := expandIotSqsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{Sqs: action})
	}

	// Legacy root attribute handling
	for _, tfMapRaw := range d.Get("step_functions").(*schema.Set).List() {
		action := expandIotStepFunctionsAction([]interface{}{tfMapRaw})

		if action == nil {
			continue
		}

		actions = append(actions, &iot.Action{StepFunctions: action})
	}

	// Prevent sending empty Actions:
	// - missing required field, CreateTopicRuleInput.TopicRulePayload.Actions
	if len(actions) == 0 {
		actions = []*iot.Action{}
	}

	return &iot.TopicRulePayload{
		Actions:          actions,
		AwsIotSqlVersion: aws.String(d.Get("sql_version").(string)),
		Description:      aws.String(d.Get("description").(string)),
		RuleDisabled:     aws.Bool(!d.Get("enabled").(bool)),
		Sql:              aws.String(d.Get("sql").(string)),
	}
}

func flattenIotCloudwatchAlarmAction(apiObject *iot.CloudwatchAlarmAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.AlarmName; v != nil {
		tfMap["alarm_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.StateReason; v != nil {
		tfMap["state_reason"] = aws.StringValue(v)
	}

	if v := apiObject.StateValue; v != nil {
		tfMap["state_value"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotCloudWatchAlarmActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.CloudwatchAlarm; v != nil {
			results = append(results, flattenIotCloudwatchAlarmAction(v)...)
		}
	}

	return results
}

// Legacy root attribute handling
func flattenIotCloudwatchMetricActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.CloudwatchMetric; v != nil {
			results = append(results, flattenIotCloudwatchMetricAction(v)...)
		}
	}

	return results
}

func flattenIotCloudwatchMetricAction(apiObject *iot.CloudwatchMetricAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.MetricName; v != nil {
		tfMap["metric_name"] = aws.StringValue(v)
	}

	if v := apiObject.MetricNamespace; v != nil {
		tfMap["metric_namespace"] = aws.StringValue(v)
	}

	if v := apiObject.MetricTimestamp; v != nil {
		tfMap["metric_timestamp"] = aws.StringValue(v)
	}

	if v := apiObject.MetricUnit; v != nil {
		tfMap["metric_unit"] = aws.StringValue(v)
	}

	if v := apiObject.MetricValue; v != nil {
		tfMap["metric_value"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotDynamoDbActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.DynamoDB; v != nil {
			results = append(results, flattenIotDynamoDBAction(v)...)
		}
	}

	return results
}

func flattenIotDynamoDBAction(apiObject *iot.DynamoDBAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.HashKeyField; v != nil {
		tfMap["hash_key_field"] = aws.StringValue(v)
	}

	if v := apiObject.HashKeyType; v != nil {
		tfMap["hash_key_type"] = aws.StringValue(v)
	}

	if v := apiObject.HashKeyValue; v != nil {
		tfMap["hash_key_value"] = aws.StringValue(v)
	}

	if v := apiObject.PayloadField; v != nil {
		tfMap["payload_field"] = aws.StringValue(v)
	}

	if v := apiObject.Operation; v != nil {
		tfMap["operation"] = aws.StringValue(v)
	}

	if v := apiObject.RangeKeyField; v != nil {
		tfMap["range_key_field"] = aws.StringValue(v)
	}

	if v := apiObject.RangeKeyType; v != nil {
		tfMap["range_key_type"] = aws.StringValue(v)
	}

	if v := apiObject.RangeKeyValue; v != nil {
		tfMap["range_key_value"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.TableName; v != nil {
		tfMap["table_name"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotDynamoDbv2Actions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.DynamoDBv2; v != nil {
			results = append(results, flattenIotDynamoDBv2Action(v)...)
		}
	}

	return results
}

func flattenIotDynamoDBv2Action(apiObject *iot.DynamoDBv2Action) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.PutItem; v != nil {
		tfMap["put_item"] = flattenIotPutItemInput(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotElasticsearchActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Elasticsearch; v != nil {
			results = append(results, flattenIotElasticsearchAction(v)...)
		}
	}

	return results
}

func flattenIotElasticsearchAction(apiObject *iot.ElasticsearchAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Endpoint; v != nil {
		tfMap["endpoint"] = aws.StringValue(v)
	}

	if v := apiObject.Id; v != nil {
		tfMap["id"] = aws.StringValue(v)
	}

	if v := apiObject.Index; v != nil {
		tfMap["index"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotFirehoseActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Firehose; v != nil {
			results = append(results, flattenIotFirehoseAction(v)...)
		}
	}

	return results
}

func flattenIotFirehoseAction(apiObject *iot.FirehoseAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.DeliveryStreamName; v != nil {
		tfMap["delivery_stream_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Separator; v != nil {
		tfMap["separator"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotIotAnalyticsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.IotAnalytics; v != nil {
			results = append(results, flattenIotIotAnalyticsAction(v)...)
		}
	}

	return results
}

func flattenIotIotAnalyticsAction(apiObject *iot.IotAnalyticsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.ChannelName; v != nil {
		tfMap["channel_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotIotEventsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.IotEvents; v != nil {
			results = append(results, flattenIotIotEventsAction(v)...)
		}
	}

	return results
}

func flattenIotIotEventsAction(apiObject *iot.IotEventsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.InputName; v != nil {
		tfMap["input_name"] = aws.StringValue(v)
	}

	if v := apiObject.MessageId; v != nil {
		tfMap["message_id"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotKinesisActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Kinesis; v != nil {
			results = append(results, flattenIotKinesisAction(v)...)
		}
	}

	return results
}

func flattenIotKinesisAction(apiObject *iot.KinesisAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.PartitionKey; v != nil {
		tfMap["partition_key"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.StreamName; v != nil {
		tfMap["stream_name"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotLambdaActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Lambda; v != nil {
			results = append(results, flattenIotLambdaAction(v)...)
		}
	}

	return results
}

func flattenIotLambdaAction(apiObject *iot.LambdaAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.FunctionArn; v != nil {
		tfMap["function_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func flattenIotPutItemInput(apiObject *iot.PutItemInput) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.TableName; v != nil {
		tfMap["table_name"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotRepublishActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Republish; v != nil {
			results = append(results, flattenIotRepublishAction(v)...)
		}
	}

	return results
}

func flattenIotRepublishAction(apiObject *iot.RepublishAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.Qos; v != nil {
		tfMap["qos"] = aws.Int64Value(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.Topic; v != nil {
		tfMap["topic"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotS3Actions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.S3; v != nil {
			results = append(results, flattenIotS3Action(v)...)
		}
	}

	return results
}

func flattenIotS3Action(apiObject *iot.S3Action) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.BucketName; v != nil {
		tfMap["bucket_name"] = aws.StringValue(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap["key"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotSnsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Sns; v != nil {
			results = append(results, flattenIotSnsAction(v)...)
		}
	}

	return results
}

func flattenIotSnsAction(apiObject *iot.SnsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.MessageFormat; v != nil {
		tfMap["message_format"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.TargetArn; v != nil {
		tfMap["target_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotSqsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.Sqs; v != nil {
			results = append(results, flattenIotSqsAction(v)...)
		}
	}

	return results
}

func flattenIotSqsAction(apiObject *iot.SqsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.QueueUrl; v != nil {
		tfMap["queue_url"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.UseBase64; v != nil {
		tfMap["use_base64"] = aws.BoolValue(v)
	}

	return []interface{}{tfMap}
}

// Legacy root attribute handling
func flattenIotStepFunctionsActions(actions []*iot.Action) []interface{} {
	results := make([]interface{}, 0)

	for _, action := range actions {
		if action == nil {
			continue
		}

		if v := action.StepFunctions; v != nil {
			results = append(results, flattenIotStepFunctionsAction(v)...)
		}
	}

	return results
}

func flattenIotStepFunctionsAction(apiObject *iot.StepFunctionsAction) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})

	if v := apiObject.ExecutionNamePrefix; v != nil {
		tfMap["execution_name_prefix"] = aws.StringValue(v)
	}

	if v := apiObject.StateMachineName; v != nil {
		tfMap["state_machine_name"] = aws.StringValue(v)
	}

	if v := apiObject.RoleArn; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
