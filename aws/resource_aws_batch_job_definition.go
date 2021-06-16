package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/equivalency"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsBatchJobDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBatchJobDefinitionCreate,
		Read:   resourceAwsBatchJobDefinitionRead,
		Update: resourceAwsBatchJobDefinitionUpdate,
		Delete: resourceAwsBatchJobDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"platform_capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(batch.PlatformCapability_Values(), false),
				},
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
						"evaluate_on_exit": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MinItems: 0,
							MaxItems: 5,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										StateFunc: func(v interface{}) string {
											return strings.ToLower(v.(string))
										},
										ValidateFunc: validation.StringInSlice(batch.RetryAction_Values(), true),
									},
									"on_exit_code": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexp.MustCompile(`^[0-9]*\*?$`), "must contain only numbers, and can optionally end with an asterisk"),
										),
									},
									"on_reason": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
										),
									},
									"on_status_reason": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\.:\s]*\*?$`), "must contain letters, numbers, periods, colons, and white space, and can optionally end with an asterisk"),
										),
									},
								},
							},
						},
					},
				},
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"propagate_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
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

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsBatchJobDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(name),
		Type:              aws.String(d.Get("type").(string)),
		PropagateTags:     aws.Bool(d.Get("propagate_tags").(bool)),
	}

	if v, ok := d.GetOk("container_properties"); ok {
		props, err := expandBatchJobContainerProperties(v.(string))
		if err != nil {
			return err
		}

		input.ContainerProperties = props
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandJobDefinitionParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
		input.PlatformCapabilities = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("retry_strategy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.RetryStrategy = expandBatchRetryStrategy(v.([]interface{})[0].(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().BatchTags()
	}

	if v, ok := d.GetOk("timeout"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Timeout = expandBatchJobTimeout(v.([]interface{})[0].(map[string]interface{}))
	}

	output, err := conn.RegisterJobDefinition(input)

	if err != nil {
		return fmt.Errorf("error creating Batch Job Definition (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.JobDefinitionArn))

	return resourceAwsBatchJobDefinitionRead(d, meta)
}

func resourceAwsBatchJobDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	jobDefinition, err := finder.JobDefinitionByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Batch Job Definition (%s): %w", d.Id(), err)
	}

	d.Set("arn", jobDefinition.JobDefinitionArn)

	containerProperties, err := flattenBatchContainerProperties(jobDefinition.ContainerProperties)

	if err != nil {
		return fmt.Errorf("error converting Batch Container Properties to JSON: %w", err)
	}

	if err := d.Set("container_properties", containerProperties); err != nil {
		return fmt.Errorf("error setting container_properties: %w", err)
	}

	d.Set("name", jobDefinition.JobDefinitionName)
	d.Set("parameters", aws.StringValueMap(jobDefinition.Parameters))
	d.Set("platform_capabilities", aws.StringValueSlice(jobDefinition.PlatformCapabilities))
	d.Set("propagate_tags", jobDefinition.PropagateTags)

	if jobDefinition.RetryStrategy != nil {
		if err := d.Set("retry_strategy", []interface{}{flattenBatchRetryStrategy(jobDefinition.RetryStrategy)}); err != nil {
			return fmt.Errorf("error setting retry_strategy: %w", err)
		}
	} else {
		d.Set("retry_strategy", nil)
	}

	tags := keyvaluetags.BatchKeyValueTags(jobDefinition.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if jobDefinition.Timeout != nil {
		if err := d.Set("timeout", []interface{}{flattenBatchJobTimeout(jobDefinition.Timeout)}); err != nil {
			return fmt.Errorf("error setting timeout: %w", err)
		}
	} else {
		d.Set("timeout", nil)
	}

	d.Set("revision", jobDefinition.Revision)
	d.Set("type", jobDefinition.Type)

	return nil
}

func resourceAwsBatchJobDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.BatchUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return nil
}

func resourceAwsBatchJobDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	_, err := conn.DeregisterJobDefinition(&batch.DeregisterJobDefinitionInput{
		JobDefinition: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting Batch Job Definition (%s): %w", d.Id(), err)
	}

	return nil
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

func expandBatchRetryStrategy(tfMap map[string]interface{}) *batch.RetryStrategy {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.RetryStrategy{}

	if v, ok := tfMap["attempts"].(int); ok && v != 0 {
		apiObject.Attempts = aws.Int64(int64(v))
	}

	if v, ok := tfMap["evaluate_on_exit"].([]interface{}); ok && len(v) > 0 {
		apiObject.EvaluateOnExit = expandBatchEvaluateOnExits(v)
	}

	return apiObject
}

func expandBatchEvaluateOnExit(tfMap map[string]interface{}) *batch.EvaluateOnExit {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.EvaluateOnExit{}

	if v, ok := tfMap["action"].(string); ok && v != "" {
		apiObject.Action = aws.String(v)
	}

	if v, ok := tfMap["on_exit_code"].(string); ok && v != "" {
		apiObject.OnExitCode = aws.String(v)
	}

	if v, ok := tfMap["on_reason"].(string); ok && v != "" {
		apiObject.OnReason = aws.String(v)
	}

	if v, ok := tfMap["on_status_reason"].(string); ok && v != "" {
		apiObject.OnStatusReason = aws.String(v)
	}

	return apiObject
}

func expandBatchEvaluateOnExits(tfList []interface{}) []*batch.EvaluateOnExit {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*batch.EvaluateOnExit

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandBatchEvaluateOnExit(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenBatchRetryStrategy(apiObject *batch.RetryStrategy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Attempts; v != nil {
		tfMap["attempts"] = aws.Int64Value(v)
	}

	if v := apiObject.EvaluateOnExit; v != nil {
		tfMap["evaluate_on_exit"] = flattenBatchEvaluateOnExits(v)
	}

	return tfMap
}

func flattenBatchEvaluateOnExit(apiObject *batch.EvaluateOnExit) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Action; v != nil {
		tfMap["action"] = aws.StringValue(v)
	}

	if v := apiObject.OnExitCode; v != nil {
		tfMap["on_exit_code"] = aws.StringValue(v)
	}

	if v := apiObject.OnReason; v != nil {
		tfMap["on_reason"] = aws.StringValue(v)
	}

	if v := apiObject.OnStatusReason; v != nil {
		tfMap["on_status_reason"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenBatchEvaluateOnExits(apiObjects []*batch.EvaluateOnExit) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenBatchEvaluateOnExit(apiObject))
	}

	return tfList
}

func expandBatchJobTimeout(tfMap map[string]interface{}) *batch.JobTimeout {
	if tfMap == nil {
		return nil
	}

	apiObject := &batch.JobTimeout{}

	if v, ok := tfMap["attempt_duration_seconds"].(int); ok && v != 0 {
		apiObject.AttemptDurationSeconds = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenBatchJobTimeout(apiObject *batch.JobTimeout) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AttemptDurationSeconds; v != nil {
		tfMap["attempt_duration_seconds"] = aws.Int64Value(v)
	}

	return tfMap
}
