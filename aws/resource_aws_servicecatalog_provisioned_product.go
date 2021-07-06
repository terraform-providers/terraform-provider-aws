package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	tfservicecatalog "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsServiceCatalogProvisionedProduct() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogProvisionedProductCreate,
		Read:   resourceAwsServiceCatalogProvisionedProductRead,
		Update: resourceAwsServiceCatalogProvisionedProductUpdate,
		Delete: resourceAwsServiceCatalogProvisionedProductDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "en",
				ValidateFunc: validation.StringInSlice(tfservicecatalog.AcceptLanguage_Values(), false),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_dashboard_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ignore_errors": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"last_provisioning_record_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_record_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_successful_provisioning_record_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"notification_arns": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"path_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ConflictsWith: []string{
					"path_name",
				},
			},
			"path_name": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"path_id",
				},
			},
			"product_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ExactlyOneOf: []string{
					"product_id",
					"product_name",
				},
			},
			"product_name": {
				Type:     schema.TypeString,
				Optional: true,
				ExactlyOneOf: []string{
					"product_id",
					"product_name",
				},
			},
			"provisioning_artifact_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ExactlyOneOf: []string{
					"provisioning_artifact_id",
					"provisioning_artifact_name",
				},
			},
			"provisioning_artifact_name": {
				Type:     schema.TypeString,
				Optional: true,
				ExactlyOneOf: []string{
					"provisioning_artifact_id",
					"provisioning_artifact_name",
				},
			},
			"provisioning_parameters": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"use_previous_value": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"retain_physical_resources": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"stack_set_provisioning_preferences": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accounts": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"failure_tolerance_count": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.failure_tolerance_count",
								"stack_set_provisioning_preferences.0.failure_tolerance_percentage",
							},
						},
						"failure_tolerance_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.failure_tolerance_count",
								"stack_set_provisioning_preferences.0.failure_tolerance_percentage",
							},
						},
						"max_concurrency_count": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.max_concurrency_count",
								"stack_set_provisioning_preferences.0.max_concurrency_percentage",
							},
						},
						"max_concurrency_percentage": {
							Type:     schema.TypeInt,
							Optional: true,
							ExactlyOneOf: []string{
								"stack_set_provisioning_preferences.0.max_concurrency_count",
								"stack_set_provisioning_preferences.0.max_concurrency_percentage",
							},
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsServiceCatalogProvisionedProductCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &servicecatalog.ProvisionProductInput{
		ProvisionToken:         aws.String(resource.UniqueId()),
		ProvisionedProductName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_arns"); ok && len(v.([]interface{})) > 0 {
		input.NotificationArns = expandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("path_id"); ok {
		input.PathId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("path_name"); ok {
		input.PathName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_id"); ok {
		input.ProductId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_name"); ok {
		input.ProductName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_artifact_id"); ok {
		input.ProvisioningArtifactId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_artifact_name"); ok {
		input.ProvisioningArtifactName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_parameters"); ok && len(v.([]interface{})) > 0 {
		input.ProvisioningParameters = expandServiceCatalogProvisioningParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("stack_set_provisioning_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ProvisioningPreferences = expandServiceCatalogProvisioningPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().ServicecatalogTags()
	}

	var output *servicecatalog.ProvisionProductOutput

	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.ProvisionProduct(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.ProvisionProduct(input)
	}

	if err != nil {
		return fmt.Errorf("error provisioning Service Catalog Product: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error provisioning Service Catalog Product: empty response")
	}

	if output.RecordDetail == nil {
		return fmt.Errorf("error provisioning Service Catalog Product: no product view detail or summary")
	}

	d.SetId(aws.StringValue(output.RecordDetail.ProvisionedProductId))

	return resourceAwsServiceCatalogProvisionedProductRead(d, meta)
}

func resourceAwsServiceCatalogProvisionedProductRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	// There are two API operations for getting information about provisioned products:
	// 1. DescribeProvisionedProduct (used in waiter.ProvisionedProductReady) and
	// 2. DescribeRecord (used in waiter.RecordReady)

	// They provide some overlapping information. Most of the unique information available from
	// DescribeRecord is available in the data source aws_servicecatalog_record.

	acceptLanguage := tfservicecatalog.AcceptLanguageEnglish

	if v, ok := d.GetOk("accept_language"); ok {
		acceptLanguage = v.(string)
	}

	output, err := waiter.ProvisionedProductReady(conn, acceptLanguage, d.Id(), "")

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Provisioned Product (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Provisioned Product (%s): %w", d.Id(), err)
	}

	if output == nil || output.ProvisionedProductDetail == nil {
		return fmt.Errorf("error getting Service Catalog Provisioned Product (%s): empty response", d.Id())
	}

	detail := output.ProvisionedProductDetail

	d.Set("arn", detail.Arn)
	d.Set("cloudwatch_dashboard_names", aws.StringValueSlice(flattenServiceCatalogCloudWatchDashboards(output.CloudWatchDashboards)))

	if detail.CreatedTime != nil {
		d.Set("created_time", detail.CreatedTime.Format(time.RFC3339))
	} else {
		d.Set("created_time", nil)
	}

	d.Set("last_provisioning_record_id", detail.LastProvisioningRecordId)
	d.Set("last_record_id", detail.LastRecordId)
	d.Set("last_successful_provisioning_record_id", detail.LastSuccessfulProvisioningRecordId)
	d.Set("launch_role_arn", detail.LaunchRoleArn)
	d.Set("name", detail.Name)
	d.Set("product_id", detail.ProductId)
	d.Set("provisioning_artifact_id", detail.ProvisioningArtifactId)
	d.Set("status", detail.Status)
	d.Set("status_message", detail.StatusMessage)
	d.Set("type", detail.Type)

	// tags are only available from the record tied to the provisioned product

	recordOutput, err := waiter.RecordReady(conn, acceptLanguage, aws.StringValue(detail.LastProvisioningRecordId))

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Provisioned Product (%s) Record (%s) not found, unable to set tags", d.Id(), aws.StringValue(detail.LastProvisioningRecordId))
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog Provisioned Product (%s) Record (%s): %w", d.Id(), aws.StringValue(detail.LastProvisioningRecordId), err)
	}

	if recordOutput == nil || recordOutput.RecordDetail == nil {
		return fmt.Errorf("error getting Service Catalog Provisioned Product (%s) Record (%s): empty response", d.Id(), aws.StringValue(detail.LastProvisioningRecordId))
	}

	d.Set("path_id", recordOutput.RecordDetail.PathId)

	tags := keyvaluetags.ServicecatalogRecordKeyValueTags(recordOutput.RecordDetail.RecordTags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsServiceCatalogProvisionedProductUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.UpdateProvisionedProductInput{
		UpdateToken:          aws.String(resource.UniqueId()),
		ProvisionedProductId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("path_id"); ok {
		input.PathId = aws.String(v.(string))
	} else if v, ok := d.GetOk("path_name"); ok {
		input.PathName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("product_id"); ok {
		input.ProductId = aws.String(v.(string))
	} else if v, ok := d.GetOk("product_name"); ok {
		input.ProductName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_artifact_id"); ok {
		input.ProvisioningArtifactId = aws.String(v.(string))
	} else if v, ok := d.GetOk("provisioning_artifact_name"); ok {
		input.ProvisioningArtifactName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("provisioning_parameters"); ok && len(v.([]interface{})) > 0 {
		input.ProvisioningParameters = expandServiceCatalogUpdateProvisioningParameters(v.([]interface{}))
	}

	if v, ok := d.GetOk("stack_set_provisioning_preferences"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ProvisioningPreferences = expandServiceCatalogUpdateProvisioningPreferences(v.([]interface{})[0].(map[string]interface{}))
	}

	if d.HasChanges("tags", "tags_all") {
		defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
		tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

		if len(tags) > 0 {
			input.Tags = tags.IgnoreAws().ServicecatalogTags()
		} else {
			input.Tags = nil
		}
	}

	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.UpdateProvisionedProduct(input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateProvisionedProduct(input)
	}

	if err != nil {
		return fmt.Errorf("error updating Service Catalog Provisioned Product (%s): %w", d.Id(), err)
	}

	return resourceAwsServiceCatalogProvisionedProductRead(d, meta)
}

func resourceAwsServiceCatalogProvisionedProductDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	input := &servicecatalog.TerminateProvisionedProductInput{
		TerminateToken:       aws.String(resource.UniqueId()),
		ProvisionedProductId: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ignore_errors"); ok {
		input.IgnoreErrors = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("retain_physical_resources"); ok {
		input.RetainPhysicalResources = aws.Bool(v.(bool))
	}

	_, err := conn.TerminateProvisionedProduct(input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error terminating Service Catalog Provisioned Product (%s): %w", d.Id(), err)
	}

	err = waiter.ProvisionedProductTerminated(conn, d.Get("accept_language").(string), d.Id(), "")

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error waiting for Service Catalog Provisioned Product (%s) to be terminated: %w", d.Id(), err)
	}

	return nil
}

func expandServiceCatalogProvisioningParameter(tfMap map[string]interface{}) *servicecatalog.ProvisioningParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.ProvisioningParameter{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandServiceCatalogProvisioningParameters(tfList []interface{}) []*servicecatalog.ProvisioningParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*servicecatalog.ProvisioningParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandServiceCatalogProvisioningParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandServiceCatalogProvisioningPreferences(tfMap map[string]interface{}) *servicecatalog.ProvisioningPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.ProvisioningPreferences{}

	if v, ok := tfMap["account"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetAccounts = expandStringList(v)
	}

	if v, ok := tfMap["failure_tolerance_count"].(int64); ok && v != 0 {
		apiObject.StackSetFailureToleranceCount = aws.Int64(v)
	}

	if v, ok := tfMap["failure_tolerance_percentage"].(int64); ok && v != 0 {
		apiObject.StackSetFailureTolerancePercentage = aws.Int64(v)
	}

	if v, ok := tfMap["max_concurrency_count"].(int64); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyCount = aws.Int64(v)
	}

	if v, ok := tfMap["max_concurrency_percentage"].(int64); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyPercentage = aws.Int64(v)
	}

	if v, ok := tfMap["regions"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetRegions = expandStringList(v)
	}

	return apiObject
}

func expandServiceCatalogUpdateProvisioningParameter(tfMap map[string]interface{}) *servicecatalog.UpdateProvisioningParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.UpdateProvisioningParameter{}

	if v, ok := tfMap["key"].(string); ok && v != "" {
		apiObject.Key = aws.String(v)
	}

	if v, ok := tfMap["use_previous_value"].(bool); ok && v {
		apiObject.UsePreviousValue = aws.Bool(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		apiObject.Value = aws.String(v)
	}

	return apiObject
}

func expandServiceCatalogUpdateProvisioningParameters(tfList []interface{}) []*servicecatalog.UpdateProvisioningParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*servicecatalog.UpdateProvisioningParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandServiceCatalogUpdateProvisioningParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandServiceCatalogUpdateProvisioningPreferences(tfMap map[string]interface{}) *servicecatalog.UpdateProvisioningPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &servicecatalog.UpdateProvisioningPreferences{}

	if v, ok := tfMap["account"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetAccounts = expandStringList(v)
	}

	if v, ok := tfMap["failure_tolerance_count"].(int64); ok && v != 0 {
		apiObject.StackSetFailureToleranceCount = aws.Int64(v)
	}

	if v, ok := tfMap["failure_tolerance_percentage"].(int64); ok && v != 0 {
		apiObject.StackSetFailureTolerancePercentage = aws.Int64(v)
	}

	if v, ok := tfMap["max_concurrency_count"].(int64); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyCount = aws.Int64(v)
	}

	if v, ok := tfMap["max_concurrency_percentage"].(int64); ok && v != 0 {
		apiObject.StackSetMaxConcurrencyPercentage = aws.Int64(v)
	}

	if v, ok := tfMap["regions"].([]interface{}); ok && len(v) > 0 {
		apiObject.StackSetRegions = expandStringList(v)
	}

	return apiObject
}

func flattenServiceCatalogCloudWatchDashboards(apiObjects []*servicecatalog.CloudWatchDashboard) []*string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []*string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, apiObject.Name)
	}

	return tfList
}
