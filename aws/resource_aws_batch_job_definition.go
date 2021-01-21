package aws

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/equivalency"
)

func resourceAwsBatchJobDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBatchJobDefinitionCreate,
		Read:   resourceAwsBatchJobDefinitionRead,
		Update: resourceAwsBatchJobDefinitionUpdate,
		Delete: resourceAwsBatchJobDefinitionDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("arn", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateBatchName,
			},
			"container_properties": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalency.EquivalentBatchContainerPropertiesJSON(old, new)

					return equal
				},
				ValidateFunc: validateAwsBatchJobContainerProperties,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"retry_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 10),
						},
					},
				},
			},
			"tags": tagsSchema(),
			"timeout": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempt_duration_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(60),
						},
					},
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{batch.JobDefinitionTypeContainer}, true),
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsBatchJobDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	name := d.Get("name").(string)

	input := &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(name),
		Type:              aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("container_properties"); ok {
		props, err := expandBatchJobContainerProperties(v.(string))
		if err != nil {
			return fmt.Errorf("%s %q", err, name)
		}
		input.ContainerProperties = props
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandJobDefinitionParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("retry_strategy"); ok {
		input.RetryStrategy = expandJobDefinitionRetryStrategy(v.([]interface{}))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().BatchTags()
	}

	if v, ok := d.GetOk("timeout"); ok {
		input.Timeout = expandJobDefinitionTimeout(v.([]interface{}))
	}

	out, err := conn.RegisterJobDefinition(input)
	if err != nil {
		return fmt.Errorf("%s %q", err, name)
	}
	d.SetId(aws.StringValue(out.JobDefinitionArn))
	d.Set("arn", out.JobDefinitionArn)
	return resourceAwsBatchJobDefinitionRead(d, meta)
}

func resourceAwsBatchJobDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	job, err := getJobDefinition(conn, arn)
	if err != nil {
		return fmt.Errorf("%s %q", err, arn)
	}
	if job == nil {
		d.SetId("")
		return nil
	}
	d.Set("arn", job.JobDefinitionArn)

	containerProperties, err := flattenBatchContainerProperties(job.ContainerProperties)

	if err != nil {
		return fmt.Errorf("error converting Batch Container Properties to JSON: %s", err)
	}

	if err := d.Set("container_properties", containerProperties); err != nil {
		return fmt.Errorf("error setting container_properties: %s", err)
	}

	d.Set("name", job.JobDefinitionName)

	d.Set("parameters", aws.StringValueMap(job.Parameters))

	if err := d.Set("retry_strategy", flattenBatchRetryStrategy(job.RetryStrategy)); err != nil {
		return fmt.Errorf("error setting retry_strategy: %s", err)
	}

	if err := d.Set("tags", keyvaluetags.BatchKeyValueTags(job.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("timeout", flattenBatchJobTimeout(job.Timeout)); err != nil {
		return fmt.Errorf("error setting timeout: %s", err)
	}

	d.Set("revision", job.Revision)
	d.Set("type", job.Type)
	return nil
}

func resourceAwsBatchJobDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.BatchUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return nil
}

func resourceAwsBatchJobDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	arn := d.Get("arn").(string)
	_, err := conn.DeregisterJobDefinition(&batch.DeregisterJobDefinitionInput{
		JobDefinition: aws.String(arn),
	})
	if err != nil {
		return fmt.Errorf("%s %q", err, arn)
	}

	return nil
}

func getJobDefinition(conn *batch.Batch, arn string) (*batch.JobDefinition, error) {
	describeOpts := &batch.DescribeJobDefinitionsInput{
		JobDefinitions: []*string{aws.String(arn)},
	}
	resp, err := conn.DescribeJobDefinitions(describeOpts)
	if err != nil {
		return nil, err
	}

	numJobDefinitions := len(resp.JobDefinitions)
	switch {
	case numJobDefinitions == 0:
		return nil, nil
	case numJobDefinitions == 1:
		if *resp.JobDefinitions[0].Status == "ACTIVE" {
			return resp.JobDefinitions[0], nil
		}
		return nil, nil
	case numJobDefinitions > 1:
		return nil, fmt.Errorf("Multiple Job Definitions with name %s", arn)
	}
	return nil, nil
}

func validateAwsBatchJobContainerProperties(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandBatchJobContainerProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job container_properties is invalid: %s", err))
	}
	return
}

func expandBatchJobContainerProperties(rawProps string) (*batch.ContainerProperties, error) {
	var props *batch.ContainerProperties

	err := json.Unmarshal([]byte(rawProps), &props)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %s", err)
	}

	return props, nil
}

// Convert batch.ContainerProperties object into its JSON representation
func flattenBatchContainerProperties(containerProperties *batch.ContainerProperties) (string, error) {
	b, err := jsonutil.BuildJSON(containerProperties)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func expandJobDefinitionParameters(params map[string]interface{}) map[string]*string {
	var jobParams = make(map[string]*string)
	for k, v := range params {
		jobParams[k] = aws.String(v.(string))
	}

	return jobParams
}

func expandJobDefinitionRetryStrategy(item []interface{}) *batch.RetryStrategy {
	retryStrategy := &batch.RetryStrategy{}
	data := item[0].(map[string]interface{})

	if v, ok := data["attempts"].(int); ok && v > 0 && v <= 10 {
		retryStrategy.Attempts = aws.Int64(int64(v))
	}

	return retryStrategy
}

func flattenBatchRetryStrategy(item *batch.RetryStrategy) []map[string]interface{} {
	data := []map[string]interface{}{}
	if item != nil && item.Attempts != nil {
		data = append(data, map[string]interface{}{
			"attempts": int(aws.Int64Value(item.Attempts)),
		})
	}
	return data
}

func expandJobDefinitionTimeout(item []interface{}) *batch.JobTimeout {
	timeout := &batch.JobTimeout{}
	data := item[0].(map[string]interface{})

	if v, ok := data["attempt_duration_seconds"].(int); ok && v >= 60 {
		timeout.AttemptDurationSeconds = aws.Int64(int64(v))
	}

	return timeout
}

func flattenBatchJobTimeout(item *batch.JobTimeout) []map[string]interface{} {
	data := []map[string]interface{}{}
	if item != nil && item.AttemptDurationSeconds != nil {
		data = append(data, map[string]interface{}{
			"attempt_duration_seconds": int(aws.Int64Value(item.AttemptDurationSeconds)),
		})
	}
	return data
}
