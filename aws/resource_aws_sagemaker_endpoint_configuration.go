package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSagemakerEndpointConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerEndpointConfigurationCreate,
		Read:   resourceAwsSagemakerEndpointConfigurationRead,
		Update: resourceAwsSagemakerEndpointConfigurationUpdate,
		Delete: resourceAwsSagemakerEndpointConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateSagemakerName,
			},

			"production_variants": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"variant_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},

						"model_name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"initial_instance_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},

						"instance_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantInstanceType_Values(), false),
						},

						"initial_variant_weight": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatAtLeast(0),
							Default:      1,
						},

						"accelerator_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.ProductionVariantAcceleratorType_Values(), false),
						},
					},
				},
			},

			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"tags": tagsSchema(),

			"data_capture_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_capture": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},

						"initial_sampling_percentage": {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(0, 100),
						},

						"destination_s3_uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringMatch(regexp.MustCompile(`^(https|s3)://([^/])/?(.*)$`), ""),
								validation.StringLenBetween(1, 512),
							)},

						"kms_key_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},

						"capture_options": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 2,
							MinItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capture_mode": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(sagemaker.CaptureMode_Values(), false),
									},
								},
							},
						},

						"capture_content_type_header": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"csv_content_types": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 10,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*\/[a-zA-Z0-9](-*[a-zA-Z0-9.])*`), ""),
												validation.StringLenBetween(1, 256),
											),
										},
										Optional: true,
										ForceNew: true,
									},
									"json_content_types": {
										Type:     schema.TypeSet,
										MinItems: 1,
										MaxItems: 10,
										Elem: &schema.Schema{
											Type: schema.TypeString,
											ValidateFunc: validation.All(
												validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*\/[a-zA-Z0-9](-*[a-zA-Z0-9.])*`), ""),
												validation.StringLenBetween(1, 256),
											),
										},
										Optional: true,
										ForceNew: true,
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

func resourceAwsSagemakerEndpointConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.UniqueId()
	}

	createOpts := &sagemaker.CreateEndpointConfigInput{
		EndpointConfigName: aws.String(name),
		ProductionVariants: expandSagemakerProductionVariants(d.Get("production_variants").([]interface{})),
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		createOpts.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		createOpts.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	if v, ok := d.GetOk("data_capture_config"); ok {
		createOpts.DataCaptureConfig = expandSagemakerDataCaptureConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker Endpoint Configuration create config: %#v", *createOpts)
	_, err := conn.CreateEndpointConfig(createOpts)
	if err != nil {
		return fmt.Errorf("error creating SageMaker Endpoint Configuration: %s", err)
	}
	d.SetId(name)

	return resourceAwsSagemakerEndpointConfigurationRead(d, meta)
}

func resourceAwsSagemakerEndpointConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	request := &sagemaker.DescribeEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}

	endpointConfig, err := conn.DescribeEndpointConfig(request)
	if err != nil {
		if isAWSErr(err, "ValidationException", "Could not find endpoint configuration") {
			log.Printf("[INFO] unable to find the SageMaker Endpoint Configuration resource and therefore it is removed from the state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading SageMaker Endpoint Configuration %s: %s", d.Id(), err)
	}

	if err := d.Set("arn", endpointConfig.EndpointConfigArn); err != nil {
		return err
	}
	if err := d.Set("name", endpointConfig.EndpointConfigName); err != nil {
		return err
	}
	if err := d.Set("production_variants", flattenProductionVariants(endpointConfig.ProductionVariants)); err != nil {
		return err
	}
	if err := d.Set("kms_key_arn", endpointConfig.KmsKeyId); err != nil {
		return err
	}
	if err := d.Set("data_capture_config", flattenSagemakerDataCaptureConfig(endpointConfig.DataCaptureConfig)); err != nil {
		return err
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, aws.StringValue(endpointConfig.EndpointConfigArn))
	if err != nil {
		return fmt.Errorf("error listing tags for Sagemaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSagemakerEndpointConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Sagemaker Endpoint Configuration (%s) tags: %s", d.Id(), err)
		}
	}
	return resourceAwsSagemakerEndpointConfigurationRead(d, meta)
}

func resourceAwsSagemakerEndpointConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	deleteOpts := &sagemaker.DeleteEndpointConfigInput{
		EndpointConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Endpoint Configuration: %s", d.Id())

	_, err := conn.DeleteEndpointConfig(deleteOpts)

	if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SageMaker Endpoint Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

func expandSagemakerProductionVariants(configured []interface{}) []*sagemaker.ProductionVariant {
	containers := make([]*sagemaker.ProductionVariant, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.ProductionVariant{
			InstanceType:         aws.String(data["instance_type"].(string)),
			ModelName:            aws.String(data["model_name"].(string)),
			InitialInstanceCount: aws.Int64(int64(data["initial_instance_count"].(int))),
		}

		if v, ok := data["variant_name"]; ok {
			l.VariantName = aws.String(v.(string))
		} else {
			l.VariantName = aws.String(resource.UniqueId())
		}

		if v, ok := data["initial_variant_weight"]; ok {
			l.InitialVariantWeight = aws.Float64(v.(float64))
		}

		if v, ok := data["accelerator_type"]; ok && v.(string) != "" {
			l.AcceleratorType = aws.String(data["accelerator_type"].(string))
		}

		containers = append(containers, l)
	}

	return containers
}

func flattenProductionVariants(list []*sagemaker.ProductionVariant) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))

	for _, i := range list {
		l := map[string]interface{}{
			"accelerator_type":       aws.StringValue(i.AcceleratorType),
			"initial_instance_count": aws.Int64Value(i.InitialInstanceCount),
			"initial_variant_weight": aws.Float64Value(i.InitialVariantWeight),
			"instance_type":          aws.StringValue(i.InstanceType),
			"model_name":             aws.StringValue(i.ModelName),
			"variant_name":           aws.StringValue(i.VariantName),
		}

		result = append(result, l)
	}
	return result
}

func expandSagemakerDataCaptureConfig(configured []interface{}) *sagemaker.DataCaptureConfig {
	if len(configured) == 0 {
		return nil
	}

	m := configured[0].(map[string]interface{})

	c := &sagemaker.DataCaptureConfig{
		InitialSamplingPercentage: aws.Int64(int64(m["initial_sampling_percentage"].(int))),
		DestinationS3Uri:          aws.String(m["destination_s3_uri"].(string)),
		CaptureOptions:            expandSagemakerCaptureOptions(m["capture_options"].([]interface{})),
	}

	if v, ok := m["enable_capture"]; ok {
		c.EnableCapture = aws.Bool(v.(bool))
	}

	if v, ok := m["kms_key_id"]; ok && v.(string) != "" {
		c.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := m["capture_content_type_header"]; ok && (len(v.([]interface{})) > 0) {
		c.CaptureContentTypeHeader = expandSagemakerCaptureContentTypeHeader(v.([]interface{})[0].(map[string]interface{}))
	}

	return c
}

func flattenSagemakerDataCaptureConfig(dataCaptureConfig *sagemaker.DataCaptureConfig) []map[string]interface{} {
	if dataCaptureConfig == nil {
		return []map[string]interface{}{}
	}

	cfg := map[string]interface{}{
		"initial_sampling_percentage": aws.Int64Value(dataCaptureConfig.InitialSamplingPercentage),
		"destination_s3_uri":          aws.StringValue(dataCaptureConfig.DestinationS3Uri),
		"capture_options":             flattenSagemakerCaptureOptions(dataCaptureConfig.CaptureOptions),
	}

	if dataCaptureConfig.EnableCapture != nil {
		cfg["enable_capture"] = aws.BoolValue(dataCaptureConfig.EnableCapture)
	}

	if dataCaptureConfig.KmsKeyId != nil {
		cfg["kms_key_id"] = aws.StringValue(dataCaptureConfig.KmsKeyId)
	}

	if dataCaptureConfig.CaptureContentTypeHeader != nil {
		cfg["capture_content_type_header"] = flattenSagemakerCaptureContentTypeHeader(dataCaptureConfig.CaptureContentTypeHeader)
	}

	return []map[string]interface{}{cfg}
}

func expandSagemakerCaptureOptions(configured []interface{}) []*sagemaker.CaptureOption {
	containers := make([]*sagemaker.CaptureOption, 0, len(configured))

	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &sagemaker.CaptureOption{
			CaptureMode: aws.String(data["capture_mode"].(string)),
		}
		containers = append(containers, l)
	}

	return containers
}

func flattenSagemakerCaptureOptions(list []*sagemaker.CaptureOption) []map[string]interface{} {
	containers := make([]map[string]interface{}, 0, len(list))

	for _, lRaw := range list {
		captureOption := make(map[string]interface{})
		captureOption["capture_mode"] = aws.StringValue(lRaw.CaptureMode)
		containers = append(containers, captureOption)
	}

	return containers
}

func expandSagemakerCaptureContentTypeHeader(m map[string]interface{}) *sagemaker.CaptureContentTypeHeader {
	c := &sagemaker.CaptureContentTypeHeader{}

	if v, ok := m["csv_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.CsvContentTypes = expandStringSet(v)
	}

	if v, ok := m["json_content_types"].(*schema.Set); ok && v.Len() > 0 {
		c.JsonContentTypes = expandStringSet(v)
	}

	return c
}

func flattenSagemakerCaptureContentTypeHeader(contentTypeHeader *sagemaker.CaptureContentTypeHeader) []map[string]interface{} {
	if contentTypeHeader == nil {
		return []map[string]interface{}{}
	}

	l := make(map[string]interface{})

	if contentTypeHeader.CsvContentTypes != nil {
		l["csv_content_types"] = flattenStringSet(contentTypeHeader.CsvContentTypes)
	}

	if contentTypeHeader.JsonContentTypes != nil {
		l["json_content_types"] = flattenStringSet(contentTypeHeader.JsonContentTypes)
	}

	return []map[string]interface{}{l}
}
