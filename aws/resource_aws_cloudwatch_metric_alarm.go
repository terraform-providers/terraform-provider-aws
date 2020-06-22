package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCloudWatchMetricAlarm() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create:        resourceAwsCloudWatchMetricAlarmCreate,
		Read:          resourceAwsCloudWatchMetricAlarmRead,
		Update:        resourceAwsCloudWatchMetricAlarmUpdate,
		Delete:        resourceAwsCloudWatchMetricAlarmDelete,
		SchemaVersion: 1,
		MigrateState:  resourceAwsCloudWatchMetricAlarmMigrateState,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"alarm_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comparison_operator": {
				Type:     schema.TypeString,
				Required: true,
			},
			"evaluation_periods": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"metric_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"metric_query"},
			},
			"metric_query": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"metric_name"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"metric": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
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
										Optional: true,
									},
									"period": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"stat": {
										Type:     schema.TypeString,
										Required: true,
									},
									"unit": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"label": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"return_data": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"namespace": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"metric_query"},
			},
			"period": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"metric_query"},
			},
			"statistic": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"extended_statistic", "metric_query"},
			},
			"threshold": {
				Type:          schema.TypeFloat,
				Optional:      true,
				ConflictsWith: []string{"threshold_metric_id"},
			},
			"threshold_metric_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"threshold"},
				ValidateFunc:  validation.StringLenBetween(1, 255),
			},
			"actions_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"alarm_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.Any(
						validateArn,
						validateEC2AutomateARN,
					),
				},
				Set: schema.HashString,
			},
			"alarm_description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"datapoints_to_alarm": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"dimensions": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"metric_query"},
				Elem:          &schema.Schema{Type: schema.TypeString},
			},
			"insufficient_data_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"ok_actions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"unit": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"extended_statistic": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"statistic", "metric_query"},
			},
			"treat_missing_data": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "missing",
				ValidateFunc: validation.StringInSlice([]string{"breaching", "notBreaching", "ignore", "missing"}, true),
			},
			"evaluate_low_sample_count_percentiles": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"evaluate", "ignore"}, true),
			},

			"tags": tagsSchema(),
		},
	}
}

func validateResourceAwsCloudWatchMetricAlarm(d *schema.ResourceData) error {
	_, metricNameOk := d.GetOk("metric_name")
	_, statisticOk := d.GetOk("statistic")
	_, extendedStatisticOk := d.GetOk("extended_statistic")

	if metricNameOk && ((!statisticOk && !extendedStatisticOk) || (statisticOk && extendedStatisticOk)) {
		return fmt.Errorf("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm")
	}

	if v := d.Get("metric_query"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			metricQueryResource := v.(map[string]interface{})
			if v, ok := metricQueryResource["expression"]; ok && v.(string) != "" {
				if v := metricQueryResource["metric"]; v != nil {
					if len(v.([]interface{})) > 0 {
						return fmt.Errorf("No metric_query may have both `expression` and a `metric` specified")
					}
				}
			}
		}
	}

	return nil
}

func resourceAwsCloudWatchMetricAlarmCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn

	err := validateResourceAwsCloudWatchMetricAlarm(d)
	if err != nil {
		return err
	}
	params := getAwsCloudWatchPutMetricAlarmInput(d)

	log.Printf("[DEBUG] Creating CloudWatch Metric Alarm: %#v", params)
	_, err = conn.PutMetricAlarm(&params)
	if err != nil {
		return fmt.Errorf("Creating metric alarm failed: %s", err)
	}
	d.SetId(d.Get("alarm_name").(string))
	log.Println("[INFO] CloudWatch Metric Alarm created")

	return resourceAwsCloudWatchMetricAlarmRead(d, meta)
}

func resourceAwsCloudWatchMetricAlarmRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := getAwsCloudWatchMetricAlarm(d, meta)
	if err != nil {
		return err
	}
	if resp == nil {
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Reading CloudWatch Metric Alarm: %s", d.Get("alarm_name"))

	d.Set("actions_enabled", resp.ActionsEnabled)

	if err := d.Set("alarm_actions", _strArrPtrToList(resp.AlarmActions)); err != nil {
		log.Printf("[WARN] Error setting Alarm Actions: %s", err)
	}
	arn := *resp.AlarmArn
	d.Set("alarm_description", resp.AlarmDescription)
	d.Set("alarm_name", resp.AlarmName)
	d.Set("arn", arn)
	d.Set("comparison_operator", resp.ComparisonOperator)
	d.Set("datapoints_to_alarm", resp.DatapointsToAlarm)
	if err := d.Set("dimensions", flattenDimensions(resp.Dimensions)); err != nil {
		return err
	}
	d.Set("evaluation_periods", resp.EvaluationPeriods)

	if err := d.Set("insufficient_data_actions", _strArrPtrToList(resp.InsufficientDataActions)); err != nil {
		log.Printf("[WARN] Error setting Insufficient Data Actions: %s", err)
	}
	d.Set("metric_name", resp.MetricName)
	d.Set("namespace", resp.Namespace)

	if resp.Metrics != nil && len(resp.Metrics) > 0 {
		metricQueries := make([]interface{}, len(resp.Metrics))
		for i, mq := range resp.Metrics {
			metricQuery := map[string]interface{}{
				"expression":  aws.StringValue(mq.Expression),
				"id":          aws.StringValue(mq.Id),
				"label":       aws.StringValue(mq.Label),
				"return_data": aws.BoolValue(mq.ReturnData),
			}
			if mq.MetricStat != nil {
				metric := map[string]interface{}{
					"metric_name": aws.StringValue(mq.MetricStat.Metric.MetricName),
					"namespace":   aws.StringValue(mq.MetricStat.Metric.Namespace),
					"period":      int(aws.Int64Value(mq.MetricStat.Period)),
					"stat":        aws.StringValue(mq.MetricStat.Stat),
					"unit":        aws.StringValue(mq.MetricStat.Unit),
					"dimensions":  flattenDimensions(mq.MetricStat.Metric.Dimensions),
				}
				metricQuery["metric"] = []interface{}{metric}
			}
			metricQueries[i] = metricQuery
		}
		if err := d.Set("metric_query", metricQueries); err != nil {
			return fmt.Errorf("error setting metric_query: %s", err)
		}
	}

	if err := d.Set("ok_actions", _strArrPtrToList(resp.OKActions)); err != nil {
		log.Printf("[WARN] Error setting OK Actions: %s", err)
	}
	d.Set("period", resp.Period)
	d.Set("statistic", resp.Statistic)
	d.Set("threshold", resp.Threshold)
	d.Set("threshold_metric_id", resp.ThresholdMetricId)
	d.Set("unit", resp.Unit)
	d.Set("extended_statistic", resp.ExtendedStatistic)
	d.Set("treat_missing_data", resp.TreatMissingData)
	d.Set("evaluate_low_sample_count_percentiles", resp.EvaluateLowSampleCountPercentile)

	tags, err := keyvaluetags.CloudwatchListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CloudWatch Metric Alarm (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsCloudWatchMetricAlarmUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn
	params := getAwsCloudWatchPutMetricAlarmInput(d)

	log.Printf("[DEBUG] Updating CloudWatch Metric Alarm: %#v", params)
	_, err := conn.PutMetricAlarm(&params)
	if err != nil {
		return fmt.Errorf("Updating metric alarm failed: %s", err)
	}
	log.Println("[INFO] CloudWatch Metric Alarm updated")

	arn := d.Get("arn").(string)
	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.CloudwatchUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating CloudWatch Metric Alarm (%s) tags: %s", arn, err)
		}
	}

	return resourceAwsCloudWatchMetricAlarmRead(d, meta)
}

func resourceAwsCloudWatchMetricAlarmDelete(d *schema.ResourceData, meta interface{}) error {
	resp, err := getAwsCloudWatchMetricAlarm(d, meta)
	if err != nil {
		return err
	}
	if resp == nil {
		log.Printf("[DEBUG] CloudWatch Metric Alarm %s is already gone", d.Id())
		return nil
	}

	log.Printf("[INFO] Deleting CloudWatch Metric Alarm: %s", d.Id())

	conn := meta.(*AWSClient).cloudwatchconn
	params := cloudwatch.DeleteAlarmsInput{
		AlarmNames: []*string{aws.String(d.Id())},
	}

	if _, err := conn.DeleteAlarms(&params); err != nil {
		return fmt.Errorf("Error deleting CloudWatch Metric Alarm: %s", err)
	}
	log.Println("[INFO] CloudWatch Metric Alarm deleted")

	return nil
}

func getAwsCloudWatchPutMetricAlarmInput(d *schema.ResourceData) cloudwatch.PutMetricAlarmInput {
	params := cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(d.Get("alarm_name").(string)),
		ComparisonOperator: aws.String(d.Get("comparison_operator").(string)),
		EvaluationPeriods:  aws.Int64(int64(d.Get("evaluation_periods").(int))),
		TreatMissingData:   aws.String(d.Get("treat_missing_data").(string)),
		Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().CloudwatchTags(),
	}

	if v := d.Get("actions_enabled"); v != nil {
		params.ActionsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		params.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("datapoints_to_alarm"); ok {
		params.DatapointsToAlarm = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("metric_name"); ok {
		params.MetricName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("namespace"); ok {
		params.Namespace = aws.String(v.(string))
	}
	if v, ok := d.GetOk("period"); ok {
		params.Period = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("unit"); ok {
		params.Unit = aws.String(v.(string))
	}

	if v, ok := d.GetOk("statistic"); ok {
		params.Statistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk("extended_statistic"); ok {
		params.ExtendedStatistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk("evaluate_low_sample_count_percentiles"); ok {
		params.EvaluateLowSampleCountPercentile = aws.String(v.(string))
	}

	if v, ok := d.GetOk("threshold_metric_id"); ok {
		params.ThresholdMetricId = aws.String(v.(string))
	} else {
		params.Threshold = aws.Float64(d.Get("threshold").(float64))
	}

	var alarmActions []*string
	if v := d.Get("alarm_actions"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			str := v.(string)
			alarmActions = append(alarmActions, aws.String(str))
		}
		params.AlarmActions = alarmActions
	}

	var insufficientDataActions []*string
	if v := d.Get("insufficient_data_actions"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			str := v.(string)
			insufficientDataActions = append(insufficientDataActions, aws.String(str))
		}
		params.InsufficientDataActions = insufficientDataActions
	}

	var metrics []*cloudwatch.MetricDataQuery
	if v := d.Get("metric_query"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			metricQueryResource := v.(map[string]interface{})
			id := metricQueryResource["id"].(string)
			if id == "" {
				continue
			}
			metricQuery := cloudwatch.MetricDataQuery{
				Id: aws.String(id),
			}
			if v, ok := metricQueryResource["expression"]; ok && v.(string) != "" {
				metricQuery.Expression = aws.String(v.(string))
			}
			if v, ok := metricQueryResource["label"]; ok && v.(string) != "" {
				metricQuery.Label = aws.String(v.(string))
			}
			if v, ok := metricQueryResource["return_data"]; ok {
				metricQuery.ReturnData = aws.Bool(v.(bool))
			}
			if v := metricQueryResource["metric"]; v != nil {
				for _, v := range v.([]interface{}) {
					metricResource := v.(map[string]interface{})
					metric := cloudwatch.Metric{
						MetricName: aws.String(metricResource["metric_name"].(string)),
					}
					metricStat := cloudwatch.MetricStat{
						Metric: &metric,
						Stat:   aws.String(metricResource["stat"].(string)),
					}
					if v, ok := metricResource["namespace"]; ok && v.(string) != "" {
						metric.Namespace = aws.String(v.(string))
					}
					if v, ok := metricResource["period"]; ok {
						metricStat.Period = aws.Int64(int64(v.(int)))
					}
					if v, ok := metricResource["unit"]; ok && v.(string) != "" {
						metricStat.Unit = aws.String(v.(string))
					}
					a := metricResource["dimensions"].(map[string]interface{})
					dimensions := make([]*cloudwatch.Dimension, 0, len(a))
					for k, v := range a {
						dimensions = append(dimensions, &cloudwatch.Dimension{
							Name:  aws.String(k),
							Value: aws.String(v.(string)),
						})
					}
					metric.Dimensions = dimensions
					metricQuery.MetricStat = &metricStat
				}
			}
			metrics = append(metrics, &metricQuery)
		}
		params.Metrics = metrics
	}

	var okActions []*string
	if v := d.Get("ok_actions"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			str := v.(string)
			okActions = append(okActions, aws.String(str))
		}
		params.OKActions = okActions
	}

	a := d.Get("dimensions").(map[string]interface{})
	var dimensions []*cloudwatch.Dimension
	for k, v := range a {
		dimensions = append(dimensions, &cloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	params.Dimensions = dimensions

	return params
}

func getAwsCloudWatchMetricAlarm(d *schema.ResourceData, meta interface{}) (*cloudwatch.MetricAlarm, error) {
	conn := meta.(*AWSClient).cloudwatchconn

	params := cloudwatch.DescribeAlarmsInput{
		AlarmNames: []*string{aws.String(d.Id())},
	}

	resp, err := conn.DescribeAlarms(&params)
	if err != nil {
		return nil, err
	}

	// Find it and return it
	for idx, ma := range resp.MetricAlarms {
		if aws.StringValue(ma.AlarmName) == d.Id() {
			return resp.MetricAlarms[idx], nil
		}
	}

	return nil, nil
}

func _strArrPtrToList(strArrPtr []*string) []string {
	var result []string
	for _, elem := range strArrPtr {
		result = append(result, aws.StringValue(elem))
	}
	return result
}

func flattenDimensions(dims []*cloudwatch.Dimension) map[string]interface{} {
	flatDims := make(map[string]interface{})
	for _, d := range dims {
		flatDims[*d.Name] = *d.Value
	}
	return flatDims
}
