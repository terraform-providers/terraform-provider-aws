package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/waiter"
)

func resourceAwsSagemakerFeatureGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSagemakerFeatureGroupCreate,
		Read:   resourceAwsSagemakerFeatureGroupRead,
		Update: resourceAwsSagemakerFeatureGroupUpdate,
		Delete: resourceAwsSagemakerFeatureGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"feature_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contain alphanumeric character and hyphens. Spaces are not allowed."),
				),
			},
			"record_identifier_feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([-_]*[a-zA-Z0-9]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
				),
			},
			"event_time_feature_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([-_]*[a-zA-Z0-9]){0,63}`),
						"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"feature_definition": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 2500,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"feature_name": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 64),
								validation.StringNotInSlice([]string{"is_deleted", "write_time", "api_invocation_time"}, false),
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]([-_]*[a-zA-Z0-9]){0,63}`),
									"Must start and end with an alphanumeric character and Can only contains alphanumeric characters, hyphens, underscores. Spaces are not allowed."),
							),
						},
						"feature_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.FeatureType_Values(), false),
						},
					},
				},
			},
			"offline_store_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				AtLeastOneOf: []string{"offline_store_config", "online_store_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_catalog_config": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"catalog": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"database": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"table_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"s3_storage_config": {
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
									"s3_uri": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"disable_glue_table_creation": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"online_store_config": {
				Type:         schema.TypeList,
				Optional:     true,
				ForceNew:     true,
				MaxItems:     1,
				AtLeastOneOf: []string{"offline_store_config", "online_store_config"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateArn,
									},
								},
							},
						},
						"enable_online_store": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSagemakerFeatureGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	name := d.Get("feature_group_name").(string)

	input := &sagemaker.CreateFeatureGroupInput{
		FeatureGroupName:            aws.String(name),
		EventTimeFeatureName:        aws.String(d.Get("event_time_feature_name").(string)),
		RecordIdentifierFeatureName: aws.String(d.Get("record_identifier_feature_name").(string)),
		RoleArn:                     aws.String(d.Get("role_arn").(string)),
		FeatureDefinitions:          expandSagemakerFeatureGroupFeatureDefinition(d.Get("feature_definition").([]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		input.Tags = keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().SagemakerTags()
	}

	if v, ok := d.GetOk("offline_store_config"); ok {
		input.OfflineStoreConfig = expandSagemakerFeatureGroupOfflineStoreConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("online_store_config"); ok {
		input.OnlineStoreConfig = expandSagemakerFeatureGroupOnlineStoreConfig(v.([]interface{}))
	}

	log.Printf("[DEBUG] Sagemaker Feature Group create config: %#v", *input)
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.CreateFeatureGroup(input)
		if err != nil {
			if isAWSErr(err, "ValidationException", "The execution role ARN is invalid.") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, "ValidationException", "Invalid S3Uri provided") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.CreateFeatureGroup(input)
	}

	if err != nil {
		return fmt.Errorf("Error creating SageMaker Feature Group: %w", err)
	}

	d.SetId(name)

	if _, err := waiter.FeatureGroupCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for SageMaker Feature Group (%s) to create: %w", d.Id(), err)
	}

	return resourceAwsSagemakerFeatureGroupRead(d, meta)
}

func resourceAwsSagemakerFeatureGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := finder.FeatureGroupByName(conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Feature Group (%s); removing from state", d.Id())
			return nil
		}
		return fmt.Errorf("error reading SageMaker Feature Group (%s): %w", d.Id(), err)

	}

	arn := aws.StringValue(output.FeatureGroupArn)
	d.Set("feature_group_name", output.FeatureGroupName)
	d.Set("event_time_feature_name", output.EventTimeFeatureName)
	d.Set("description", output.Description)
	d.Set("record_identifier_feature_name", output.RecordIdentifierFeatureName)
	d.Set("role_arn", output.RoleArn)
	d.Set("arn", arn)

	if err := d.Set("feature_definition", flattenSagemakerFeatureGroupFeatureDefinition(output.FeatureDefinitions)); err != nil {
		return fmt.Errorf("error setting feature_definition for Sagemaker Feature Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("online_store_config", flattenSagemakerFeatureGroupOnlineStoreConfig(output.OnlineStoreConfig)); err != nil {
		return fmt.Errorf("error setting online_store_config for Sagemaker Feature Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("offline_store_config", flattenSagemakerFeatureGroupOfflineStoreConfig(output.OfflineStoreConfig)); err != nil {
		return fmt.Errorf("error setting offline_store_config for Sagemaker Feature Group (%s): %w", d.Id(), err)
	}

	tags, err := keyvaluetags.SagemakerListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Sagemaker Feature Group (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSagemakerFeatureGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SagemakerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating SageMaker Feature Group (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSagemakerFeatureGroupRead(d, meta)
}

func resourceAwsSagemakerFeatureGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sagemakerconn

	input := &sagemaker.DeleteFeatureGroupInput{
		FeatureGroupName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteFeatureGroup(input); err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return nil
		}
		return fmt.Errorf("error deleting SageMaker Feature Group (%s): %w", d.Id(), err)
	}

	if _, err := waiter.FeatureGroupDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return nil
		}
		return fmt.Errorf("error waiting for SageMaker Feature Group (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

func expandSagemakerFeatureGroupFeatureDefinition(l []interface{}) []*sagemaker.FeatureDefinition {
	featureDefs := make([]*sagemaker.FeatureDefinition, 0, len(l))

	for _, lRaw := range l {
		data := lRaw.(map[string]interface{})

		featureDef := &sagemaker.FeatureDefinition{
			FeatureName: aws.String(data["feature_name"].(string)),
			FeatureType: aws.String(data["feature_type"].(string)),
		}

		featureDefs = append(featureDefs, featureDef)
	}

	return featureDefs
}

func flattenSagemakerFeatureGroupFeatureDefinition(config []*sagemaker.FeatureDefinition) []map[string]interface{} {
	features := make([]map[string]interface{}, 0, len(config))

	for _, i := range config {
		feature := map[string]interface{}{
			"feature_name": aws.StringValue(i.FeatureName),
			"feature_type": aws.StringValue(i.FeatureType),
		}

		features = append(features, feature)
	}
	return features
}

func expandSagemakerFeatureGroupOnlineStoreConfig(l []interface{}) *sagemaker.OnlineStoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OnlineStoreConfig{
		EnableOnlineStore: aws.Bool(m["enable_online_store"].(bool)),
	}

	if v, ok := m["security_config"].([]interface{}); ok && len(v) > 0 {
		config.SecurityConfig = expandSagemakerFeatureGroupOnlineStoreConfigSecurityConfig(v)
	}

	return config
}

func flattenSagemakerFeatureGroupOnlineStoreConfig(config *sagemaker.OnlineStoreConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"enable_online_store": aws.BoolValue(config.EnableOnlineStore),
	}

	if config.SecurityConfig != nil {
		m["security_config"] = flattenSagemakerFeatureGroupOnlineStoreConfigSecurityConfig(config.SecurityConfig)
	}

	return []map[string]interface{}{m}
}

func expandSagemakerFeatureGroupOnlineStoreConfigSecurityConfig(l []interface{}) *sagemaker.OnlineStoreSecurityConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OnlineStoreSecurityConfig{
		KmsKeyId: aws.String(m["kms_key_id"].(string)),
	}

	return config
}

func flattenSagemakerFeatureGroupOnlineStoreConfigSecurityConfig(config *sagemaker.OnlineStoreSecurityConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"kms_key_id": aws.StringValue(config.KmsKeyId),
	}

	return []map[string]interface{}{m}
}

func expandSagemakerFeatureGroupOfflineStoreConfig(l []interface{}) *sagemaker.OfflineStoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.OfflineStoreConfig{}

	if v, ok := m["s3_storage_config"].([]interface{}); ok && len(v) > 0 {
		config.S3StorageConfig = expandSagemakerFeatureGroupOfflineStoreConfigS3StorageConfig(v)
	}

	if v, ok := m["data_catalog_config"].([]interface{}); ok && len(v) > 0 {
		config.DataCatalogConfig = expandSagemakerFeatureGroupOfflineStoreConfigDataCatalogConfig(v)
	}

	if v, ok := m["disable_glue_table_creation"].(bool); ok {
		config.DisableGlueTableCreation = aws.Bool(v)
	}

	return config
}

func flattenSagemakerFeatureGroupOfflineStoreConfig(config *sagemaker.OfflineStoreConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"disable_glue_table_creation": aws.BoolValue(config.DisableGlueTableCreation),
	}

	if config.DataCatalogConfig != nil {
		m["data_catalog_config"] = flattenSagemakerFeatureGroupOfflineStoreConfigDataCatalogConfig(config.DataCatalogConfig)
	}

	if config.S3StorageConfig != nil {
		m["s3_storage_config"] = flattenSagemakerFeatureGroupOfflineStoreConfigS3StorageConfig(config.S3StorageConfig)
	}

	return []map[string]interface{}{m}
}

func expandSagemakerFeatureGroupOfflineStoreConfigS3StorageConfig(l []interface{}) *sagemaker.S3StorageConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.S3StorageConfig{
		S3Uri: aws.String(m["s3_uri"].(string)),
	}

	if v, ok := m["kms_key_id"].(string); ok && v != "" {
		config.KmsKeyId = aws.String(m["kms_key_id"].(string))
	}

	return config
}

func flattenSagemakerFeatureGroupOfflineStoreConfigS3StorageConfig(config *sagemaker.S3StorageConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"s3_uri": aws.StringValue(config.S3Uri),
	}

	if config.KmsKeyId != nil {
		m["kms_key_id"] = aws.StringValue(config.KmsKeyId)
	}

	return []map[string]interface{}{m}
}

func expandSagemakerFeatureGroupOfflineStoreConfigDataCatalogConfig(l []interface{}) *sagemaker.DataCatalogConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.DataCatalogConfig{
		Catalog:   aws.String(m["catalog"].(string)),
		Database:  aws.String(m["database"].(string)),
		TableName: aws.String(m["table_name"].(string)),
	}

	return config
}

func flattenSagemakerFeatureGroupOfflineStoreConfigDataCatalogConfig(config *sagemaker.DataCatalogConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"catalog":    aws.StringValue(config.Catalog),
		"database":   aws.StringValue(config.Database),
		"table_name": aws.StringValue(config.TableName),
	}

	return []map[string]interface{}{m}
}
