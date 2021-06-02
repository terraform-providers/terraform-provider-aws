package aws

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/securityhub/finder"
)

func resourceAwsSecurityHubInsight() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsSecurityHubInsightCreate,
		ReadWithoutTimeout:   resourceAwsSecurityHubInsightRead,
		UpdateWithoutTimeout: resourceAwsSecurityHubInsightUpdate,
		DeleteWithoutTimeout: resourceAwsSecurityHubInsightDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"filters": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_account_id":                      stringFilterSchema(),
						"company_name":                        stringFilterSchema(),
						"compliance_status":                   stringFilterSchema(),
						"confidence":                          numberFilterSchema(),
						"created_at":                          dateFilterSchema(),
						"criticality":                         numberFilterSchema(),
						"description":                         stringFilterSchema(),
						"finding_provider_fields_confidence":  numberFilterSchema(),
						"finding_provider_fields_criticality": numberFilterSchema(),
						"finding_provider_fields_related_findings_id":          stringFilterSchema(),
						"finding_provider_fields_related_findings_product_arn": stringFilterSchema(),
						"finding_provider_fields_severity_label":               stringFilterSchema(),
						"finding_provider_fields_severity_original":            stringFilterSchema(),
						"finding_provider_fields_types":                        stringFilterSchema(),
						"first_observed_at":                                    dateFilterSchema(),
						"generator_id":                                         stringFilterSchema(),
						"id":                                                   stringFilterSchema(),
						"keyword":                                              keywordFilterSchema(),
						"last_observed_at":                                     dateFilterSchema(),
						"malware_name":                                         stringFilterSchema(),
						"malware_path":                                         stringFilterSchema(),
						"malware_state":                                        stringFilterSchema(),
						"malware_type":                                         stringFilterSchema(),
						"network_destination_domain":                           stringFilterSchema(),
						"network_destination_ipv4":                             ipFilterSchema(),
						"network_destination_ipv6":                             ipFilterSchema(),
						"network_destination_port":                             numberFilterSchema(),
						"network_direction":                                    stringFilterSchema(),
						"network_protocol":                                     stringFilterSchema(),
						"network_source_domain":                                stringFilterSchema(),
						"network_source_ipv4":                                  ipFilterSchema(),
						"network_source_ipv6":                                  ipFilterSchema(),
						"network_source_mac":                                   stringFilterSchema(),
						"network_source_port":                                  numberFilterSchema(),
						"note_text":                                            stringFilterSchema(),
						"note_updated_at":                                      dateFilterSchema(),
						"note_updated_by":                                      stringFilterSchema(),
						"process_launched_at":                                  dateFilterSchema(),
						"process_name":                                         stringFilterSchema(),
						"process_parent_pid":                                   numberFilterSchema(),
						"process_path":                                         stringFilterSchema(),
						"process_pid":                                          numberFilterSchema(),
						"process_terminated_at":                                dateFilterSchema(),
						"product_arn":                                          stringFilterSchema(),
						"product_fields":                                       mapFilterSchema(),
						"product_name":                                         stringFilterSchema(),
						"recommendation_text":                                  stringFilterSchema(),
						"record_state":                                         stringFilterSchema(),
						"related_findings_id":                                  stringFilterSchema(),
						"related_findings_product_arn":                         stringFilterSchema(),
						"resource_aws_ec2_instance_iam_instance_profile_arn": stringFilterSchema(),
						"resource_aws_ec2_instance_image_id":                 stringFilterSchema(),
						"resource_aws_ec2_instance_ipv4_addresses":           ipFilterSchema(),
						"resource_aws_ec2_instance_ipv6_addresses":           ipFilterSchema(),
						"resource_aws_ec2_instance_key_name":                 stringFilterSchema(),
						"resource_aws_ec2_instance_launched_at":              dateFilterSchema(),
						"resource_aws_ec2_instance_subnet_id":                stringFilterSchema(),
						"resource_aws_ec2_instance_type":                     stringFilterSchema(),
						"resource_aws_ec2_instance_vpc_id":                   stringFilterSchema(),
						"resource_aws_iam_access_key_created_at":             dateFilterSchema(),
						"resource_aws_iam_access_key_status":                 stringFilterSchema(),
						"resource_aws_iam_access_key_user_name":              stringFilterSchema(),
						"resource_aws_s3_bucket_owner_id":                    stringFilterSchema(),
						"resource_aws_s3_bucket_owner_name":                  stringFilterSchema(),
						"resource_container_image_id":                        stringFilterSchema(),
						"resource_container_image_name":                      stringFilterSchema(),
						"resource_container_launched_at":                     dateFilterSchema(),
						"resource_container_name":                            stringFilterSchema(),
						"resource_details_other":                             mapFilterSchema(),
						"resource_id":                                        stringFilterSchema(),
						"resource_partition":                                 stringFilterSchema(),
						"resource_region":                                    stringFilterSchema(),
						"resource_tags":                                      mapFilterSchema(),
						"resource_type":                                      stringFilterSchema(),
						"severity_label":                                     stringFilterSchema(),
						"source_url":                                         stringFilterSchema(),
						"threat_intel_indicator_category":                    stringFilterSchema(),
						"threat_intel_indicator_last_observed_at":            dateFilterSchema(),
						"threat_intel_indicator_source":                      stringFilterSchema(),
						"threat_intel_indicator_source_url":                  stringFilterSchema(),
						"threat_intel_indicator_type":                        stringFilterSchema(),
						"threat_intel_indicator_value":                       stringFilterSchema(),
						"title":                                              stringFilterSchema(),
						"type":                                               stringFilterSchema(),
						"updated_at":                                         dateFilterSchema(),
						"user_defined_values":                                mapFilterSchema(),
						"verification_state":                                 stringFilterSchema(),
						"workflow_status":                                    workflowStatusSchema(),
					},
				},
			},

			"group_by_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsSecurityHubInsightCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).securityhubconn

	name := d.Get("name").(string)

	input := &securityhub.CreateInsightInput{
		GroupByAttribute: aws.String(d.Get("group_by_attribute").(string)),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("filters"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Filters = expandSecurityHubSecurityFindingFilters(v.([]interface{}))
	}

	output, err := conn.CreateInsightWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Security Hub Insight (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating Security Hub Insight (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.InsightArn))

	return resourceAwsSecurityHubInsightRead(ctx, d, meta)
}

func resourceAwsSecurityHubInsightRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).securityhubconn

	insight, err := finder.Insight(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Security Hub Insight (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Security Hub Insight (%s): %w", d.Id(), err))
	}

	if insight == nil {
		if d.IsNewResource() {
			return diag.FromErr(fmt.Errorf("error reading Security Hub Insight (%s): empty output", d.Id()))
		}
		log.Printf("[WARN] Security Hub Insight (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", insight.InsightArn)
	if err := d.Set("filters", flattenSecurityHubSecurityFindingFilters(insight.Filters)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting filters: %w", err))
	}
	d.Set("group_by_attribute", insight.GroupByAttribute)
	d.Set("name", insight.Name)

	return nil
}

func resourceAwsSecurityHubInsightUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).securityhubconn

	input := &securityhub.UpdateInsightInput{
		InsightArn: aws.String(d.Id()),
	}

	if d.HasChange("filters") {
		input.Filters = expandSecurityHubSecurityFindingFilters(d.Get("filters").([]interface{}))
	}

	if d.HasChange("group_by_attribute") {
		input.GroupByAttribute = aws.String(d.Get("group_by_attribute").(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	_, err := conn.UpdateInsightWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Security Hub Insight (%s): %w", d.Id(), err))
	}

	return resourceAwsSecurityHubInsightRead(ctx, d, meta)
}

func resourceAwsSecurityHubInsightDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).securityhubconn

	input := &securityhub.DeleteInsightInput{
		InsightArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteInsightWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Security Hub Insight (%s): %w", d.Id(), err))
	}

	return nil
}

func dateFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_range": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"unit": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(securityhub.DateRangeUnit_Values(), true),
							},
							"value": {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},
				"end": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"start": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func ipFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cidr": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateCIDRNetworkAddress,
				},
			},
		},
	}
}

func keywordFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func mapFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(securityhub.MapFilterComparison_Values(), false),
				},
				"key": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func numberFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"eq": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateTypeStringNullableFloat,
				},
				"gte": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateTypeStringNullableFloat,
				},
				"lte": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateTypeStringNullableFloat,
				},
			},
		},
	}
}

func stringFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(securityhub.StringFilterComparison_Values(), false),
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func workflowStatusSchema() *schema.Schema {
	s := stringFilterSchema()

	s.Elem.(*schema.Resource).Schema["value"].ValidateFunc = validation.StringInSlice(securityhub.WorkflowStatus_Values(), false)

	return s
}

func expandSecurityHubDateFilterDateRange(l []interface{}) *securityhub.DateRange {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dr := &securityhub.DateRange{}

	if v, ok := tfMap["unit"].(string); ok && v != "" {
		dr.Unit = aws.String(v)
	}

	if v, ok := tfMap["value"].(int); ok {
		dr.Value = aws.Int64(int64(v))
	}

	return dr
}

func expandSecurityHubDateFilters(l []interface{}) []*securityhub.DateFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var dateFilters []*securityhub.DateFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		df := &securityhub.DateFilter{}

		if v, ok := tfMap["date_range"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			df.DateRange = expandSecurityHubDateFilterDateRange(v)
		}

		if v, ok := tfMap["end"].(string); ok && v != "" {
			df.End = aws.String(v)
		}

		if v, ok := tfMap["start"].(string); ok && v != "" {
			df.Start = aws.String(v)
		}

		dateFilters = append(dateFilters, df)
	}

	return dateFilters
}

func expandSecurityHubSecurityFindingFilters(l []interface{}) *securityhub.AwsSecurityFindingFilters {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filters := &securityhub.AwsSecurityFindingFilters{}

	if v, ok := tfMap["aws_account_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.AwsAccountId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["company_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.CompanyName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_status"].(*schema.Set); ok && v.Len() > 0 {
		filters.ComplianceStatus = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["confidence"].(*schema.Set); ok && v.Len() > 0 {
		filters.Confidence = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["created_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.CreatedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["criticality"].(*schema.Set); ok && v.Len() > 0 {
		filters.Criticality = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["description"].(*schema.Set); ok && v.Len() > 0 {
		filters.Description = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_confidence"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsConfidence = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_criticality"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsCriticality = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_related_findings_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsRelatedFindingsId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_related_findings_product_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsRelatedFindingsProductArn = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_severity_label"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsSeverityLabel = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_severity_original"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsSeverityOriginal = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_types"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsTypes = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["first_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.FirstObservedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["generator_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.GeneratorId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["id"].(*schema.Set); ok && v.Len() > 0 {
		filters.Id = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["keyword"].(*schema.Set); ok && v.Len() > 0 {
		filters.Keyword = expandSecurityHubKeywordFilters(v.List())
	}

	if v, ok := tfMap["last_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.LastObservedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["malware_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwareName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["malware_path"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwarePath = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["malware_state"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwareState = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["malware_type"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwareType = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["network_destination_domain"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationDomain = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["network_destination_ipv4"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationIpV4 = expandSecurityHubIpFilters(v.List())
	}

	if v, ok := tfMap["network_destination_ipv6"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationIpV6 = expandSecurityHubIpFilters(v.List())
	}

	if v, ok := tfMap["network_destination_port"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationPort = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["network_direction"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDirection = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["network_protocol"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkProtocol = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_domain"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceDomain = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_ipv4"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceIpV4 = expandSecurityHubIpFilters(v.List())
	}

	if v, ok := tfMap["network_source_ipv6"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceIpV6 = expandSecurityHubIpFilters(v.List())
	}

	if v, ok := tfMap["network_source_mac"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceMac = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_port"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourcePort = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["note_text"].(*schema.Set); ok && v.Len() > 0 {
		filters.NoteText = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["note_updated_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.NoteUpdatedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["note_updated_by"].(*schema.Set); ok && v.Len() > 0 {
		filters.NoteUpdatedBy = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["process_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessLaunchedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["process_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["process_parent_pid"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessParentPid = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["process_path"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessPath = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["process_pid"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessPid = expandSecurityHubNumberFilters(v.List())
	}

	if v, ok := tfMap["process_terminated_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessTerminatedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["product_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProductArn = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["product_fields"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProductFields = expandSecurityHubMapFilters(v.List())
	}

	if v, ok := tfMap["product_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProductName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["recommendation_text"].(*schema.Set); ok && v.Len() > 0 {
		filters.RecommendationText = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["record_state"].(*schema.Set); ok && v.Len() > 0 {
		filters.RecordState = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.RelatedFindingsId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_product_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.RelatedFindingsProductArn = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_iam_instance_profile_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceIamInstanceProfileArn = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_image_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceImageId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_ipv4_addresses"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceIpV4Addresses = expandSecurityHubIpFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_ipv6_addresses"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceIpV6Addresses = expandSecurityHubIpFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_key_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceKeyName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceLaunchedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_subnet_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceSubnetId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_type"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceType = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_vpc_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceVpcId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_created_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsIamAccessKeyCreatedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_status"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsIamAccessKeyStatus = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_user_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsIamAccessKeyUserName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_s3_bucket_owner_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsS3BucketOwnerId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_s3_bucket_owner_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsS3BucketOwnerName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_image_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerImageId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_image_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerImageName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerLaunchedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["resource_container_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerName = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_details_other"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceDetailsOther = expandSecurityHubMapFilters(v.List())
	}

	if v, ok := tfMap["resource_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceId = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_partition"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourcePartition = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_region"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceRegion = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["resource_tags"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceTags = expandSecurityHubMapFilters(v.List())
	}

	if v, ok := tfMap["resource_type"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceType = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["severity_label"].(*schema.Set); ok && v.Len() > 0 {
		filters.SeverityLabel = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["source_url"].(*schema.Set); ok && v.Len() > 0 {
		filters.SourceUrl = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_category"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorCategory = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_last_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorLastObservedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_source"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorSource = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_source_url"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorSourceUrl = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_type"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorType = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_value"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorValue = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["title"].(*schema.Set); ok && v.Len() > 0 {
		filters.Title = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["type"].(*schema.Set); ok && v.Len() > 0 {
		filters.Type = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["updated_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.UpdatedAt = expandSecurityHubDateFilters(v.List())
	}

	if v, ok := tfMap["user_defined_values"].(*schema.Set); ok && v.Len() > 0 {
		filters.UserDefinedFields = expandSecurityHubMapFilters(v.List())
	}

	if v, ok := tfMap["verification_state"].(*schema.Set); ok && v.Len() > 0 {
		filters.VerificationState = expandSecurityHubStringFilters(v.List())
	}

	if v, ok := tfMap["workflow_status"].(*schema.Set); ok && v.Len() > 0 {
		filters.WorkflowStatus = expandSecurityHubStringFilters(v.List())
	}

	return filters
}

func expandSecurityHubIpFilters(l []interface{}) []*securityhub.IpFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var ipFilters []*securityhub.IpFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		ipFilter := &securityhub.IpFilter{}

		if v, ok := tfMap["cidr"].(string); ok && v != "" {
			ipFilter.Cidr = aws.String(v)
		}

		ipFilters = append(ipFilters, ipFilter)
	}

	return ipFilters
}

func expandSecurityHubKeywordFilters(l []interface{}) []*securityhub.KeywordFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var keywordFilters []*securityhub.KeywordFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		kf := &securityhub.KeywordFilter{}

		if v, ok := tfMap["value"].(string); ok && v != "" {
			kf.Value = aws.String(v)
		}

		keywordFilters = append(keywordFilters, kf)
	}

	return keywordFilters
}

func expandSecurityHubMapFilters(l []interface{}) []*securityhub.MapFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var mapFilters []*securityhub.MapFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		mf := &securityhub.MapFilter{}

		if v, ok := tfMap["comparison"].(string); ok && v != "" {
			mf.Comparison = aws.String(v)
		}

		if v, ok := tfMap["key"].(string); ok && v != "" {
			mf.Key = aws.String(v)
		}

		if v, ok := tfMap["value"].(string); ok && v != "" {
			mf.Value = aws.String(v)
		}

		mapFilters = append(mapFilters, mf)
	}

	return mapFilters
}

func expandSecurityHubNumberFilters(l []interface{}) []*securityhub.NumberFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var numFilters []*securityhub.NumberFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		nf := &securityhub.NumberFilter{}

		if v, ok := tfMap["eq"].(string); ok && v != "" {
			val, err := strconv.ParseFloat(v, 64)
			if err == nil {
				nf.Eq = aws.Float64(val)
			}
		}

		if v, ok := tfMap["gte"].(string); ok && v != "" {
			val, err := strconv.ParseFloat(v, 64)
			if err == nil {
				nf.Gte = aws.Float64(val)
			}
		}

		if v, ok := tfMap["lte"].(string); ok && v != "" {
			val, err := strconv.ParseFloat(v, 64)
			if err == nil {
				nf.Lte = aws.Float64(val)
			}
		}

		numFilters = append(numFilters, nf)
	}

	return numFilters
}

func expandSecurityHubStringFilters(l []interface{}) []*securityhub.StringFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var stringFilters []*securityhub.StringFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		sf := &securityhub.StringFilter{}

		if v, ok := tfMap["comparison"].(string); ok && v != "" {
			sf.Comparison = aws.String(v)
		}

		if v, ok := tfMap["value"].(string); ok && v != "" {
			sf.Value = aws.String(v)
		}

		stringFilters = append(stringFilters, sf)
	}

	return stringFilters
}

func flattenSecurityHubDateFilterDateRange(dateRange *securityhub.DateRange) []interface{} {
	if dateRange == nil {
		return nil
	}

	m := map[string]interface{}{
		"unit":  aws.StringValue(dateRange.Unit),
		"value": aws.Int64Value(dateRange.Value),
	}

	return []interface{}{m}
}

func flattenSecurityHubDateFilters(filters []*securityhub.DateFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var dateFilters []interface{}

	for _, filter := range filters {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{
			"date_range": flattenSecurityHubDateFilterDateRange(filter.DateRange),
			"end":        aws.StringValue(filter.End),
			"start":      aws.StringValue(filter.Start),
		}

		dateFilters = append(dateFilters, m)
	}

	return dateFilters
}

func flattenSecurityHubIpFilters(filters []*securityhub.IpFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var ipFilters []interface{}

	for _, filter := range filters {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{
			"cidr": aws.StringValue(filter.Cidr),
		}

		ipFilters = append(ipFilters, m)
	}

	return ipFilters
}

func flattenSecurityHubKeywordFilters(filters []*securityhub.KeywordFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var keywordFilters []interface{}

	for _, filter := range filters {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{
			"value": aws.StringValue(filter.Value),
		}

		keywordFilters = append(keywordFilters, m)
	}

	return keywordFilters
}

func flattenSecurityHubMapFilters(filters []*securityhub.MapFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var mapFilters []interface{}

	for _, filter := range filters {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{
			"comparison": aws.StringValue(filter.Comparison),
			"key":        aws.StringValue(filter.Key),
			"value":      aws.StringValue(filter.Value),
		}

		mapFilters = append(mapFilters, m)
	}

	return mapFilters
}

func flattenSecurityHubNumberFilters(filters []*securityhub.NumberFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var numFilters []interface{}

	for _, filter := range filters {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{}

		if filter.Eq != nil {
			m["eq"] = strconv.FormatFloat(aws.Float64Value(filter.Eq), 'f', -1, 64)
		}

		if filter.Gte != nil {
			m["gte"] = strconv.FormatFloat(aws.Float64Value(filter.Gte), 'f', -1, 64)
		}

		if filter.Lte != nil {
			m["lte"] = strconv.FormatFloat(aws.Float64Value(filter.Lte), 'f', -1, 64)
		}

		numFilters = append(numFilters, m)
	}

	return numFilters
}

func flattenSecurityHubSecurityFindingFilters(filters *securityhub.AwsSecurityFindingFilters) []interface{} {
	if filters == nil {
		return nil
	}

	m := map[string]interface{}{
		"aws_account_id":                      flattenSecurityHubStringFilters(filters.AwsAccountId),
		"company_name":                        flattenSecurityHubStringFilters(filters.CompanyName),
		"compliance_status":                   flattenSecurityHubStringFilters(filters.ComplianceStatus),
		"confidence":                          flattenSecurityHubNumberFilters(filters.Confidence),
		"created_at":                          flattenSecurityHubDateFilters(filters.CreatedAt),
		"criticality":                         flattenSecurityHubNumberFilters(filters.Criticality),
		"description":                         flattenSecurityHubStringFilters(filters.Description),
		"finding_provider_fields_confidence":  flattenSecurityHubNumberFilters(filters.FindingProviderFieldsConfidence),
		"finding_provider_fields_criticality": flattenSecurityHubNumberFilters(filters.FindingProviderFieldsCriticality),
		"finding_provider_fields_related_findings_id":          flattenSecurityHubStringFilters(filters.FindingProviderFieldsRelatedFindingsId),
		"finding_provider_fields_related_findings_product_arn": flattenSecurityHubStringFilters(filters.FindingProviderFieldsRelatedFindingsProductArn),
		"finding_provider_fields_severity_label":               flattenSecurityHubStringFilters(filters.FindingProviderFieldsSeverityLabel),
		"finding_provider_fields_severity_original":            flattenSecurityHubStringFilters(filters.FindingProviderFieldsSeverityOriginal),
		"finding_provider_fields_types":                        flattenSecurityHubStringFilters(filters.FindingProviderFieldsTypes),
		"first_observed_at":                                    flattenSecurityHubDateFilters(filters.FirstObservedAt),
		"generator_id":                                         flattenSecurityHubStringFilters(filters.GeneratorId),
		"id":                                                   flattenSecurityHubStringFilters(filters.Id),
		"keyword":                                              flattenSecurityHubKeywordFilters(filters.Keyword),
		"last_observed_at":                                     flattenSecurityHubDateFilters(filters.LastObservedAt),
		"malware_name":                                         flattenSecurityHubStringFilters(filters.MalwareName),
		"malware_path":                                         flattenSecurityHubStringFilters(filters.MalwarePath),
		"malware_state":                                        flattenSecurityHubStringFilters(filters.MalwareState),
		"malware_type":                                         flattenSecurityHubStringFilters(filters.MalwareType),
		"network_destination_domain":                           flattenSecurityHubStringFilters(filters.NetworkDestinationDomain),
		"network_destination_ipv4":                             flattenSecurityHubIpFilters(filters.NetworkDestinationIpV4),
		"network_destination_ipv6":                             flattenSecurityHubIpFilters(filters.NetworkDestinationIpV6),
		"network_destination_port":                             flattenSecurityHubNumberFilters(filters.NetworkDestinationPort),
		"network_direction":                                    flattenSecurityHubStringFilters(filters.NetworkDirection),
		"network_protocol":                                     flattenSecurityHubStringFilters(filters.NetworkProtocol),
		"network_source_domain":                                flattenSecurityHubStringFilters(filters.NetworkSourceDomain),
		"network_source_ipv4":                                  flattenSecurityHubIpFilters(filters.NetworkSourceIpV4),
		"network_source_ipv6":                                  flattenSecurityHubIpFilters(filters.NetworkSourceIpV6),
		"network_source_mac":                                   flattenSecurityHubStringFilters(filters.NetworkSourceMac),
		"network_source_port":                                  flattenSecurityHubNumberFilters(filters.NetworkSourcePort),
		"note_text":                                            flattenSecurityHubStringFilters(filters.NoteText),
		"note_updated_at":                                      flattenSecurityHubDateFilters(filters.NoteUpdatedAt),
		"note_updated_by":                                      flattenSecurityHubStringFilters(filters.NoteUpdatedBy),
		"process_launched_at":                                  flattenSecurityHubDateFilters(filters.ProcessLaunchedAt),
		"process_name":                                         flattenSecurityHubStringFilters(filters.ProcessName),
		"process_parent_pid":                                   flattenSecurityHubNumberFilters(filters.ProcessParentPid),
		"process_path":                                         flattenSecurityHubStringFilters(filters.ProcessPath),
		"process_pid":                                          flattenSecurityHubNumberFilters(filters.ProcessPid),
		"process_terminated_at":                                flattenSecurityHubDateFilters(filters.ProcessTerminatedAt),
		"product_arn":                                          flattenSecurityHubStringFilters(filters.ProductArn),
		"product_fields":                                       flattenSecurityHubMapFilters(filters.ProductFields),
		"product_name":                                         flattenSecurityHubStringFilters(filters.ProductName),
		"recommendation_text":                                  flattenSecurityHubStringFilters(filters.RecommendationText),
		"record_state":                                         flattenSecurityHubStringFilters(filters.RecordState),
		"related_findings_id":                                  flattenSecurityHubStringFilters(filters.RelatedFindingsId),
		"related_findings_product_arn":                         flattenSecurityHubStringFilters(filters.RelatedFindingsProductArn),
		"resource_aws_ec2_instance_iam_instance_profile_arn": flattenSecurityHubStringFilters(filters.ResourceAwsEc2InstanceIamInstanceProfileArn),
		"resource_aws_ec2_instance_image_id":                 flattenSecurityHubStringFilters(filters.ResourceAwsEc2InstanceImageId),
		"resource_aws_ec2_instance_ipv4_addresses":           flattenSecurityHubIpFilters(filters.ResourceAwsEc2InstanceIpV4Addresses),
		"resource_aws_ec2_instance_ipv6_addresses":           flattenSecurityHubIpFilters(filters.ResourceAwsEc2InstanceIpV6Addresses),
		"resource_aws_ec2_instance_key_name":                 flattenSecurityHubStringFilters(filters.ResourceAwsEc2InstanceKeyName),
		"resource_aws_ec2_instance_launched_at":              flattenSecurityHubDateFilters(filters.ResourceAwsEc2InstanceLaunchedAt),
		"resource_aws_ec2_instance_subnet_id":                flattenSecurityHubStringFilters(filters.ResourceAwsEc2InstanceSubnetId),
		"resource_aws_ec2_instance_type":                     flattenSecurityHubStringFilters(filters.ResourceAwsEc2InstanceType),
		"resource_aws_ec2_instance_vpc_id":                   flattenSecurityHubStringFilters(filters.ResourceAwsEc2InstanceVpcId),
		"resource_aws_iam_access_key_created_at":             flattenSecurityHubDateFilters(filters.ResourceAwsIamAccessKeyCreatedAt),
		"resource_aws_iam_access_key_status":                 flattenSecurityHubStringFilters(filters.ResourceAwsIamAccessKeyStatus),
		"resource_aws_iam_access_key_user_name":              flattenSecurityHubStringFilters(filters.ResourceAwsIamAccessKeyUserName),
		"resource_aws_s3_bucket_owner_id":                    flattenSecurityHubStringFilters(filters.ResourceAwsS3BucketOwnerId),
		"resource_aws_s3_bucket_owner_name":                  flattenSecurityHubStringFilters(filters.ResourceAwsS3BucketOwnerName),
		"resource_container_image_id":                        flattenSecurityHubStringFilters(filters.ResourceContainerImageId),
		"resource_container_image_name":                      flattenSecurityHubStringFilters(filters.ResourceContainerImageName),
		"resource_container_launched_at":                     flattenSecurityHubDateFilters(filters.ResourceContainerLaunchedAt),
		"resource_container_name":                            flattenSecurityHubStringFilters(filters.ResourceContainerName),
		"resource_details_other":                             flattenSecurityHubMapFilters(filters.ResourceDetailsOther),
		"resource_id":                                        flattenSecurityHubStringFilters(filters.ResourceId),
		"resource_partition":                                 flattenSecurityHubStringFilters(filters.ResourcePartition),
		"resource_region":                                    flattenSecurityHubStringFilters(filters.ResourceRegion),
		"resource_tags":                                      flattenSecurityHubMapFilters(filters.ResourceTags),
		"resource_type":                                      flattenSecurityHubStringFilters(filters.ResourceType),
		"severity_label":                                     flattenSecurityHubStringFilters(filters.SeverityLabel),
		"source_url":                                         flattenSecurityHubStringFilters(filters.ThreatIntelIndicatorSourceUrl),
		"threat_intel_indicator_category":                    flattenSecurityHubStringFilters(filters.ThreatIntelIndicatorCategory),
		"threat_intel_indicator_last_observed_at":            flattenSecurityHubDateFilters(filters.ThreatIntelIndicatorLastObservedAt),
		"threat_intel_indicator_source":                      flattenSecurityHubStringFilters(filters.ThreatIntelIndicatorSource),
		"threat_intel_indicator_source_url":                  flattenSecurityHubStringFilters(filters.ThreatIntelIndicatorSourceUrl),
		"threat_intel_indicator_type":                        flattenSecurityHubStringFilters(filters.ThreatIntelIndicatorType),
		"threat_intel_indicator_value":                       flattenSecurityHubStringFilters(filters.ThreatIntelIndicatorValue),
		"title":                                              flattenSecurityHubStringFilters(filters.Title),
		"type":                                               flattenSecurityHubStringFilters(filters.Type),
		"updated_at":                                         flattenSecurityHubDateFilters(filters.UpdatedAt),
		"user_defined_values":                                flattenSecurityHubMapFilters(filters.UserDefinedFields),
		"verification_state":                                 flattenSecurityHubStringFilters(filters.VerificationState),
		"workflow_status":                                    flattenSecurityHubStringFilters(filters.WorkflowStatus),
	}

	return []interface{}{m}
}

func flattenSecurityHubStringFilters(filters []*securityhub.StringFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var stringFilters []interface{}

	for _, filter := range filters {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{
			"comparison": aws.StringValue(filter.Comparison),
			"value":      aws.StringValue(filter.Value),
		}

		stringFilters = append(stringFilters, m)
	}

	return stringFilters
}
