package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsImageBuilderInfrastructureConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsImageBuilderInfrastructureConfigurationCreate,
		Read:   resourceAwsImageBuilderInfrastructureConfigurationRead,
		Update: resourceAwsImageBuilderInfrastructureConfigurationUpdate,
		Delete: resourceAwsImageBuilderInfrastructureConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"instance_profile_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"key_pair": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"logging": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_logs": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"s3_bucket_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
									},
									"s3_key_prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 1024),
										Default:      "/",
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_tags": tagsSchema(),
			"security_group_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"sns_topic_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"tags": tagsSchema(),
			"terminate_instance_on_failure": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAwsImageBuilderInfrastructureConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.CreateInfrastructureConfigurationInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_profile_name"); ok {
		input.InstanceProfileName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_types"); ok && v.(*schema.Set).Len() > 0 {
		input.InstanceTypes = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("key_pair"); ok {
		input.KeyPair = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Logging = expandImageBuilderLogging(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("resource_tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResourceTags = keyvaluetags.New(v.(map[string]interface{})).ImagebuilderTags()
	}

	if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroupIds = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sns_topic_arn"); ok {
		input.SnsTopicArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		input.SubnetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().ImagebuilderTags()
	}

	if v, ok := d.GetOk("terminate_instance_on_failure"); ok {
		input.TerminateInstanceOnFailure = aws.Bool(v.(bool))
	}

	var output *imagebuilder.CreateInfrastructureConfigurationOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.CreateInfrastructureConfiguration(input)

		if tfawserr.ErrMessageContains(err, imagebuilder.ErrCodeInvalidParameterValueException, "instance profile does not exist") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateInfrastructureConfiguration(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Image Builder Infrastructure Configuration: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error creating Image Builder Infrastructure Configuration: empty response")
	}

	d.SetId(aws.StringValue(output.InfrastructureConfigurationArn))

	return resourceAwsImageBuilderInfrastructureConfigurationRead(d, meta)
}

func resourceAwsImageBuilderInfrastructureConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	output, err := conn.GetInfrastructureConfiguration(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Image Builder Infrastructure Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
	}

	if output == nil || output.InfrastructureConfiguration == nil {
		return fmt.Errorf("error getting Image Builder Infrastructure Configuration (%s): empty response", d.Id())
	}

	infrastructureConfiguration := output.InfrastructureConfiguration

	d.Set("arn", infrastructureConfiguration.Arn)
	d.Set("date_created", infrastructureConfiguration.DateCreated)
	d.Set("date_updated", infrastructureConfiguration.DateUpdated)
	d.Set("description", infrastructureConfiguration.Description)
	d.Set("instance_profile_name", infrastructureConfiguration.InstanceProfileName)
	d.Set("instance_types", aws.StringValueSlice(infrastructureConfiguration.InstanceTypes))
	d.Set("key_pair", infrastructureConfiguration.KeyPair)
	if infrastructureConfiguration.Logging != nil {
		d.Set("logging", []interface{}{flattenImageBuilderLogging(infrastructureConfiguration.Logging)})
	} else {
		d.Set("logging", nil)
	}
	d.Set("name", infrastructureConfiguration.Name)
	d.Set("resource_tags", keyvaluetags.ImagebuilderKeyValueTags(infrastructureConfiguration.ResourceTags).Map())
	d.Set("security_group_ids", aws.StringValueSlice(infrastructureConfiguration.SecurityGroupIds))
	d.Set("sns_topic_arn", infrastructureConfiguration.SnsTopicArn)
	d.Set("subnet_id", infrastructureConfiguration.SubnetId)
	d.Set("tags", keyvaluetags.ImagebuilderKeyValueTags(infrastructureConfiguration.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	d.Set("terminate_instance_on_failure", infrastructureConfiguration.TerminateInstanceOnFailure)

	return nil
}

func resourceAwsImageBuilderInfrastructureConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	if d.HasChanges(
		"description",
		"instance_profile_name",
		"instance_types",
		"key_pair",
		"logging",
		"resource_tags",
		"security_group_ids",
		"sns_topic_arn",
		"subnet_id",
		"terminate_instance_on_failure",
	) {
		input := &imagebuilder.UpdateInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("instance_profile_name"); ok {
			input.InstanceProfileName = aws.String(v.(string))
		}

		if v, ok := d.GetOk("instance_types"); ok && v.(*schema.Set).Len() > 0 {
			input.InstanceTypes = expandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("key_pair"); ok {
			input.KeyPair = aws.String(v.(string))
		}

		if v, ok := d.GetOk("logging"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Logging = expandImageBuilderLogging(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("resource_tags"); ok && len(v.(map[string]interface{})) > 0 {
			input.ResourceTags = keyvaluetags.New(v.(map[string]interface{})).ImagebuilderTags()
		}

		if v, ok := d.GetOk("security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
			input.SecurityGroupIds = expandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("sns_topic_arn"); ok {
			input.SnsTopicArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("subnet_id"); ok {
			input.SubnetId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("terminate_instance_on_failure"); ok {
			input.TerminateInstanceOnFailure = aws.Bool(v.(bool))
		}

		err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
			_, err := conn.UpdateInfrastructureConfiguration(input)

			if tfawserr.ErrMessageContains(err, imagebuilder.ErrCodeInvalidParameterValueException, "instance profile does not exist") {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateInfrastructureConfiguration(input)
		}

		if err != nil {
			return fmt.Errorf("error updating Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.ImagebuilderUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags for Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
		}
	}

	return resourceAwsImageBuilderInfrastructureConfigurationRead(d, meta)
}

func resourceAwsImageBuilderInfrastructureConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).imagebuilderconn

	input := &imagebuilder.DeleteInfrastructureConfigurationInput{
		InfrastructureConfigurationArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteInfrastructureConfiguration(input)

	if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Image Builder Infrastructure Configuration (%s): %w", d.Id(), err)
	}

	return nil
}

func expandImageBuilderLogging(tfMap map[string]interface{}) *imagebuilder.Logging {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.Logging{}

	if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.S3Logs = expandImageBuilderS3Logs(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandImageBuilderS3Logs(tfMap map[string]interface{}) *imagebuilder.S3Logs {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.S3Logs{}

	if v, ok := tfMap["s3_bucket_name"].(string); ok && v != "" {
		apiObject.S3BucketName = aws.String(v)
	}

	if v, ok := tfMap["s3_key_prefix"].(string); ok && v != "" {
		apiObject.S3KeyPrefix = aws.String(v)
	}

	return apiObject
}

func flattenImageBuilderLogging(apiObject *imagebuilder.Logging) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3Logs; v != nil {
		tfMap["s3_logs"] = []interface{}{flattenImageBuilderS3Logs(v)}
	}

	return tfMap
}

func flattenImageBuilderS3Logs(apiObject *imagebuilder.S3Logs) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.S3BucketName; v != nil {
		tfMap["s3_bucket_name"] = aws.StringValue(v)
	}

	if v := apiObject.S3KeyPrefix; v != nil {
		tfMap["s3_key_prefix"] = aws.StringValue(v)
	}

	return tfMap
}
