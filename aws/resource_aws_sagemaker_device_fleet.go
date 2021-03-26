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
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerDeviceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("device_fleet_name").(string)
	input := &sagemaker.CreateDeviceFleetInput{
		DeviceFleetName: aws.String(name),
		OutputConfig:    expandSagemakerFeatureDeviceFleetOutputConfig(d.Get("output_config").([]interface{})),
	}

	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
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
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	deviceFleet, err := finder.DeviceFleetByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "does not exist") {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker DeviceFleet (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker DeviceFleet (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(deviceFleet.DeviceFleetArn)
	d.Set("deviceFleet_name", deviceFleet.DeviceFleetName)
	d.Set("arn", arn)
	d.Set("role_arn", deviceFleet.RoleArn)
	d.Set("description", deviceFleet.Description)
	d.Set("iot_role_alias", deviceFleet.IotRoleAlias)

	if err := d.Set("output_config", flattenSagemakerFeatureDeviceFleetOutputConfig(deviceFleet.OutputConfig)); err != nil {
		return fmt.Errorf("error setting output_config for Sagemaker Device Fleet (%s): %w", d.Id(), err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for SageMaker DeviceFleet (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSagemakerDeviceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChangeExcept("tags") {

		input := &sagemaker.UpdateDeviceFleetInput{
			DeviceFleetName: aws.String(d.Id()),
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("role_arn") {
			input.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("output_config") {
			input.OutputConfig = expandSagemakerFeatureDeviceFleetOutputConfig(d.Get("output_config").([]interface{}))
		}

		log.Printf("[DEBUG] sagemaker DeviceFleet update config: %#v", *input)
		_, err := conn.UpdateDeviceFleet(input)
		if err != nil {
			return fmt.Errorf("error updating SageMaker Device Fleet: %w", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Device Fleet (%s) tags: %s", d.Id(), err)
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
		if isAWSErr(err, sagemaker.ErrCodeResourceNotFound, "No Device Fleet with the name") {
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
