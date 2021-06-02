package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mwaa"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/mwaa/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/mwaa/waiter"
)

func resourceAwsMwaaEnvironment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMwaaEnvironmentCreate,
		Read:   resourceAwsMwaaEnvironmentRead,
		Update: resourceAwsMwaaEnvironmentUpdate,
		Delete: resourceAwsMwaaEnvironmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"airflow_configuration_options": {
				Type:      schema.TypeMap,
				Optional:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"airflow_version": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dag_s3_path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment_class": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"kms_key": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"last_updated": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"error": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"error_code": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"error_message": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"logging_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dag_processing_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     mwaaEnvironmentModuleLoggingConfigurationSchema(),
						},
						"scheduler_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     mwaaEnvironmentModuleLoggingConfigurationSchema(),
						},
						"task_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     mwaaEnvironmentModuleLoggingConfigurationSchema(),
						},
						"webserver_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     mwaaEnvironmentModuleLoggingConfigurationSchema(),
						},
						"worker_logs": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem:     mwaaEnvironmentModuleLoggingConfigurationSchema(),
						},
					},
				},
			},
			"max_workers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"min_workers": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							MinItems: 1,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MinItems: 2,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"plugins_s3_object_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"plugins_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"requirements_s3_object_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"requirements_s3_path": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"service_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_bucket_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"webserver_access_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(mwaa.WebserverAccessMode_Values(), false),
			},
			"webserver_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"weekly_maintenance_window_start": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsMwaaEnvironmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mwaaconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := mwaa.CreateEnvironmentInput{
		DagS3Path:            aws.String(d.Get("dag_s3_path").(string)),
		ExecutionRoleArn:     aws.String(d.Get("execution_role_arn").(string)),
		Name:                 aws.String(d.Get("name").(string)),
		NetworkConfiguration: expandMwaaEnvironmentNetworkConfigurationCreate(d.Get("network_configuration").([]interface{})),
		SourceBucketArn:      aws.String(d.Get("source_bucket_arn").(string)),
	}

	if v, ok := d.GetOk("airflow_configuration_options"); ok {
		input.AirflowConfigurationOptions = expandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("airflow_version"); ok {
		input.AirflowVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("environment_class"); ok {
		input.EnvironmentClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_key"); ok {
		input.KmsKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("logging_configuration"); ok {
		input.LoggingConfiguration = expandMwaaEnvironmentLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("max_workers"); ok {
		input.MaxWorkers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("min_workers"); ok {
		input.MinWorkers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("plugins_s3_object_version"); ok {
		input.PluginsS3ObjectVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("plugins_s3_path"); ok {
		input.PluginsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("requirements_s3_object_version"); ok {
		input.RequirementsS3ObjectVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("requirements_s3_path"); ok {
		input.RequirementsS3Path = aws.String(v.(string))
	}

	if v, ok := d.GetOk("webserver_access_mode"); ok {
		input.WebserverAccessMode = aws.String(v.(string))
	}

	if v, ok := d.GetOk("weekly_maintenance_window_start"); ok {
		input.WeeklyMaintenanceWindowStart = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().MwaaTags()
	}

	log.Printf("[INFO] Creating MWAA Environment: %s", input)
	_, err := conn.CreateEnvironment(&input)
	if err != nil {
		return fmt.Errorf("error creating MWAA Environment: %w", err)
	}

	d.SetId(aws.StringValue(input.Name))

	if _, err := waiter.EnvironmentCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for MWAA Environment (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsMwaaEnvironmentRead(d, meta)
}

func resourceAwsMwaaEnvironmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mwaaconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[INFO] Reading MWAA Environment: %s", d.Id())

	environment, err := finder.EnvironmentByName(conn, d.Id())

	if err != nil {
		if isAWSErr(err, mwaa.ErrCodeResourceNotFoundException, "") && !d.IsNewResource() {
			log.Printf("[WARN] MWAA Environment %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading MWAA Environment (%s): %w", d.Id(), err)
	}

	if environment == nil {
		return fmt.Errorf("error reading MWAA Environment (%s): empty response", d.Id())
	}

	d.Set("airflow_configuration_options", aws.StringValueMap(environment.AirflowConfigurationOptions))
	d.Set("airflow_version", environment.AirflowVersion)
	d.Set("arn", environment.Arn)
	d.Set("created_at", aws.TimeValue(environment.CreatedAt).String())
	d.Set("dag_s3_path", environment.DagS3Path)
	d.Set("environment_class", environment.EnvironmentClass)
	d.Set("execution_role_arn", environment.ExecutionRoleArn)
	d.Set("kms_key", environment.KmsKey)
	if err := d.Set("last_updated", flattenMwaaLastUpdate(environment.LastUpdate)); err != nil {
		return fmt.Errorf("error reading MWAA Environment (%s): %w", d.Id(), err)
	}
	if err := d.Set("logging_configuration", flattenMwaaLoggingConfiguration(environment.LoggingConfiguration)); err != nil {
		return fmt.Errorf("error reading MWAA Environment (%s): %w", d.Id(), err)
	}
	d.Set("max_workers", environment.MaxWorkers)
	d.Set("min_workers", environment.MinWorkers)
	d.Set("name", environment.Name)
	if err := d.Set("network_configuration", flattenMwaaNetworkConfiguration(environment.NetworkConfiguration)); err != nil {
		return fmt.Errorf("error reading MWAA Environment (%s): %w", d.Id(), err)
	}
	d.Set("plugins_s3_object_version", environment.PluginsS3ObjectVersion)
	d.Set("plugins_s3_path", environment.PluginsS3Path)
	d.Set("requirements_s3_object_version", environment.RequirementsS3ObjectVersion)
	d.Set("requirements_s3_path", environment.RequirementsS3Path)
	d.Set("service_role_arn", environment.ServiceRoleArn)
	d.Set("source_bucket_arn", environment.SourceBucketArn)
	d.Set("status", environment.Status)
	d.Set("webserver_access_mode", environment.WebserverAccessMode)
	d.Set("webserver_url", environment.WebserverUrl)
	d.Set("weekly_maintenance_window_start", environment.WeeklyMaintenanceWindowStart)

	tags := keyvaluetags.MwaaKeyValueTags(environment.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAwsMwaaEnvironmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mwaaconn

	input := mwaa.UpdateEnvironmentInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if d.HasChangesExcept("tags", "tags_all") {
		if d.HasChange("airflow_configuration_options") {
			options, ok := d.GetOk("airflow_configuration_options")
			if !ok {
				options = map[string]interface{}{}
			}

			input.AirflowConfigurationOptions = expandStringMap(options.(map[string]interface{}))
		}

		if d.HasChange("airflow_version") {
			input.AirflowVersion = aws.String(d.Get("airflow_version").(string))
		}

		if d.HasChange("dag_s3_path") {
			input.DagS3Path = aws.String(d.Get("dag_s3_path").(string))
		}

		if d.HasChange("environment_class") {
			input.EnvironmentClass = aws.String(d.Get("environment_class").(string))
		}

		if d.HasChange("execution_role_arn") {
			input.ExecutionRoleArn = aws.String(d.Get("execution_role_arn").(string))
		}

		if d.HasChange("logging_configuration") {
			input.LoggingConfiguration = expandMwaaEnvironmentLoggingConfiguration(d.Get("logging_configuration").([]interface{}))
		}

		if d.HasChange("max_workers") {
			input.MaxWorkers = aws.Int64(int64(d.Get("max_workers").(int)))
		}

		if d.HasChange("min_workers") {
			input.MinWorkers = aws.Int64(int64(d.Get("min_workers").(int)))
		}

		if d.HasChange("network_configuration") {
			input.NetworkConfiguration = expandMwaaEnvironmentNetworkConfigurationUpdate(d.Get("network_configuration").([]interface{}))
		}

		if d.HasChange("plugins_s3_object_version") {
			input.PluginsS3ObjectVersion = aws.String(d.Get("plugins_s3_object_version").(string))
		}

		if d.HasChange("plugins_s3_path") {
			input.PluginsS3Path = aws.String(d.Get("plugins_s3_path").(string))
		}

		if d.HasChange("requirements_s3_object_version") {
			input.RequirementsS3ObjectVersion = aws.String(d.Get("requirements_s3_object_version").(string))
		}

		if d.HasChange("requirements_s3_path") {
			input.RequirementsS3Path = aws.String(d.Get("requirements_s3_path").(string))
		}

		if d.HasChange("source_bucket_arn") {
			input.SourceBucketArn = aws.String(d.Get("source_bucket_arn").(string))
		}

		if d.HasChange("webserver_access_mode") {
			input.WebserverAccessMode = aws.String(d.Get("webserver_access_mode").(string))
		}

		if d.HasChange("weekly_maintenance_window_start") {
			input.WeeklyMaintenanceWindowStart = aws.String(d.Get("weekly_maintenance_window_start").(string))
		}

		log.Printf("[INFO] Updating MWAA Environment: %s", input)
		_, err := conn.UpdateEnvironment(&input)

		if err != nil {
			return fmt.Errorf("error updating MWAA Environment (%s): %w", d.Id(), err)
		}

		if _, err := waiter.EnvironmentUpdated(conn, d.Id()); err != nil {
			return fmt.Errorf("error waiting for MWAA Environment (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.MwaaUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating MWAA Environment (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsMwaaEnvironmentRead(d, meta)
}

func resourceAwsMwaaEnvironmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mwaaconn

	log.Printf("[INFO] Deleting MWAA Environment: %s", d.Id())
	_, err := conn.DeleteEnvironment(&mwaa.DeleteEnvironmentInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, mwaa.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return fmt.Errorf("error deleting MWAA Environment (%s): %w", d.Id(), err)
	}

	_, err = waiter.EnvironmentDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for MWAA Environment (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func mwaaEnvironmentModuleLoggingConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cloud_watch_log_group_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(mwaa.LoggingLevel_Values(), false),
			},
		},
	}
}

func expandMwaaEnvironmentLoggingConfiguration(l []interface{}) *mwaa.LoggingConfigurationInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	input := &mwaa.LoggingConfigurationInput{}

	m := l[0].(map[string]interface{})

	if v, ok := m["dag_processing_logs"]; ok {
		input.DagProcessingLogs = expandMwaaEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["scheduler_logs"]; ok {
		input.SchedulerLogs = expandMwaaEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["task_logs"]; ok {
		input.TaskLogs = expandMwaaEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["webserver_logs"]; ok {
		input.WebserverLogs = expandMwaaEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	if v, ok := m["worker_logs"]; ok {
		input.WorkerLogs = expandMwaaEnvironmentModuleLoggingConfiguration(v.([]interface{}))
	}

	return input
}

func expandMwaaEnvironmentModuleLoggingConfiguration(l []interface{}) *mwaa.ModuleLoggingConfigurationInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	input := &mwaa.ModuleLoggingConfigurationInput{}
	m := l[0].(map[string]interface{})

	input.Enabled = aws.Bool(m["enabled"].(bool))
	input.LogLevel = aws.String(m["log_level"].(string))

	return input
}

func expandMwaaEnvironmentNetworkConfigurationCreate(l []interface{}) *mwaa.NetworkConfiguration {
	m := l[0].(map[string]interface{})

	return &mwaa.NetworkConfiguration{
		SecurityGroupIds: expandStringSet(m["security_group_ids"].(*schema.Set)),
		SubnetIds:        expandStringSet(m["subnet_ids"].(*schema.Set)),
	}
}

func expandMwaaEnvironmentNetworkConfigurationUpdate(l []interface{}) *mwaa.UpdateNetworkConfigurationInput {
	m := l[0].(map[string]interface{})

	return &mwaa.UpdateNetworkConfigurationInput{
		SecurityGroupIds: expandStringSet(m["security_group_ids"].(*schema.Set)),
	}
}

func flattenMwaaLastUpdate(lastUpdate *mwaa.LastUpdate) []interface{} {
	if lastUpdate == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if lastUpdate.CreatedAt != nil {
		m["created_at"] = aws.TimeValue(lastUpdate.CreatedAt).String()
	}

	if lastUpdate.Error != nil {
		m["error"] = flattenMwaaLastUpdateError(lastUpdate.Error)
	}

	if lastUpdate.Status != nil {
		m["status"] = lastUpdate.Status
	}

	return []interface{}{m}
}

func flattenMwaaLastUpdateError(error *mwaa.UpdateError) []interface{} {
	if error == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if error.ErrorCode != nil {
		m["error_code"] = error.ErrorCode
	}

	if error.ErrorMessage != nil {
		m["error_message"] = error.ErrorMessage
	}

	return []interface{}{m}
}

func flattenMwaaLoggingConfiguration(loggingConfiguration *mwaa.LoggingConfiguration) []interface{} {
	if loggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if loggingConfiguration.DagProcessingLogs != nil {
		m["dag_processing_logs"] = flattenMwaaModuleLoggingConfiguration(loggingConfiguration.DagProcessingLogs)
	}

	if loggingConfiguration.SchedulerLogs != nil {
		m["scheduler_logs"] = flattenMwaaModuleLoggingConfiguration(loggingConfiguration.SchedulerLogs)
	}

	if loggingConfiguration.TaskLogs != nil {
		m["task_logs"] = flattenMwaaModuleLoggingConfiguration(loggingConfiguration.TaskLogs)
	}

	if loggingConfiguration.WebserverLogs != nil {
		m["webserver_logs"] = flattenMwaaModuleLoggingConfiguration(loggingConfiguration.WebserverLogs)
	}

	if loggingConfiguration.WorkerLogs != nil {
		m["worker_logs"] = flattenMwaaModuleLoggingConfiguration(loggingConfiguration.WorkerLogs)
	}

	return []interface{}{m}
}

func flattenMwaaModuleLoggingConfiguration(moduleLoggingConfiguration *mwaa.ModuleLoggingConfiguration) []interface{} {
	if moduleLoggingConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cloud_watch_log_group_arn": aws.StringValue(moduleLoggingConfiguration.CloudWatchLogGroupArn),
		"enabled":                   aws.BoolValue(moduleLoggingConfiguration.Enabled),
		"log_level":                 aws.StringValue(moduleLoggingConfiguration.LogLevel),
	}

	return []interface{}{m}
}

func flattenMwaaNetworkConfiguration(networkConfiguration *mwaa.NetworkConfiguration) []interface{} {
	if networkConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"security_group_ids": flattenStringSet(networkConfiguration.SecurityGroupIds),
		"subnet_ids":         flattenStringSet(networkConfiguration.SubnetIds),
	}

	return []interface{}{m}
}
