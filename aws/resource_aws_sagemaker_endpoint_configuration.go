package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

						"initial_variant_weight": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.FloatAtLeast(0),
							Default:      1,
						},

						"accelerator_type": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
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
		if isAWSErr(err, "ValidationException", "") {
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
