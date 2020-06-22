package aws

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAthenaWorkgroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAthenaWorkgroupCreate,
		Read:   resourceAwsAthenaWorkgroupRead,
		Update: resourceAwsAthenaWorkgroupUpdate,
		Delete: resourceAwsAthenaWorkgroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bytes_scanned_cutoff_per_query": {
							Type:     schema.TypeInt,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.IntAtLeast(10485760),
								validation.IntInSlice([]int{0}),
							),
						},
						"enforce_workgroup_configuration": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"publish_cloudwatch_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"result_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"encryption_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"encryption_option": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														athena.EncryptionOptionCseKms,
														athena.EncryptionOptionSseKms,
														athena.EncryptionOptionSseS3,
													}, false),
												},
												"kms_key_arn": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validateArn,
												},
											},
										},
									},
									"output_location": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._-]+$`), "must contain only alphanumeric characters, periods, underscores, and hyphens"),
				),
			},
			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  athena.WorkGroupStateEnabled,
				ValidateFunc: validation.StringInSlice([]string{
					athena.WorkGroupStateDisabled,
					athena.WorkGroupStateEnabled,
				}, false),
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAthenaWorkgroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	name := d.Get("name").(string)

	input := &athena.CreateWorkGroupInput{
		Configuration: expandAthenaWorkGroupConfiguration(d.Get("configuration").([]interface{})),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	// Prevent the below error:
	// InvalidRequestException: Tags provided upon WorkGroup creation must not be empty
	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().AthenaTags()
	}

	_, err := conn.CreateWorkGroup(input)

	if err != nil {
		return fmt.Errorf("error creating Athena WorkGroup: %s", err)
	}

	d.SetId(name)

	if v := d.Get("state").(string); v == athena.WorkGroupStateDisabled {
		input := &athena.UpdateWorkGroupInput{
			State:     aws.String(v),
			WorkGroup: aws.String(d.Id()),
		}

		if _, err := conn.UpdateWorkGroup(input); err != nil {
			return fmt.Errorf("error disabling Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	return resourceAwsAthenaWorkgroupRead(d, meta)
}

func resourceAwsAthenaWorkgroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &athena.GetWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	resp, err := conn.GetWorkGroup(input)

	if isAWSErr(err, athena.ErrCodeInvalidRequestException, "is not found") {
		log.Printf("[WARN] Athena WorkGroup (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Athena WorkGroup (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "athena",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("workgroup/%s", d.Id()),
	}

	d.Set("arn", arn.String())
	d.Set("description", resp.WorkGroup.Description)

	if err := d.Set("configuration", flattenAthenaWorkGroupConfiguration(resp.WorkGroup.Configuration)); err != nil {
		return fmt.Errorf("error setting configuration: %s", err)
	}

	d.Set("name", resp.WorkGroup.Name)
	d.Set("state", resp.WorkGroup.State)

	if v, ok := d.GetOk("force_destroy"); ok {
		d.Set("force_destroy", v.(bool))
	} else {
		d.Set("force_destroy", false)
	}

	tags, err := keyvaluetags.AthenaListTags(conn, arn.String())

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsAthenaWorkgroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	input := &athena.DeleteWorkGroupInput{
		WorkGroup: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("force_destroy"); ok {
		input.RecursiveDeleteOption = aws.Bool(v.(bool))
	}
	_, err := conn.DeleteWorkGroup(input)

	if err != nil {
		return fmt.Errorf("error deleting Athena WorkGroup (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsAthenaWorkgroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).athenaconn

	workGroupUpdate := false

	input := &athena.UpdateWorkGroupInput{
		WorkGroup: aws.String(d.Get("name").(string)),
	}

	if d.HasChange("configuration") {
		workGroupUpdate = true
		input.ConfigurationUpdates = expandAthenaWorkGroupConfigurationUpdates(d.Get("configuration").([]interface{}))
	}

	if d.HasChange("description") {
		workGroupUpdate = true
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("state") {
		workGroupUpdate = true
		input.State = aws.String(d.Get("state").(string))
	}

	if workGroupUpdate {
		_, err := conn.UpdateWorkGroup(input)

		if err != nil {
			return fmt.Errorf("error updating Athena WorkGroup (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AthenaUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsAthenaWorkgroupRead(d, meta)
}

func expandAthenaWorkGroupConfiguration(l []interface{}) *athena.WorkGroupConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &athena.WorkGroupConfiguration{}

	if v, ok := m["bytes_scanned_cutoff_per_query"]; ok && v.(int) > 0 {
		configuration.BytesScannedCutoffPerQuery = aws.Int64(int64(v.(int)))
	}

	if v, ok := m["enforce_workgroup_configuration"]; ok {
		configuration.EnforceWorkGroupConfiguration = aws.Bool(v.(bool))
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"]; ok {
		configuration.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["result_configuration"]; ok {
		configuration.ResultConfiguration = expandAthenaWorkGroupResultConfiguration(v.([]interface{}))
	}

	return configuration
}

func expandAthenaWorkGroupConfigurationUpdates(l []interface{}) *athena.WorkGroupConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configurationUpdates := &athena.WorkGroupConfigurationUpdates{}

	if v, ok := m["bytes_scanned_cutoff_per_query"]; ok && v.(int) > 0 {
		configurationUpdates.BytesScannedCutoffPerQuery = aws.Int64(int64(v.(int)))
	} else {
		configurationUpdates.RemoveBytesScannedCutoffPerQuery = aws.Bool(true)
	}

	if v, ok := m["enforce_workgroup_configuration"]; ok {
		configurationUpdates.EnforceWorkGroupConfiguration = aws.Bool(v.(bool))
	}

	if v, ok := m["publish_cloudwatch_metrics_enabled"]; ok {
		configurationUpdates.PublishCloudWatchMetricsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["result_configuration"]; ok {
		configurationUpdates.ResultConfigurationUpdates = expandAthenaWorkGroupResultConfigurationUpdates(v.([]interface{}))
	}

	return configurationUpdates
}

func expandAthenaWorkGroupResultConfiguration(l []interface{}) *athena.ResultConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfiguration := &athena.ResultConfiguration{}

	if v, ok := m["encryption_configuration"]; ok {
		resultConfiguration.EncryptionConfiguration = expandAthenaWorkGroupEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := m["output_location"]; ok && v.(string) != "" {
		resultConfiguration.OutputLocation = aws.String(v.(string))
	}

	return resultConfiguration
}

func expandAthenaWorkGroupResultConfigurationUpdates(l []interface{}) *athena.ResultConfigurationUpdates {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	resultConfigurationUpdates := &athena.ResultConfigurationUpdates{}

	if v, ok := m["encryption_configuration"]; ok {
		resultConfigurationUpdates.EncryptionConfiguration = expandAthenaWorkGroupEncryptionConfiguration(v.([]interface{}))
	} else {
		resultConfigurationUpdates.RemoveEncryptionConfiguration = aws.Bool(true)
	}

	if v, ok := m["output_location"]; ok && v.(string) != "" {
		resultConfigurationUpdates.OutputLocation = aws.String(v.(string))
	} else {
		resultConfigurationUpdates.RemoveOutputLocation = aws.Bool(true)
	}

	return resultConfigurationUpdates
}

func expandAthenaWorkGroupEncryptionConfiguration(l []interface{}) *athena.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	encryptionConfiguration := &athena.EncryptionConfiguration{}

	if v, ok := m["encryption_option"]; ok && v.(string) != "" {
		encryptionConfiguration.EncryptionOption = aws.String(v.(string))
	}

	if v, ok := m["kms_key_arn"]; ok && v.(string) != "" {
		encryptionConfiguration.KmsKey = aws.String(v.(string))
	}

	return encryptionConfiguration
}

func flattenAthenaWorkGroupConfiguration(configuration *athena.WorkGroupConfiguration) []interface{} {
	if configuration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bytes_scanned_cutoff_per_query":     aws.Int64Value(configuration.BytesScannedCutoffPerQuery),
		"enforce_workgroup_configuration":    aws.BoolValue(configuration.EnforceWorkGroupConfiguration),
		"publish_cloudwatch_metrics_enabled": aws.BoolValue(configuration.PublishCloudWatchMetricsEnabled),
		"result_configuration":               flattenAthenaWorkGroupResultConfiguration(configuration.ResultConfiguration),
	}

	return []interface{}{m}
}

func flattenAthenaWorkGroupResultConfiguration(resultConfiguration *athena.ResultConfiguration) []interface{} {
	if resultConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encryption_configuration": flattenAthenaWorkGroupEncryptionConfiguration(resultConfiguration.EncryptionConfiguration),
		"output_location":          aws.StringValue(resultConfiguration.OutputLocation),
	}

	return []interface{}{m}
}

func flattenAthenaWorkGroupEncryptionConfiguration(encryptionConfiguration *athena.EncryptionConfiguration) []interface{} {
	if encryptionConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"encryption_option": aws.StringValue(encryptionConfiguration.EncryptionOption),
		"kms_key_arn":       aws.StringValue(encryptionConfiguration.KmsKey),
	}

	return []interface{}{m}
}
