package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/datasync/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/datasync/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsDataSyncTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDataSyncTaskCreate,
		Read:   resourceAwsDataSyncTaskRead,
		Update: resourceAwsDataSyncTaskUpdate,
		Delete: resourceAwsDataSyncTaskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_log_group_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"destination_location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"excludes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(datasync.FilterType_Values(), false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"atime": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.AtimeBestEffort,
							ValidateFunc: validation.StringInSlice(datasync.Atime_Values(), false),
						},
						"bytes_per_second": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      -1,
							ValidateFunc: validation.IntAtLeast(-1),
						},
						"gid": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.GidIntValue,
							ValidateFunc: validation.StringInSlice(datasync.Gid_Values(), false),
						},
						"log_level": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.LogLevelOff,
							ValidateFunc: validation.StringInSlice(datasync.LogLevel_Values(), false),
						},
						"mtime": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.MtimePreserve,
							ValidateFunc: validation.StringInSlice(datasync.Mtime_Values(), false),
						},
						"overwrite_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.OverwriteModeAlways,
							ValidateFunc: validation.StringInSlice(datasync.OverwriteMode_Values(), false),
						},
						"posix_permissions": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.PosixPermissionsPreserve,
							ValidateFunc: validation.StringInSlice(datasync.PosixPermissions_Values(), false),
						},
						"preserve_deleted_files": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.PreserveDeletedFilesPreserve,
							ValidateFunc: validation.StringInSlice(datasync.PreserveDeletedFiles_Values(), false),
						},
						"preserve_devices": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.PreserveDevicesNone,
							ValidateFunc: validation.StringInSlice(datasync.PreserveDevices_Values(), false),
						},
						"task_queueing": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.TaskQueueingEnabled,
							ValidateFunc: validation.StringInSlice(datasync.TaskQueueing_Values(), false),
						},
						"transfer_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.TransferModeChanged,
							ValidateFunc: validation.StringInSlice(datasync.TransferMode_Values(), false),
						},
						"uid": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.UidIntValue,
							ValidateFunc: validation.StringInSlice(datasync.Uid_Values(), false),
						},
						"verify_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.VerifyModePointInTimeConsistent,
							ValidateFunc: validation.StringInSlice(datasync.VerifyMode_Values(), false),
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule_expression": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\ \_\*\?\,\|\^\-\/\#\s\(\)\+]*$`),
									"Schedule expressions must have the following syntax: rate(<number>\\\\s?(minutes?|hours?|days?)), cron(<cron_expression>) or at(yyyy-MM-dd'T'HH:mm:ss)."),
							),
						},
					},
				},
			},
			"source_location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsDataSyncTaskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateTaskInput{
		DestinationLocationArn: aws.String(d.Get("destination_location_arn").(string)),
		Options:                expandDataSyncOptions(d.Get("options").([]interface{})),
		SourceLocationArn:      aws.String(d.Get("source_location_arn").(string)),
		Tags:                   tags.IgnoreAws().DatasyncTags(),
	}

	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok {
		input.CloudWatchLogGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excludes"); ok {
		input.Excludes = expandAwsDataSyncFilterRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = expandAwsDataSyncTaskSchedule(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating DataSync Task: %s", input)
	output, err := conn.CreateTask(input)

	if err != nil {
		return fmt.Errorf("error creating DataSync Task: %w", err)
	}

	d.SetId(aws.StringValue(output.TaskArn))

	if _, err := waiter.TaskAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for DataSync Task (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsDataSyncTaskRead(d, meta)
}

func resourceAwsDataSyncTaskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	output, err := finder.TaskByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Task (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Task (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.TaskArn)
	d.Set("cloudwatch_log_group_arn", output.CloudWatchLogGroupArn)
	d.Set("destination_location_arn", output.DestinationLocationArn)
	if err := d.Set("excludes", flattenAwsDataSyncFilterRules(output.Excludes)); err != nil {
		return fmt.Errorf("error setting excludes: %w", err)
	}
	d.Set("name", output.Name)
	if err := d.Set("options", flattenDataSyncOptions(output.Options)); err != nil {
		return fmt.Errorf("error setting options: %w", err)
	}
	if err := d.Set("schedule", flattenAwsDataSyncTaskSchedule(output.Schedule)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}
	d.Set("source_location_arn", output.SourceLocationArn)

	tags, err := keyvaluetags.DatasyncListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Task (%s): %w", d.Id(), err)
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

func resourceAwsDataSyncTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &datasync.UpdateTaskInput{
			TaskArn: aws.String(d.Id()),
		}

		if d.HasChanges("cloudwatch_log_group_arn") {
			input.CloudWatchLogGroupArn = aws.String(d.Get("cloudwatch_log_group_arn").(string))
		}

		if d.HasChanges("excludes") {
			input.Excludes = expandAwsDataSyncFilterRules(d.Get("excludes").([]interface{}))
		}

		if d.HasChanges("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChanges("options") {
			input.Options = expandDataSyncOptions(d.Get("options").([]interface{}))
		}

		if d.HasChanges("schedule") {
			input.Schedule = expandAwsDataSyncTaskSchedule(d.Get("schedule").([]interface{}))
		}

		log.Printf("[DEBUG] Updating DataSync Task: %s", input)
		if _, err := conn.UpdateTask(input); err != nil {
			return fmt.Errorf("error updating DataSync Task (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Task (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsDataSyncTaskRead(d, meta)
}

func resourceAwsDataSyncTaskDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.DeleteTaskInput{
		TaskArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Task: %s", input)
	_, err := conn.DeleteTask(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Task (%s): %w", d.Id(), err)
	}

	return nil
}

func expandAwsDataSyncTaskSchedule(l []interface{}) *datasync.TaskSchedule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	schedule := &datasync.TaskSchedule{
		ScheduleExpression: aws.String(m["schedule_expression"].(string)),
	}

	return schedule
}

func flattenAwsDataSyncTaskSchedule(schedule *datasync.TaskSchedule) []interface{} {
	if schedule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"schedule_expression": aws.StringValue(schedule.ScheduleExpression),
	}

	return []interface{}{m}
}

func expandAwsDataSyncFilterRules(l []interface{}) []*datasync.FilterRule {
	filterRules := []*datasync.FilterRule{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filterRule := &datasync.FilterRule{
			FilterType: aws.String(m["filter_type"].(string)),
			Value:      aws.String(m["value"].(string)),
		}
		filterRules = append(filterRules, filterRule)
	}

	return filterRules
}

func flattenAwsDataSyncFilterRules(filterRules []*datasync.FilterRule) []interface{} {
	l := []interface{}{}

	for _, filterRule := range filterRules {
		m := map[string]interface{}{
			"filter_type": aws.StringValue(filterRule.FilterType),
			"value":       aws.StringValue(filterRule.Value),
		}
		l = append(l, m)
	}

	return l
}
