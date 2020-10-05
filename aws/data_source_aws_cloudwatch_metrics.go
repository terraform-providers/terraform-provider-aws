package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
)

func dataSourceAwsCloudwatchMetrics() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudwatchMetricsRead,

		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metric_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dimensions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
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
			"metrics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:     schema.TypeString,
							Required: true,
						},
						"dimensions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
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
						"metric_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsCloudwatchMetricsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatchconn
	params := &cloudwatch.ListMetricsInput{}

	if v, ok := d.GetOk("namespace"); ok {
		params.Namespace = aws.String(v.(string))
	}
	if v, ok := d.GetOk("metric_name"); ok {
		params.MetricName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("dimensions"); ok {
		params.Dimensions = dataSourceAwsCloudwatchMetricsDimensionFilter(v.(*schema.Set))
	}

	metrics, err := dataSourceAwsCloudwatchMetricsLookup(conn, params)
	if err != nil {
		return err
	}
	if len(metrics) == 0 {
		return fmt.Errorf("no metrics found with the current filters: %+v", params)
	}

	d.SetId(fmt.Sprintf("%d", hashcode.String(params.String())))

	if err := d.Set("metrics", dataSourceAwsCloudwatchMetricsConvertMetricToTypeList(metrics)); err != nil {
		return fmt.Errorf("Error settings metrics: %s", err)
	}

	return nil
}

func dataSourceAwsCloudwatchMetricsConvertMetricToTypeList(input []*cloudwatch.Metric) []map[string]interface{} {
	metricsTypeList := make([]map[string]interface{}, len(input))
	for k, metric := range input {
		dimensionList := []map[string]string{}
		for _, dimension := range metric.Dimensions {
			dimensionList = append(dimensionList, map[string]string{
				"name":  aws.StringValue(dimension.Name),
				"value": aws.StringValue(dimension.Value),
			})
		}
		metricsTypeList[k] = map[string]interface{}{
			"namespace":   aws.StringValue(metric.Namespace),
			"dimensions":  dimensionList,
			"metric_name": aws.StringValue(metric.MetricName),
		}
	}

	return metricsTypeList
}

func dataSourceAwsCloudwatchMetricsLookup(conn *cloudwatch.CloudWatch, params *cloudwatch.ListMetricsInput) ([]*cloudwatch.Metric, error) {
	var metrics []*cloudwatch.Metric
	err := conn.ListMetricsPages(params, func(page *cloudwatch.ListMetricsOutput, lastPage bool) bool {
		metrics = append(metrics, page.Metrics...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func dataSourceAwsCloudwatchMetricsDimensionFilter(set *schema.Set) []*cloudwatch.DimensionFilter {
	var dimensionFilters []*cloudwatch.DimensionFilter
	for _, v := range set.List() {
		m := v.(map[string]interface{})
		dimensionFilters = append(dimensionFilters, &cloudwatch.DimensionFilter{
			Name:  aws.String(m["name"].(string)),
			Value: aws.String(m["value"].(string)),
		})
	}
	return dimensionFilters
}
