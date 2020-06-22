package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGlueJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueJobCreate,
		Read:   resourceAwsGlueJobRead,
		Update: resourceAwsGlueJobUpdate,
		Delete: resourceAwsGlueJobDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"allocated_capacity": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"max_capacity", "number_of_workers", "worker_type"},
				Deprecated:    "Please use attribute `max_capacity' instead. This attribute might be removed in future releases.",
				ValidateFunc:  validation.IntAtLeast(2),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"command": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "glueetl",
						},
						"script_location": {
							Type:     schema.TypeString,
							Required: true,
						},
						"python_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"2", "3"}, true),
						},
					},
				},
			},
			"connections": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"glue_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"execution_property": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_concurrent_runs": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"max_capacity": {
				Type:          schema.TypeFloat,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"allocated_capacity", "number_of_workers", "worker_type"},
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"notification_property": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"notify_delay_after": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
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
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2880,
			},
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"worker_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"allocated_capacity", "max_capacity"},
				ValidateFunc: validation.StringInSlice([]string{
					glue.WorkerTypeG1x,
					glue.WorkerTypeG2x,
					glue.WorkerTypeStandard,
				}, false),
			},
			"number_of_workers": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"allocated_capacity", "max_capacity"},
				ValidateFunc:  validation.IntAtLeast(2),
			},
		},
	}
}

func resourceAwsGlueJobCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.CreateJobInput{
		Command: expandGlueJobCommand(d.Get("command").([]interface{})),
		Name:    aws.String(name),
		Role:    aws.String(d.Get("role_arn").(string)),
		Tags:    keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GlueTags(),
		Timeout: aws.Int64(int64(d.Get("timeout").(int))),
	}

	if v, ok := d.GetOk("max_capacity"); ok {
		input.MaxCapacity = aws.Float64(v.(float64))
	} else {
		if v, ok := d.GetOk("allocated_capacity"); ok {
			input.MaxCapacity = aws.Float64(float64(v.(int)))
			log.Printf("[WARN] Using deprecated `allocated_capacity' attribute.")
		}
	}

	if v, ok := d.GetOk("connections"); ok {
		input.Connections = &glue.ConnectionsList{
			Connections: expandStringList(v.([]interface{})),
		}
	}

	if kv, ok := d.GetOk("default_arguments"); ok {
		defaultArgumentsMap := make(map[string]string)
		for k, v := range kv.(map[string]interface{}) {
			defaultArgumentsMap[k] = v.(string)
		}
		input.DefaultArguments = aws.StringMap(defaultArgumentsMap)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_property"); ok {
		input.ExecutionProperty = expandGlueExecutionProperty(v.([]interface{}))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		input.MaxRetries = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("notification_property"); ok {
		input.NotificationProperty = expandGlueNotificationProperty(v.([]interface{}))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating Glue Job: %s", input)
	_, err := conn.CreateJob(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Job (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceAwsGlueJobRead(d, meta)
}

func resourceAwsGlueJobRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &glue.GetJobInput{
		JobName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Job: %s", input)
	output, err := conn.GetJob(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Job (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Job (%s): %s", d.Id(), err)
	}

	job := output.Job
	if job == nil {
		log.Printf("[WARN] Glue Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	jobARN := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "glue",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("job/%s", d.Id()),
	}.String()
	d.Set("arn", jobARN)

	if err := d.Set("command", flattenGlueJobCommand(job.Command)); err != nil {
		return fmt.Errorf("error setting command: %s", err)
	}
	if err := d.Set("connections", flattenGlueConnectionsList(job.Connections)); err != nil {
		return fmt.Errorf("error setting connections: %s", err)
	}
	if err := d.Set("default_arguments", aws.StringValueMap(job.DefaultArguments)); err != nil {
		return fmt.Errorf("error setting default_arguments: %s", err)
	}
	d.Set("description", job.Description)
	d.Set("glue_version", job.GlueVersion)
	if err := d.Set("execution_property", flattenGlueExecutionProperty(job.ExecutionProperty)); err != nil {
		return fmt.Errorf("error setting execution_property: %s", err)
	}
	d.Set("max_capacity", aws.Float64Value(job.MaxCapacity))
	d.Set("max_retries", int(aws.Int64Value(job.MaxRetries)))
	if err := d.Set("notification_property", flattenGlueNotificationProperty(job.NotificationProperty)); err != nil {
		return fmt.Errorf("error setting notification_property: #{err}")
	}
	d.Set("name", job.Name)
	d.Set("role_arn", job.Role)

	tags, err := keyvaluetags.GlueListTags(conn, jobARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Job (%s): %s", jobARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("timeout", int(aws.Int64Value(job.Timeout)))
	if err := d.Set("security_configuration", job.SecurityConfiguration); err != nil {
		return fmt.Errorf("error setting security_configuration: %s", err)
	}

	d.Set("worker_type", job.WorkerType)
	d.Set("number_of_workers", int(aws.Int64Value(job.NumberOfWorkers)))

	// TODO: Deprecated fields - remove in next major version
	d.Set("allocated_capacity", int(aws.Int64Value(job.AllocatedCapacity)))

	return nil
}

func resourceAwsGlueJobUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	if d.HasChanges("allocated_capacity", "command", "connections", "default_arguments", "description",
		"execution_property", "glue_version", "max_capacity", "max_retries", "notification_property", "number_of_workers",
		"role_arn", "security_configuration", "timeout", "worker_type") {
		jobUpdate := &glue.JobUpdate{
			Command: expandGlueJobCommand(d.Get("command").([]interface{})),
			Role:    aws.String(d.Get("role_arn").(string)),
			Timeout: aws.Int64(int64(d.Get("timeout").(int))),
		}

		if v, ok := d.GetOk("number_of_workers"); ok {
			jobUpdate.NumberOfWorkers = aws.Int64(int64(v.(int)))
		} else {
			if v, ok := d.GetOk("max_capacity"); ok {
				jobUpdate.MaxCapacity = aws.Float64(v.(float64))
			}
			if d.HasChange("allocated_capacity") {
				jobUpdate.MaxCapacity = aws.Float64(float64(d.Get("allocated_capacity").(int)))
				log.Printf("[WARN] Using deprecated `allocated_capacity' attribute.")
			}
		}

		if v, ok := d.GetOk("connections"); ok {
			jobUpdate.Connections = &glue.ConnectionsList{
				Connections: expandStringList(v.([]interface{})),
			}
		}

		if kv, ok := d.GetOk("default_arguments"); ok {
			defaultArgumentsMap := make(map[string]string)
			for k, v := range kv.(map[string]interface{}) {
				defaultArgumentsMap[k] = v.(string)
			}
			jobUpdate.DefaultArguments = aws.StringMap(defaultArgumentsMap)
		}

		if v, ok := d.GetOk("description"); ok {
			jobUpdate.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("glue_version"); ok {
			jobUpdate.GlueVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("execution_property"); ok {
			jobUpdate.ExecutionProperty = expandGlueExecutionProperty(v.([]interface{}))
		}

		if v, ok := d.GetOk("max_retries"); ok {
			jobUpdate.MaxRetries = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("notification_property"); ok {
			jobUpdate.NotificationProperty = expandGlueNotificationProperty(v.([]interface{}))
		}

		if v, ok := d.GetOk("security_configuration"); ok {
			jobUpdate.SecurityConfiguration = aws.String(v.(string))
		}

		if v, ok := d.GetOk("worker_type"); ok {
			jobUpdate.WorkerType = aws.String(v.(string))
		}

		input := &glue.UpdateJobInput{
			JobName:   aws.String(d.Id()),
			JobUpdate: jobUpdate,
		}

		log.Printf("[DEBUG] Updating Glue Job: %s", input)
		_, err := conn.UpdateJob(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Job (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GlueUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsGlueJobRead(d, meta)
}

func resourceAwsGlueJobDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Deleting Glue Job: %s", d.Id())
	err := deleteGlueJob(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Job (%s): %s", d.Id(), err)
	}

	return nil
}

func deleteGlueJob(conn *glue.Glue, jobName string) error {
	input := &glue.DeleteJobInput{
		JobName: aws.String(jobName),
	}

	_, err := conn.DeleteJob(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func expandGlueExecutionProperty(l []interface{}) *glue.ExecutionProperty {
	m := l[0].(map[string]interface{})

	executionProperty := &glue.ExecutionProperty{
		MaxConcurrentRuns: aws.Int64(int64(m["max_concurrent_runs"].(int))),
	}

	return executionProperty
}

func expandGlueJobCommand(l []interface{}) *glue.JobCommand {
	m := l[0].(map[string]interface{})

	jobCommand := &glue.JobCommand{
		Name:           aws.String(m["name"].(string)),
		ScriptLocation: aws.String(m["script_location"].(string)),
	}

	if v, ok := m["python_version"].(string); ok && v != "" {
		jobCommand.PythonVersion = aws.String(v)
	}

	return jobCommand
}

func expandGlueNotificationProperty(l []interface{}) *glue.NotificationProperty {
	m := l[0].(map[string]interface{})

	notificationProperty := &glue.NotificationProperty{
		NotifyDelayAfter: aws.Int64(int64(m["notify_delay_after"].(int))),
	}

	return notificationProperty
}

func flattenGlueConnectionsList(connectionsList *glue.ConnectionsList) []interface{} {
	if connectionsList == nil {
		return []interface{}{}
	}

	return flattenStringList(connectionsList.Connections)
}

func flattenGlueExecutionProperty(executionProperty *glue.ExecutionProperty) []map[string]interface{} {
	if executionProperty == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"max_concurrent_runs": int(aws.Int64Value(executionProperty.MaxConcurrentRuns)),
	}

	return []map[string]interface{}{m}
}

func flattenGlueJobCommand(jobCommand *glue.JobCommand) []map[string]interface{} {
	if jobCommand == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"name":            aws.StringValue(jobCommand.Name),
		"script_location": aws.StringValue(jobCommand.ScriptLocation),
		"python_version":  aws.StringValue(jobCommand.PythonVersion),
	}

	return []map[string]interface{}{m}
}

func flattenGlueNotificationProperty(notificationProperty *glue.NotificationProperty) []map[string]interface{} {
	if notificationProperty == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notify_delay_after": int(aws.Int64Value(notificationProperty.NotifyDelayAfter)),
	}

	return []map[string]interface{}{m}
}
