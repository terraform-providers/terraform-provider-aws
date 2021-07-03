package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func resourceAwsSagemakerDeviceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerDeviceFleetCreate,
		Read:   resourceAwsSagemakerDeviceFleetRead,
		Update: resourceAwsSagemakerDeviceFleetUpdate,
		Delete: resourceAwsSagemakerDeviceFleetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 800),
			},
			"device_fleet_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"enable_iot_role_alias": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"iot_role_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
						"s3_output_location": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsSagemakerDeviceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("device_fleet_name").(string)
	input := &sagemaker.CreateDeviceFleetInput{
		DeviceFleetName:    aws.String(name),
		OutputConfig:       expandSagemakerFeatureDeviceFleetOutputConfig(d.Get("output_config").([]interface{})),
		EnableIotRoleAlias: aws.Bool(d.Get("enable_iot_role_alias").(bool)),
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().SagemakerTags()
	}

	_, err := retryOnAwsCode("ValidationException", func() (interface{}, error) {
		return conn.CreateDeviceFleet(input)
	})
	if err != nil {
		return fmt.Errorf("error creating SageMaker Device Fleet %s: %w", name, err)
	}

	d.SetId(name)

	return resourceAwsSagemakerDeviceFleetRead(d, meta)
}

func resourceAwsSagemakerDeviceFleetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	deviceFleet, err := finder.DeviceFleetByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ValidationException", "No devicefleet with name") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Device Fleet (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker Device Fleet (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(deviceFleet.DeviceFleetArn)
	d.Set("device_fleet_name", deviceFleet.DeviceFleetName)
	d.Set("arn", arn)
	d.Set("role_arn", deviceFleet.RoleArn)
	d.Set("description", deviceFleet.Description)

	iotAlias := aws.StringValue(deviceFleet.IotRoleAlias)
	d.Set("iot_role_alias", iotAlias)
	d.Set("enable_iot_role_alias", len(iotAlias) > 0)

	if err := d.Set("output_config", flattenSagemakerFeatureDeviceFleetOutputConfig(deviceFleet.OutputConfig)); err != nil {
		return fmt.Errorf("error setting output_config for Sagemaker Device Fleet (%s): %w", d.Id(), err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker Device Fleet (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsSagemakerDeviceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChangeExcept("tags_all") {
		input := &sagemaker.UpdateDeviceFleetInput{
			DeviceFleetName:    aws.String(d.Id()),
			EnableIotRoleAlias: aws.Bool(d.Get("enable_iot_role_alias").(bool)),
			OutputConfig:       expandSagemakerFeatureDeviceFleetOutputConfig(d.Get("output_config").([]interface{})),
			RoleArn:            aws.String(d.Get("role_arn").(string)),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		log.Printf("[DEBUG] sagemaker DeviceFleet update config: %#v", *input)
		_, err := conn.UpdateDeviceFleet(input)
		if err != nil {
			return fmt.Errorf("error updating SageMaker Device Fleet: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Device Fleet (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSagemakerDeviceFleetRead(d, meta)
}

func resourceAwsSagemakerDeviceFleetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteDeviceFleetInput{
		DeviceFleetName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteDeviceFleet(input); err != nil {
		if isAWSErr(err, "ValidationException", "DeviceFleet with name") {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Device Fleet (%s): %w", d.Id(), err)
	}

	return nil
}

func expandSagemakerFeatureDeviceFleetOutputConfig(l []interface{}) *sagemaker.EdgeOutputConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.EdgeOutputConfig{
		S3OutputLocation: aws.String(m["s3_output_location"].(string)),
	}

	if v, ok := m["kms_key_id"].(string); ok && v != "" {
		config.KmsKeyId = aws.String(m["kms_key_id"].(string))
	}

	return config
}

func flattenSagemakerFeatureDeviceFleetOutputConfig(config *sagemaker.EdgeOutputConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"s3_output_location": aws.StringValue(config.S3OutputLocation),
	}

	if config.KmsKeyId != nil {
		m["kms_key_id"] = aws.StringValue(config.KmsKeyId)
	}

	return []map[string]interface{}{m}
}
