package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsCloudWatchMetricAlarm() *schema.Resource {
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
				Type:     schema.TypeFloat,
				Required: true,
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
					ValidateFunc: validateAny(
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
	a, err := getAwsCloudWatchMetricAlarm(d, meta)
	if err != nil {
		return err
	}
	if a == nil {
		d.SetId("")
		return nil
	}

	log.Printf("[DEBUG] Reading CloudWatch Metric Alarm: %s", d.Get("alarm_name"))

	d.Set("actions_enabled", a.ActionsEnabled)

	if err := d.Set("alarm_actions", _strArrPtrToList(a.AlarmActions)); err != nil {
		log.Printf("[WARN] Error setting Alarm Actions: %s", err)
	}
	d.Set("alarm_description", a.AlarmDescription)
	d.Set("alarm_name", a.AlarmName)
	d.Set("arn", a.AlarmArn)
	d.Set("comparison_operator", a.ComparisonOperator)
	d.Set("datapoints_to_alarm", a.DatapointsToAlarm)
	if err := d.Set("dimensions", flattenDimensions(a.Dimensions)); err != nil {
		return err
	}
	d.Set("evaluation_periods", a.EvaluationPeriods)

	if err := d.Set("insufficient_data_actions", _strArrPtrToList(a.InsufficientDataActions)); err != nil {
		log.Printf("[WARN] Error setting Insufficient Data Actions: %s", err)
	}
	d.Set("metric_name", a.MetricName)
	d.Set("namespace", a.Namespace)

	if a.Metrics != nil && len(a.Metrics) > 0 {
		metricQueries := make([]interface{}, len(a.Metrics))
		for i, mq := range a.Metrics {
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

	if err := d.Set("ok_actions", _strArrPtrToList(a.OKActions)); err != nil {
		log.Printf("[WARN] Error setting OK Actions: %s", err)
	}
	d.Set("period", a.Period)
	d.Set("statistic", a.Statistic)
	d.Set("threshold", a.Threshold)
	d.Set("unit", a.Unit)
	d.Set("extended_statistic", a.ExtendedStatistic)
	d.Set("treat_missing_data", a.TreatMissingData)
	d.Set("evaluate_low_sample_count_percentiles", a.EvaluateLowSampleCountPercentile)

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

	return resourceAwsCloudWatchMetricAlarmRead(d, meta)
}

func resourceAwsCloudWatchMetricAlarmDelete(d *schema.ResourceData, meta interface{}) error {
	p, err := getAwsCloudWatchMetricAlarm(d, meta)
	if err != nil {
		return err
	}
	if p == nil {
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
		Threshold:          aws.Float64(d.Get("threshold").(float64)),
		TreatMissingData:   aws.String(d.Get("treat_missing_data").(string)),
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
			metricQuery := cloudwatch.MetricDataQuery{
				Id: aws.String(metricQueryResource["id"].(string)),
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
