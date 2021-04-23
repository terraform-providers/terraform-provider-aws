package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/synthetics/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/synthetics/waiter"
)

const awsMutexCanary = `aws_synthetics_canary`

func resourceAwsSyntheticsCanary() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSyntheticsCanaryCreate,
		Read:   resourceAwsSyntheticsCanaryRead,
		Update: resourceAwsSyntheticsCanaryUpdate,
		Delete: resourceAwsSyntheticsCanaryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"artifact_s3_location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimPrefix(new, "s3://") == old
				},
			},
			"engine_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"failure_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31,
				ValidateFunc: validation.IntBetween(1, 455),
			},
			"handler": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 21),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z_\-]+$`), "must contain only alphanumeric, hyphen, underscore."),
				),
			},
			"run_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_tracing": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"memory_in_mb": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.All(
								validation.IntDivisibleBy(64),
								validation.IntAtLeast(960),
							),
						},
						"timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(60, 14*60),
							Default:      840,
						},
					},
				},
			},
			"runtime_version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"s3_bucket": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
				RequiredWith:  []string{"s3_key"},
			},
			"s3_key": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
				RequiredWith:  []string{"s3_bucket"},
			},
			"s3_version": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zip_file"},
			},
			"schedule": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"duration_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"expression": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return new == "rate(0 minute)" && old == "rate(0 hour)"
							},
						},
					},
				},
			},
			"source_location_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"start_canary": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"success_retention_period": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      31,
				ValidateFunc: validation.IntBetween(1, 455),
			},
			"tags": tagsSchema(),
			"timeline": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_modified": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_started": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_stopped": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"vpc_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"zip_file": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"s3_bucket", "s3_key", "s3_version"},
			},
		},
	}
}

func resourceAwsSyntheticsCanaryCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn

	input := &synthetics.CreateCanaryInput{
		Name:               aws.String(d.Get("name").(string)),
		ArtifactS3Location: aws.String(d.Get("artifact_s3_location").(string)),
		ExecutionRoleArn:   aws.String(d.Get("execution_role_arn").(string)),
		RuntimeVersion:     aws.String(d.Get("runtime_version").(string)),
	}

	code, err := expandAwsSyntheticsCanaryCode(d)
	if err != nil {
		return err
	}

	input.Code = code

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().SyntheticsTags()
	}

	if v, ok := d.GetOk("run_config"); ok {
		input.RunConfig = expandAwsSyntheticsCanaryRunConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = expandAwsSyntheticsCanarySchedule(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		input.VpcConfig = expandAwsSyntheticsCanaryVpcConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("failure_retention_period"); ok {
		input.FailureRetentionPeriodInDays = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("success_retention_period"); ok {
		input.SuccessRetentionPeriodInDays = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] creating Synthetics Canary: %#v", input)

	resp, err := conn.CreateCanary(input)
	if err != nil {
		return fmt.Errorf("error creating Synthetics Canary: %w", err)
	}

	d.SetId(aws.StringValue(resp.Canary.Name))

	if _, err := waiter.CanaryReady(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) creation: %w", d.Id(), err)
	}

	if v := d.Get("start_canary"); v.(bool) {
		if err := syntheticsStartCanary(d.Id(), conn); err != nil {
			return err
		}
	}

	return resourceAwsSyntheticsCanaryRead(d, meta)
}

func resourceAwsSyntheticsCanaryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	resp, err := finder.CanaryByName(conn, d.Id())
	if err != nil {
		if isAWSErr(err, synthetics.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Synthetics Canary (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Synthetics Canary: %w", err)
	}

	canary := resp.Canary
	d.Set("artifact_s3_location", canary.ArtifactS3Location)
	d.Set("engine_arn", canary.EngineArn)
	d.Set("execution_role_arn", canary.ExecutionRoleArn)
	d.Set("failure_retention_period", canary.FailureRetentionPeriodInDays)
	d.Set("handler", canary.Code.Handler)
	d.Set("name", canary.Name)
	d.Set("runtime_version", canary.RuntimeVersion)
	d.Set("source_location_arn", canary.Code.SourceLocationArn)
	d.Set("status", canary.Status.State)
	d.Set("success_retention_period", canary.SuccessRetentionPeriodInDays)

	canaryArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   synthetics.ServiceName,
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("canary:%s", aws.StringValue(canary.Name)),
	}.String()

	d.Set("arn", canaryArn)

	if err := d.Set("vpc_config", flattenAwsSyntheticsCanaryVpcConfig(canary.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc config: %w", err)
	}

	if err := d.Set("run_config", flattenAwsSyntheticsCanaryRunConfig(canary.RunConfig)); err != nil {
		return fmt.Errorf("error setting run config: %w", err)
	}

	if err := d.Set("schedule", flattenAwsSyntheticsCanarySchedule(canary.Schedule)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}

	if err := d.Set("timeline", flattenAwsSyntheticsCanaryTimeline(canary.Timeline)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.SyntheticsKeyValueTags(canary.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}

func resourceAwsSyntheticsCanaryUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn

	if d.HasChangesExcept("tags", "start_canary") {
		input := &synthetics.UpdateCanaryInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange("vpc_config") {
			input.VpcConfig = expandAwsSyntheticsCanaryVpcConfig(d.Get("vpc_config").([]interface{}))
		}

		if d.HasChange("runtime_version") {
			input.RuntimeVersion = aws.String(d.Get("runtime_version").(string))
		}

		if d.HasChanges("handler", "zip_file", "s3_bucket", "s3_key", "s3_version") {
			code, err := expandAwsSyntheticsCanaryCode(d)
			if err != nil {
				return err
			}
			input.Code = code
		}

		if d.HasChange("run_config") {
			input.RunConfig = expandAwsSyntheticsCanaryRunConfig(d.Get("run_config").([]interface{}))
		}

		if d.HasChange("schedule") {
			input.Schedule = expandAwsSyntheticsCanarySchedule(d.Get("schedule").([]interface{}))
		}

		if d.HasChange("success_retention_period") {
			_, n := d.GetChange("success_retention_period")
			input.SuccessRetentionPeriodInDays = aws.Int64(int64(n.(int)))
		}

		if d.HasChange("failure_retention_period") {
			_, n := d.GetChange("failure_retention_period")
			input.FailureRetentionPeriodInDays = aws.Int64(int64(n.(int)))
		}

		if d.HasChange("execution_role_arn") {
			_, n := d.GetChange("execution_role_arn")
			input.ExecutionRoleArn = aws.String(n.(string))
		}

		status := d.Get("status").(string)
		if status == synthetics.CanaryStateRunning {
			if err := syntheticsStopCanary(d.Id(), conn); err != nil {
				return err
			}
		}

		_, err := conn.UpdateCanary(input)
		if err != nil {
			return fmt.Errorf("error updating Synthetics Canary: %w", err)
		}

		if status != synthetics.CanaryStateReady {
			if _, err := waiter.CanaryStopped(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for Synthetics Canary (%s) to be stopped: %w", d.Id(), err)
			}
		} else {
			if _, err := waiter.CanaryReady(conn, d.Id()); err != nil {
				return fmt.Errorf("error waiting for Synthetics Canary (%s) updating: %w", d.Id(), err)
			}
		}

		if v := d.Get("start_canary"); v.(bool) {
			if err := syntheticsStartCanary(d.Id(), conn); err != nil {
				return err
			}
		}
	}

	if d.HasChange("start_canary") {
		status := d.Get("status").(string)
		if v := d.Get("start_canary"); v.(bool) {
			if status != synthetics.CanaryStateRunning {
				if err := syntheticsStartCanary(d.Id(), conn); err != nil {
					return err
				}
			}
		} else {
			if status == synthetics.CanaryStateRunning {
				if err := syntheticsStopCanary(d.Id(), conn); err != nil {
					return err
				}
			}
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.SyntheticsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Synthetics Canary (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAwsSyntheticsCanaryRead(d, meta)
}

func resourceAwsSyntheticsCanaryDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).syntheticsconn

	if status := d.Get("status"); status.(string) == synthetics.CanaryStateRunning {
		if err := syntheticsStopCanary(d.Id(), conn); err != nil {
			log.Print("could not stop canary before delete")
		}
	}

	input := &synthetics.DeleteCanaryInput{
		Name: aws.String(d.Id()),
	}

	_, err := conn.DeleteCanary(input)
	if err != nil {
		if isAWSErr(err, synthetics.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error deleting Synthetics Canary: %w", err)
	}

	if _, err := waiter.CanaryDeleted(conn, d.Id()); err != nil {
		if isAWSErr(err, synthetics.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("error waiting for Synthetics Canary (%s) deletion: %w", d.Id(), err)
	}

	return nil
}

func expandAwsSyntheticsCanaryCode(d *schema.ResourceData) (*synthetics.CanaryCodeInput, error) {
	codeConfig := &synthetics.CanaryCodeInput{
		Handler: aws.String(d.Get("handler").(string)),
	}

	if v, ok := d.GetOk("zip_file"); ok {
		awsMutexKV.Lock(awsMutexCanary)
		defer awsMutexKV.Unlock(awsMutexCanary)
		file, err := loadFileContent(v.(string))
		if err != nil {
			return nil, fmt.Errorf("unable to load %q: %w", v.(string), err)
		}
		codeConfig.ZipFile = file
	} else {
		codeConfig.S3Bucket = aws.String(d.Get("s3_bucket").(string))
		codeConfig.S3Key = aws.String(d.Get("s3_key").(string))

		if v, ok := d.GetOk("s3_version"); ok {
			codeConfig.S3Version = aws.String(v.(string))
		}
	}

	return codeConfig, nil
}

func expandAwsSyntheticsCanarySchedule(l []interface{}) *synthetics.CanaryScheduleInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &synthetics.CanaryScheduleInput{
		Expression: aws.String(m["expression"].(string)),
	}

	if v, ok := m["duration_in_seconds"]; ok {
		codeConfig.DurationInSeconds = aws.Int64(int64(v.(int)))
	}

	return codeConfig
}

func flattenAwsSyntheticsCanarySchedule(canarySchedule *synthetics.CanaryScheduleOutput) []interface{} {
	if canarySchedule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"expression":          aws.StringValue(canarySchedule.Expression),
		"duration_in_seconds": aws.Int64Value(canarySchedule.DurationInSeconds),
	}

	return []interface{}{m}
}

func expandAwsSyntheticsCanaryRunConfig(l []interface{}) *synthetics.CanaryRunConfigInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &synthetics.CanaryRunConfigInput{
		TimeoutInSeconds: aws.Int64(int64(m["timeout_in_seconds"].(int))),
	}

	if v, ok := m["memory_in_mb"].(int); ok && v > 0 {
		codeConfig.MemoryInMB = aws.Int64(int64(v))
	}

	if v, ok := m["active_tracing"].(bool); ok {
		codeConfig.ActiveTracing = aws.Bool(v)
	}

	return codeConfig
}

func flattenAwsSyntheticsCanaryRunConfig(canaryCodeOut *synthetics.CanaryRunConfigOutput) []interface{} {
	if canaryCodeOut == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"timeout_in_seconds": aws.Int64Value(canaryCodeOut.TimeoutInSeconds),
		"memory_in_mb":       aws.Int64Value(canaryCodeOut.MemoryInMB),
		"active_tracing":     aws.BoolValue(canaryCodeOut.ActiveTracing),
	}

	return []interface{}{m}
}

func flattenAwsSyntheticsCanaryVpcConfig(canaryVpcOutput *synthetics.VpcConfigOutput) []interface{} {
	if canaryVpcOutput == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"subnet_ids":         flattenStringSet(canaryVpcOutput.SubnetIds),
		"security_group_ids": flattenStringSet(canaryVpcOutput.SecurityGroupIds),
		"vpc_id":             aws.StringValue(canaryVpcOutput.VpcId),
	}

	return []interface{}{m}
}

func expandAwsSyntheticsCanaryVpcConfig(l []interface{}) *synthetics.VpcConfigInput {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	codeConfig := &synthetics.VpcConfigInput{
		SubnetIds:        expandStringSet(m["subnet_ids"].(*schema.Set)),
		SecurityGroupIds: expandStringSet(m["security_group_ids"].(*schema.Set)),
	}

	return codeConfig
}

func flattenAwsSyntheticsCanaryTimeline(timeline *synthetics.CanaryTimeline) []interface{} {
	if timeline == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"created": aws.TimeValue(timeline.Created).Format(time.RFC3339),
	}

	if timeline.LastModified != nil {
		m["last_modified"] = aws.TimeValue(timeline.LastModified).Format(time.RFC3339)
	}

	if timeline.LastStarted != nil {
		m["last_started"] = aws.TimeValue(timeline.LastStarted).Format(time.RFC3339)
	}

	if timeline.LastStopped != nil {
		m["last_stopped"] = aws.TimeValue(timeline.LastStopped).Format(time.RFC3339)
	}

	return []interface{}{m}
}

func syntheticsStartCanary(name string, conn *synthetics.Synthetics) error {
	startInput := &synthetics.StartCanaryInput{
		Name: aws.String(name),
	}

	log.Print("starting Canary")
	_, err := conn.StartCanary(startInput)
	if err != nil {
		return fmt.Errorf("error starting Synthetics Canary: %w", err)
	}

	if _, err := waiter.CanaryRunning(conn, name); err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) to be running: %w", name, err)
	}

	return nil
}

func syntheticsStopCanary(name string, conn *synthetics.Synthetics) error {
	stopInput := &synthetics.StopCanaryInput{
		Name: aws.String(name),
	}

	_, err := conn.StopCanary(stopInput)
	if err != nil {
		return fmt.Errorf("error stopping Synthetics Canary: %w", err)
	}

	if _, err := waiter.CanaryStopped(conn, name); err != nil {
		return fmt.Errorf("error waiting for Synthetics Canary (%s) to be stopped: %w", name, err)
	}

	return nil
}
