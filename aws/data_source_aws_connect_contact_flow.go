package aws

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsConnectContactFlow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsConnectContactFlowRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"content": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func dataSourceAwsConnectContactFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var matchedInstance *connect.ContactFlow

	contactFlowID, contactFlowIDOk := d.GetOk("contact_flow_id")
	instanceID, instanceIDOk := d.GetOk("instance_id")
	name, nameOk := d.GetOk("name")

	if !instanceIDOk && (!contactFlowIDOk || !nameOk) {
		return diag.FromErr(errors.New("error instance_id and contact_flow_id or name of must be assigned"))
	}
	if contactFlowIDOk {
		resp, err := conn.DescribeContactFlow(&connect.DescribeContactFlowInput{
			ContactFlowId: aws.String(contactFlowID.(string)),
			InstanceId:    aws.String(instanceID.(string)),
		})
		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow by contact_flow_id (%s): %s", contactFlowID, err))
		}
		matchedInstance = resp.ContactFlow
	} else if nameOk {
		connectFlowSummaryList, err := dataSourceAwsConnectGetAllConnectContactFlowSummaries(ctx, conn, instanceID.(string))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing Connect Contact Flows: %s", err))
		}

		for _, connectFlowSummary := range connectFlowSummaryList {
			log.Printf("[DEBUG] Connect Contact flow summary: %s", connectFlowSummary)
			if aws.StringValue(connectFlowSummary.Name) == name.(string) {
				resp, err := conn.DescribeContactFlow(&connect.DescribeContactFlowInput{
					ContactFlowId: connectFlowSummary.Id,
					InstanceId:    aws.String(instanceID.(string)),
				})
				if err != nil {
					return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow by name (%s): %s", name, err))
				}
				matchedInstance = resp.ContactFlow
				break
			}
		}
	}

	if matchedInstance == nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Contact Flow by name: %s", name))
	}

	d.Set("arn", matchedInstance.Arn)
	d.Set("instance_id", instanceID)
	d.Set("contact_flow_id", matchedInstance.Id)
	d.Set("name", matchedInstance.Name)
	d.Set("description", matchedInstance.Description)
	d.Set("content", matchedInstance.Content)
	d.Set("type", matchedInstance.Type)
	if err := d.Set("tags", keyvaluetags.ConnectKeyValueTags(matchedInstance.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}
	d.SetId(fmt.Sprintf("%s:%s", instanceID, d.Get("contact_flow_id").(string)))

	return nil
}

func dataSourceAwsConnectGetAllConnectContactFlowSummaries(ctx context.Context, conn *connect.Connect, instanceID string) ([]*connect.ContactFlowSummary, error) {
	var instances []*connect.ContactFlowSummary
	var nextToken string

	for {
		input := &connect.ListContactFlowsInput{
			InstanceId: aws.String(instanceID),
			// MaxResults Valid Range: Minimum value of 1. Maximum value of 60
			MaxResults: aws.Int64(int64(60)),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}

		log.Printf("[DEBUG] Listing Connect Contact Flows: %s", input)

		output, err := conn.ListContactFlowsWithContext(ctx, input)
		if err != nil {
			return instances, err
		}
		instances = append(instances, output.ContactFlowSummaryList...)

		if output.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(output.NextToken)
	}

	return instances, nil
}
