package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsMediaStoreContainerMetricPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaStoreContainerMetricPolicyPut,
		Read:   resourceAwsMediaStoreContainerMetricPolicyRead,
		Update: resourceAwsMediaStoreContainerMetricPolicyPut,
		Delete: resourceAwsMediaStoreContainerMetricPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"metric_policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_level_metrics": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								mediastore.ContainerLevelMetricsEnabled,
								mediastore.ContainerLevelMetricsDisabled,
							}, false),
						},
						"metric_policy_rule": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 300,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"object_group": {
										Type:     schema.TypeString,
										Required: true,
									},
									"object_group_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceAwsMediaStoreContainerMetricPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	rawMetricPolicyList := d.Get("metric_policy").([]interface{})

	input := &mediastore.PutMetricPolicyInput{
		ContainerName: aws.String(d.Get("container_name").(string)),
		MetricPolicy:  expandMediaStoreMetricPolicy(rawMetricPolicyList[0].(map[string]interface{})),
	}

	_, err := conn.PutMetricPolicy(input)
	if err != nil {
		return fmt.Errorf("Error putting MediaStore Metric Policy: %s", err)
	}

	d.SetId(d.Get("container_name").(string))
	return resourceAwsMediaStoreContainerMetricPolicyRead(d, meta)
}

func resourceAwsMediaStoreContainerMetricPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.GetMetricPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	resp, err := conn.GetMetricPolicy(input)
	if err != nil {
		if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
			log.Printf("[WARN] MediaContainer %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, mediastore.ErrCodePolicyNotFoundException, "") {
			log.Printf("[WARN] MediaContainer Metric Policy for %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("container_name", d.Id())

	if err := d.Set("metric_policy", flattenMediaStoreMetricPolicy(resp.MetricPolicy)); err != nil {
		return fmt.Errorf("error setting metric_policy: %s", err)
	}

	return nil
}

func resourceAwsMediaStoreContainerMetricPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.DeleteMetricPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteMetricPolicy(input)
	if err != nil {
		if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
			return nil
		}
		if isAWSErr(err, mediastore.ErrCodePolicyNotFoundException, "") {
			return nil
		}
		// if isAWSErr(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
		// 	return nil
		// }
		return err
	}

	return nil
}

func expandMediaStoreMetricPolicy(rawMetricPolicy map[string]interface{}) *mediastore.MetricPolicy {
	return &mediastore.MetricPolicy{
		ContainerLevelMetrics: aws.String(rawMetricPolicy["container_level_metrics"].(string)),
		MetricPolicyRules:     expandMediaStoreMetricPolicyRules(rawMetricPolicy["metric_policy_rule"].([]interface{})),
	}
}

func flattenMediaStoreMetricPolicy(metricPolicy *mediastore.MetricPolicy) []map[string]interface{} {
	m := make(map[string]interface{})
	m["container_level_metrics"] = aws.StringValue(metricPolicy.ContainerLevelMetrics)
	m["metric_policy_rule"] = flattenMediaStoreMetricPolicyRules(metricPolicy.MetricPolicyRules)

	return []map[string]interface{}{m}
}

func expandMediaStoreMetricPolicyRules(rawMetricPolicyRules []interface{}) []*mediastore.MetricPolicyRule {
	metricPolicyRules := make([]*mediastore.MetricPolicyRule, 0, len(rawMetricPolicyRules))

	for _, m := range rawMetricPolicyRules {
		rawMetricPolicyRule := m.(map[string]interface{})
		metricPolicyRule := &mediastore.MetricPolicyRule{
			ObjectGroup:     aws.String(rawMetricPolicyRule["object_group"].(string)),
			ObjectGroupName: aws.String(rawMetricPolicyRule["object_group_name"].(string)),
		}

		metricPolicyRules = append(metricPolicyRules, metricPolicyRule)
	}

	return metricPolicyRules
}

func flattenMediaStoreMetricPolicyRules(metricPolicyRules []*mediastore.MetricPolicyRule) []interface{} {
	out := make([]interface{}, 0, len(metricPolicyRules))

	for _, metricPolicyRule := range metricPolicyRules {
		m := map[string]interface{}{
			"object_group":      aws.StringValue(metricPolicyRule.ObjectGroup),
			"object_group_name": aws.StringValue(metricPolicyRule.ObjectGroupName),
		}
		out = append(out, m)
	}

	return out
}
