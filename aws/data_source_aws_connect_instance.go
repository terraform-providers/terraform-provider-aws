package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var dataResourceConnectInstanceAttributesMapping = map[string]string{
	connect.InstanceAttributeTypeInboundCalls:          "inbound_calls_enabled",
	connect.InstanceAttributeTypeOutboundCalls:         "outbound_calls_enabled",
	connect.InstanceAttributeTypeContactflowLogs:       "contact_flow_logs_enabled",
	connect.InstanceAttributeTypeContactLens:           "contact_lens_enabled",
	connect.InstanceAttributeTypeAutoResolveBestVoices: "auto_resolve_best_voices",
	connect.InstanceAttributeTypeUseCustomTtsVoices:    "use_custom_tts_voices",
	connect.InstanceAttributeTypeEarlyMedia:            "early_media_enabled",
}

func dataSourceAwsConnectInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsConnectInstanceRead,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_alias": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"identity_management_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inbound_calls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"outbound_calls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"early_media_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"contact_flow_logs_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"contact_lens_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auto_resolve_best_voices": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"use_custom_tts_voices": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsConnectInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	var matchedInstance *connect.Instance

	instanceId, instanceIdOk := d.GetOk("instance_id")
	instanceAlias, instanceAliasOk := d.GetOk("instance_alias")

	if !instanceIdOk && !instanceAliasOk {
		return diag.FromErr(errors.New("error one instance_id or instance_alias of must be assigned"))
	}

	if instanceIdOk {
		input := connect.DescribeInstanceInput{
			InstanceId: aws.String(instanceId.(string)),
		}

		log.Printf("[DEBUG] Reading Connect Instance by instance_id: %s", input)

		output, err := conn.DescribeInstance(&input)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting Connect Instance by instance_id (%s): %s", instanceId, err))
		}

		matchedInstance = output.Instance

	} else if instanceAliasOk {
		instanceSummaryList, err := dataSourceAwsConnectGetAllConnectInstanceSummaries(ctx, conn)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing Connect Instances: %s", err))
		}

		for _, instanceSummary := range instanceSummaryList {
			log.Printf("[DEBUG] Connect Instance summary: %s", instanceSummary)
			if aws.StringValue(instanceSummary.InstanceAlias) == instanceAlias.(string) {

				matchedInstance = &connect.Instance{
					Arn:                    instanceSummary.Arn,
					CreatedTime:            instanceSummary.CreatedTime,
					Id:                     instanceSummary.Id,
					IdentityManagementType: instanceSummary.IdentityManagementType,
					InboundCallsEnabled:    instanceSummary.InboundCallsEnabled,
					InstanceAlias:          instanceSummary.InstanceAlias,
					InstanceStatus:         instanceSummary.InstanceStatus,
					OutboundCallsEnabled:   instanceSummary.OutboundCallsEnabled,
					ServiceRole:            instanceSummary.ServiceRole,
				}
				break
			}
		}
	}

	if matchedInstance == nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Instance by instance_alias: %s", instanceAlias))
	}

	d.SetId(aws.StringValue(matchedInstance.Id))

	d.Set("arn", matchedInstance.Arn)
	d.Set("created_time", matchedInstance.CreatedTime.Format(time.RFC3339))
	d.Set("identity_management_type", matchedInstance.IdentityManagementType)
	d.Set("instance_alias", matchedInstance.InstanceAlias)
	d.Set("inbound_calls_enabled", matchedInstance.InboundCallsEnabled)
	d.Set("outbound_calls_enabled", matchedInstance.OutboundCallsEnabled)
	d.Set("status", matchedInstance.InstanceStatus)
	d.Set("service_role", matchedInstance.ServiceRole)

	for att := range dataResourceConnectInstanceAttributesMapping {
		value, err := dataResourceAwsConnectInstanceReadAttribute(ctx, conn, d.Id(), att)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error reading Connect instance (%s) attribute (%s): %s", d.Id(), att, err))
		}
		d.Set(resourceConnectInstanceAttributesMapping[att], value)
	}
	return nil
}

func dataSourceAwsConnectGetAllConnectInstanceSummaries(ctx context.Context, conn *connect.Connect) ([]*connect.InstanceSummary, error) {
	var instances []*connect.InstanceSummary
	var nextToken string

	for {
		input := &connect.ListInstancesInput{
			// MaxResults Valid Range: Minimum value of 1. Maximum value of 60
			MaxResults: aws.Int64(int64(60)),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		log.Printf("[DEBUG] Listing Connect Instances: %s", input)

		output, err := conn.ListInstancesWithContext(ctx, input)
		if err != nil {
			return instances, err
		}
		instances = append(instances, output.InstanceSummaryList...)

		if output.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(output.NextToken)
	}

	return instances, nil
}

func dataResourceAwsConnectInstanceReadAttribute(ctx context.Context, conn *connect.Connect, instanceID string, attributeType string) (bool, error) {
	input := &connect.DescribeInstanceAttributeInput{
		InstanceId:    aws.String(instanceID),
		AttributeType: aws.String(attributeType),
	}

	out, err := conn.DescribeInstanceAttributeWithContext(ctx, input)

	if err != nil {
		return false, err
	}

	result, parseerr := strconv.ParseBool(*out.Attribute.Value)
	return result, parseerr
}
