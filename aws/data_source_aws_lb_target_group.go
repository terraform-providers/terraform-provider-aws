package aws

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLbTargetGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLbTargetGroupRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deregistration_delay": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"health_check": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"healthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"matcher": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"unhealthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"lambda_multi_value_headers_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"load_balancing_algorithm_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"preserve_client_ip": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"protocol_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"proxy_protocol_v2": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"slow_start": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"stickiness": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookie_duration": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"target_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLbTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).elbv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &elbv2.DescribeTargetGroupsInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.TargetGroupArns = aws.StringSlice([]string{v.(string)})
	} else if v, ok := d.GetOk("name"); ok {
		input.Names = aws.StringSlice([]string{v.(string)})
	}

	var results []*elbv2.TargetGroup

	err := conn.DescribeTargetGroupsPages(input, func(page *elbv2.DescribeTargetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		results = append(results, page.TargetGroups...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error retrieving LB Target Group: %w", err)
	}
	if len(results) != 1 {
		return fmt.Errorf("Search returned %d results, please revise so only one is returned", len(results))
	}

	targetGroup := results[0]

	d.SetId(aws.StringValue(targetGroup.TargetGroupArn))

	d.Set("arn", targetGroup.TargetGroupArn)
	d.Set("arn_suffix", lbTargetGroupSuffixFromARN(targetGroup.TargetGroupArn))
	d.Set("name", targetGroup.TargetGroupName)
	d.Set("target_type", targetGroup.TargetType)

	if err := d.Set("health_check", flattenLbTargetGroupHealthCheck(targetGroup)); err != nil {
		return fmt.Errorf("error setting health_check: %w", err)
	}

	if v, _ := d.Get("target_type").(string); v != elbv2.TargetTypeEnumLambda {
		d.Set("vpc_id", targetGroup.VpcId)
		d.Set("port", targetGroup.Port)
		d.Set("protocol", targetGroup.Protocol)
	}
	switch d.Get("protocol").(string) {
	case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
		d.Set("protocol_version", targetGroup.ProtocolVersion)
	}

	attrResp, err := conn.DescribeTargetGroupAttributes(&elbv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("error retrieving Target Group Attributes: %w", err)
	}

	for _, attr := range attrResp.Attributes {
		switch aws.StringValue(attr.Key) {
		case "deregistration_delay.timeout_seconds":
			timeout, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("error converting deregistration_delay.timeout_seconds to int: %s", aws.StringValue(attr.Value))
			}
			d.Set("deregistration_delay", timeout)
		case "lambda.multi_value_headers.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("error converting lambda.multi_value_headers.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("lambda_multi_value_headers_enabled", enabled)
		case "proxy_protocol_v2.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("error converting proxy_protocol_v2.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("proxy_protocol_v2", enabled)
		case "slow_start.duration_seconds":
			slowStart, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("error converting slow_start.duration_seconds to int: %s", aws.StringValue(attr.Value))
			}
			d.Set("slow_start", slowStart)
		case "load_balancing.algorithm.type":
			loadBalancingAlgorithm := aws.StringValue(attr.Value)
			d.Set("load_balancing_algorithm_type", loadBalancingAlgorithm)
		case "preserve_client_ip.enabled":
			_, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("error converting preserve_client_ip.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("preserve_client_ip", attr.Value)
		}
	}

	stickinessAttr, err := flattenAwsLbTargetGroupStickiness(attrResp.Attributes)
	if err != nil {
		return fmt.Errorf("error flattening stickiness: %w", err)
	}

	if err := d.Set("stickiness", stickinessAttr); err != nil {
		return fmt.Errorf("error setting stickiness: %w", err)
	}

	tags, err := keyvaluetags.Elbv2ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for LB Target Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
