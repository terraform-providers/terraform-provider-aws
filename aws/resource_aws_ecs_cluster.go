package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ecs/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ecs/waiter"
	iamwaiter "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/iam/waiter"
)

const (
	ecsClusterTimeoutDelete = 10 * time.Minute
	ecsClusterTimeoutUpdate = 10 * time.Minute
)

func resourceAwsEcsCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEcsClusterCreate,
		Read:   resourceAwsEcsClusterRead,
		Update: resourceAwsEcsClusterUpdate,
		Delete: resourceAwsEcsClusterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsEcsClusterImport,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_providers": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"execute_command_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"log_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cloud_watch_encryption_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"cloud_watch_log_group_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"s3_bucket_name": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"s3_bucket_encryption_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"s3_key_prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"logging": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ecs.ExecuteCommandLogging_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"default_capacity_provider_strategy": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 100000),
						},

						"capacity_provider": {
							Type:     schema.TypeString,
							Required: true,
						},

						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},
					},
				},
			},
			"setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ecs.ClusterSettingName_Values(), false),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsEcsClusterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("name", d.Id())
	d.SetId(arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Service:   "ecs",
		Resource:  fmt.Sprintf("cluster/%s", d.Id()),
	}.String())
	return []*schema.ResourceData{d}, nil
}

func resourceAwsEcsClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	clusterName := d.Get("name").(string)
	log.Printf("[DEBUG] Creating ECS cluster %s", clusterName)

	input := &ecs.CreateClusterInput{
		ClusterName:                     aws.String(clusterName),
		DefaultCapacityProviderStrategy: expandEcsCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set)),
		Tags:                            tags.IgnoreAws().EcsTags(),
	}

	if v, ok := d.GetOk("capacity_providers"); ok {
		input.CapacityProviders = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("setting"); ok {
		input.Settings = expandEcsSettings(v.(*schema.Set))
	}

	if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
		input.Configuration = expandECSClusterConfiguration(v.([]interface{}))
	}

	// CreateCluster will create the ECS IAM Service Linked Role on first ECS provision
	// This process does not complete before the initial API call finishes.
	var out *ecs.CreateClusterOutput
	err := resource.Retry(iamwaiter.PropagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateCluster(input)

		if tfawserr.ErrMessageContains(err, ecs.ErrCodeInvalidParameterException, "Unable to assume the service linked role") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		out, err = conn.CreateCluster(input)
	}

	if err != nil {
		return fmt.Errorf("error creating ECS Cluster (%s): %w", clusterName, err)
	}

	log.Printf("[DEBUG] ECS cluster %s created", aws.StringValue(out.Cluster.ClusterArn))

	d.SetId(aws.StringValue(out.Cluster.ClusterArn))

	if _, err := waiter.ClusterAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ECS Cluster (%s) to become Available: %w", d.Id(), err)
	}

	return resourceAwsEcsClusterRead(d, meta)
}

func resourceAwsEcsClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	var out *ecs.DescribeClustersOutput
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		var err error
		out, err = finder.ClusterByARN(conn, d.Id())

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if out == nil || len(out.Failures) > 0 {
			if d.IsNewResource() {
				return resource.RetryableError(&resource.NotFoundError{})
			}
			return resource.NonRetryableError(&resource.NotFoundError{})
		}

		return nil
	})
	if isResourceTimeoutError(err) {
		out, err = finder.ClusterByARN(conn, d.Id())
	}

	if isResourceNotFoundError(err) {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading ECS Cluster (%s): %s", d.Id(), err)
	}

	var cluster *ecs.Cluster
	for _, c := range out.Clusters {
		if aws.StringValue(c.ClusterArn) == d.Id() {
			cluster = c
			break
		}
	}

	if cluster == nil {
		log.Printf("[WARN] ECS Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	// Status==INACTIVE means deleted cluster
	if aws.StringValue(cluster.Status) == "INACTIVE" {
		log.Printf("[WARN] ECS Cluster (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", cluster.ClusterArn)
	d.Set("name", cluster.ClusterName)

	if err := d.Set("capacity_providers", aws.StringValueSlice(cluster.CapacityProviders)); err != nil {
		return fmt.Errorf("error setting capacity_providers: %w", err)
	}
	if err := d.Set("default_capacity_provider_strategy", flattenEcsCapacityProviderStrategy(cluster.DefaultCapacityProviderStrategy)); err != nil {
		return fmt.Errorf("error setting default_capacity_provider_strategy: %w", err)
	}

	if err := d.Set("setting", flattenEcsSettings(cluster.Settings)); err != nil {
		return fmt.Errorf("error setting setting: %w", err)
	}

	if cluster.Configuration != nil {
		if err := d.Set("configuration", flattenECSClusterConfiguration(cluster.Configuration)); err != nil {
			return fmt.Errorf("error setting configuration: %w", err)
		}
	}

	tags := keyvaluetags.EcsKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsEcsClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	if d.HasChanges("setting", "configuration") {
		input := ecs.UpdateClusterInput{
			Cluster: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("setting"); ok {
			input.Settings = expandEcsSettings(v.(*schema.Set))
		}

		if v, ok := d.GetOk("configuration"); ok && len(v.([]interface{})) > 0 {
			input.Configuration = expandECSClusterConfiguration(v.([]interface{}))
		}

		_, err := conn.UpdateCluster(&input)
		if err != nil {
			return fmt.Errorf("error changing ECS cluster (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ClusterAvailable(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for ECS Cluster (%s) to become Available: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.EcsUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating ECS Cluster (%s) tags: %w", d.Id(), err)
		}
	}

	if d.HasChanges("capacity_providers", "default_capacity_provider_strategy") {
		input := ecs.PutClusterCapacityProvidersInput{
			Cluster:                         aws.String(d.Id()),
			CapacityProviders:               expandStringSet(d.Get("capacity_providers").(*schema.Set)),
			DefaultCapacityProviderStrategy: expandEcsCapacityProviderStrategy(d.Get("default_capacity_provider_strategy").(*schema.Set)),
		}

		err := resource.Retry(ecsClusterTimeoutUpdate, func() *resource.RetryError {
			_, err := conn.PutClusterCapacityProviders(&input)
			if err != nil {
				if isAWSErr(err, ecs.ErrCodeClientException, "Cluster was not ACTIVE") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, ecs.ErrCodeResourceInUseException, "") {
					return resource.RetryableError(err)
				}
				if isAWSErr(err, ecs.ErrCodeUpdateInProgressException, "") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.PutClusterCapacityProviders(&input)
		}
		if err != nil {
			return fmt.Errorf("error changing ECS cluster capacity provider settings (%s): %w", d.Id(), err)
		}

		if _, err := waiter.ClusterAvailable(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for ECS Cluster (%s) to become Available: %w", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsEcsClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	log.Printf("[DEBUG] Deleting ECS cluster %s", d.Id())
	input := &ecs.DeleteClusterInput{
		Cluster: aws.String(d.Id()),
	}
	err := resource.Retry(ecsClusterTimeoutDelete, func() *resource.RetryError {
		_, err := conn.DeleteCluster(input)

		if err == nil {
			log.Printf("[DEBUG] ECS cluster %s deleted", d.Id())
			return nil
		}

		if isAWSErr(err, "ClusterContainsContainerInstancesException", "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, "ClusterContainsServicesException", "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		if isAWSErr(err, ecs.ErrCodeUpdateInProgressException, "") {
			log.Printf("[TRACE] Retrying ECS cluster %q deletion after %s", d.Id(), err)
			return resource.RetryableError(err)
		}
		return resource.NonRetryableError(err)
	})
	if isResourceTimeoutError(err) {
		_, err = conn.DeleteCluster(input)
	}
	if err != nil {
		return fmt.Errorf("Error deleting ECS cluster: %s", err)
	}

	if _, err := waiter.ClusterDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for ECS Cluster (%s) to become Deleted: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] ECS cluster %q deleted", d.Id())
	return nil
}

func expandEcsSettings(configured *schema.Set) []*ecs.ClusterSetting {
	list := configured.List()
	if len(list) == 0 {
		return nil
	}

	settings := make([]*ecs.ClusterSetting, 0, len(list))

	for _, raw := range list {
		data := raw.(map[string]interface{})

		setting := &ecs.ClusterSetting{
			Name:  aws.String(data["name"].(string)),
			Value: aws.String(data["value"].(string)),
		}

		settings = append(settings, setting)
	}

	return settings
}

func flattenEcsSettings(list []*ecs.ClusterSetting) []map[string]interface{} {
	if len(list) == 0 {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(list))
	for _, setting := range list {
		l := map[string]interface{}{
			"name":  aws.StringValue(setting.Name),
			"value": aws.StringValue(setting.Value),
		}

		result = append(result, l)
	}
	return result
}

func flattenECSClusterConfiguration(apiObject *ecs.ClusterConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ExecuteCommandConfiguration != nil {
		tfMap["execute_command_configuration"] = flattenECSClusterConfigurationExecuteCommandConfiguration(apiObject.ExecuteCommandConfiguration)
	}
	return []interface{}{tfMap}
}

func flattenECSClusterConfigurationExecuteCommandConfiguration(apiObject *ecs.ExecuteCommandConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.KmsKeyId != nil {
		tfMap["kms_key_id"] = aws.StringValue(apiObject.KmsKeyId)
	}

	if apiObject.LogConfiguration != nil {
		tfMap["log_configuration"] = flattenECSClusterConfigurationExecuteCommandConfigurationLogConfiguration(apiObject.LogConfiguration)
	}

	if apiObject.Logging != nil {
		tfMap["logging"] = aws.StringValue(apiObject.Logging)
	}

	return []interface{}{tfMap}
}

func flattenECSClusterConfigurationExecuteCommandConfigurationLogConfiguration(apiObject *ecs.ExecuteCommandLogConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["cloud_watch_encryption_enabled"] = aws.BoolValue(apiObject.CloudWatchEncryptionEnabled)
	tfMap["s3_bucket_encryption_enabled"] = aws.BoolValue(apiObject.S3EncryptionEnabled)

	if apiObject.CloudWatchLogGroupName != nil {
		tfMap["cloud_watch_log_group_name"] = aws.StringValue(apiObject.CloudWatchLogGroupName)
	}

	if apiObject.S3BucketName != nil {
		tfMap["s3_bucket_name"] = aws.StringValue(apiObject.S3BucketName)
	}

	if apiObject.S3KeyPrefix != nil {
		tfMap["s3_key_prefix"] = aws.StringValue(apiObject.S3KeyPrefix)
	}

	return []interface{}{tfMap}
}

func expandECSClusterConfiguration(nc []interface{}) *ecs.ClusterConfiguration {
	if len(nc) == 0 {
		return &ecs.ClusterConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &ecs.ClusterConfiguration{}
	if v, ok := raw["execute_command_configuration"].([]interface{}); ok && len(v) > 0 {
		config.ExecuteCommandConfiguration = expandECSClusterConfigurationExecuteCommandConfiguration(v)
	}

	return config
}

func expandECSClusterConfigurationExecuteCommandConfiguration(nc []interface{}) *ecs.ExecuteCommandConfiguration {
	if len(nc) == 0 {
		return &ecs.ExecuteCommandConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &ecs.ExecuteCommandConfiguration{}
	if v, ok := raw["log_configuration"].([]interface{}); ok && len(v) > 0 {
		config.LogConfiguration = expandECSClusterConfigurationExecuteCommandLogConfiguration(v)
	}

	if v, ok := raw["kms_key_id"].(string); ok && v != "" {
		config.KmsKeyId = aws.String(v)
	}

	if v, ok := raw["logging"].(string); ok && v != "" {
		config.Logging = aws.String(v)
	}

	return config
}

func expandECSClusterConfigurationExecuteCommandLogConfiguration(nc []interface{}) *ecs.ExecuteCommandLogConfiguration {
	if len(nc) == 0 {
		return &ecs.ExecuteCommandLogConfiguration{}
	}
	raw := nc[0].(map[string]interface{})

	config := &ecs.ExecuteCommandLogConfiguration{}

	if v, ok := raw["cloud_watch_log_group_name"].(string); ok && v != "" {
		config.CloudWatchLogGroupName = aws.String(v)
	}

	if v, ok := raw["s3_bucket_name"].(string); ok && v != "" {
		config.S3BucketName = aws.String(v)
	}

	if v, ok := raw["s3_key_prefix"].(string); ok && v != "" {
		config.S3KeyPrefix = aws.String(v)
	}

	if v, ok := raw["cloud_watch_encryption_enabled"].(bool); ok {
		config.CloudWatchEncryptionEnabled = aws.Bool(v)
	}

	if v, ok := raw["s3_bucket_encryption_enabled"].(bool); ok {
		config.S3EncryptionEnabled = aws.Bool(v)
	}

	return config
}
