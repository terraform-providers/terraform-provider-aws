package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"destination_location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				// Ignore missing configuration block
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"atime": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.AtimeBestEffort,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.AtimeBestEffort,
								datasync.AtimeNone,
							}, false),
						},
						"bytes_per_second": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      -1,
							ValidateFunc: validation.IntAtLeast(-1),
						},
						"gid": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.GidIntValue,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.GidBoth,
								datasync.GidIntValue,
								datasync.GidName,
								datasync.GidNone,
							}, false),
						},
						"mtime": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.MtimePreserve,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.MtimeNone,
								datasync.MtimePreserve,
							}, false),
						},
						"posix_permissions": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.PosixPermissionsPreserve,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.PosixPermissionsNone,
								datasync.PosixPermissionsPreserve,
							}, false),
						},
						"preserve_deleted_files": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.PreserveDeletedFilesPreserve,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.PreserveDeletedFilesPreserve,
								datasync.PreserveDeletedFilesRemove,
							}, false),
						},
						"preserve_devices": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.PreserveDevicesNone,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.PreserveDevicesNone,
								datasync.PreserveDevicesPreserve,
							}, false),
						},
						"uid": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.UidIntValue,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.UidBoth,
								datasync.UidIntValue,
								datasync.UidName,
								datasync.UidNone,
							}, false),
						},
						"verify_mode": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  datasync.VerifyModePointInTimeConsistent,
							ValidateFunc: validation.StringInSlice([]string{
								datasync.VerifyModeNone,
								datasync.VerifyModePointInTimeConsistent,
								datasync.VerifyModeOnlyFilesTransferred,
							}, false),
						},
					},
				},
			},
			"source_location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDataSyncTaskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	input := &datasync.CreateTaskInput{
		DestinationLocationArn: aws.String(d.Get("destination_location_arn").(string)),
		Options:                expandDataSyncOptions(d.Get("options").([]interface{})),
		SourceLocationArn:      aws.String(d.Get("source_location_arn").(string)),
		Tags:                   keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().DatasyncTags(),
	}

	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok {
		input.CloudWatchLogGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Task: %s", input)
	output, err := conn.CreateTask(input)
	if err != nil {
		return fmt.Errorf("error creating DataSync Task: %s", err)
	}

	d.SetId(aws.StringValue(output.TaskArn))

	// Task creation can take a few minutes\
	taskInput := &datasync.DescribeTaskInput{
		TaskArn: aws.String(d.Id()),
	}
	var taskOutput *datasync.DescribeTaskOutput
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		taskOutput, err := conn.DescribeTask(taskInput)

		if isAWSErr(err, "InvalidRequestException", "not found") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if aws.StringValue(taskOutput.Status) == datasync.TaskStatusAvailable || aws.StringValue(taskOutput.Status) == datasync.TaskStatusRunning {
			return nil
		}

		err = fmt.Errorf("waiting for DataSync Task (%s) creation: last status (%s), error code (%s), error detail: %s",
			d.Id(), aws.StringValue(taskOutput.Status), aws.StringValue(taskOutput.ErrorCode), aws.StringValue(taskOutput.ErrorDetail))

		if aws.StringValue(taskOutput.Status) == datasync.TaskStatusCreating {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err) // should only happen if err != nil
	})
	if isResourceTimeoutError(err) {
		taskOutput, err = conn.DescribeTask(taskInput)
		if isAWSErr(err, "InvalidRequestException", "not found") {
			return fmt.Errorf("Task not found after creation: %s", err)
		}
		if err != nil {
			return fmt.Errorf("Error describing task after creation: %s", err)
		}
		if aws.StringValue(taskOutput.Status) == datasync.TaskStatusCreating {
			return fmt.Errorf("Data sync task status has not finished creating")
		}
	}
	if err != nil {
		return fmt.Errorf("error waiting for DataSync Task (%s) creation: %s", d.Id(), err)
	}

	return resourceAwsDataSyncTaskRead(d, meta)
}

func resourceAwsDataSyncTaskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &datasync.DescribeTaskInput{
		TaskArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Task: %s", input)
	output, err := conn.DescribeTask(input)

	if isAWSErr(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Task %q not found - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Task (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.TaskArn)
	d.Set("cloudwatch_log_group_arn", output.CloudWatchLogGroupArn)
	d.Set("destination_location_arn", output.DestinationLocationArn)
	d.Set("name", output.Name)

	if err := d.Set("options", flattenDataSyncOptions(output.Options)); err != nil {
		return fmt.Errorf("error setting options: %s", err)
	}

	d.Set("source_location_arn", output.SourceLocationArn)

	tags, err := keyvaluetags.DatasyncListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Task (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsDataSyncTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).datasyncconn

	if d.HasChanges("options", "name") {
		input := &datasync.UpdateTaskInput{
			Options: expandDataSyncOptions(d.Get("options").([]interface{})),
			Name:    aws.String(d.Get("name").(string)),
			TaskArn: aws.String(d.Id()),
		}

		log.Printf("[DEBUG] Updating DataSync Task: %s", input)
		if _, err := conn.UpdateTask(input); err != nil {
			return fmt.Errorf("error creating DataSync Task: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.DatasyncUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Task (%s) tags: %s", d.Id(), err)
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

	if isAWSErr(err, "InvalidRequestException", "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Task (%s): %s", d.Id(), err)
	}

	return nil
}
