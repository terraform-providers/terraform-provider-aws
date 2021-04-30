package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	tfsns "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sns"
)

func resourceAwsSnsTopic() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSnsTopicCreate,
		Read:   resourceAwsSnsTopicRead,
		Update: resourceAwsSnsTopicUpdate,
		Delete: resourceAwsSnsTopicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			resourceAwsSnsTopicCustomizeDiff,
			SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
			},
			"display_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"delivery_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         false,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"application_success_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"application_success_feedback_sample_rate": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"application_failure_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"http_success_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"http_success_feedback_sample_rate": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"http_failure_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_master_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"fifo_topic": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"content_based_deduplication": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"lambda_success_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"lambda_success_feedback_sample_rate": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"lambda_failure_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sqs_success_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sqs_success_feedback_sample_rate": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 100),
			},
			"sqs_failure_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsSnsTopicCreate(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	var name string
	fifoTopic := d.Get("fifo_topic").(bool)

	if fifoTopic {
		name = naming.GenerateWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string), tfsns.FifoTopicNameSuffix)
	} else {
		name = naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))
	}

	attributes := make(map[string]*string)
	// If FifoTopic is true, then the attribute must be passed into the call to CreateTopic
	if fifoTopic {
		attributes["FifoTopic"] = aws.String(strconv.FormatBool(fifoTopic))
	}

	log.Printf("[DEBUG] SNS create topic: %s", name)

	req := &sns.CreateTopicInput{
		Name: aws.String(name),
		Tags: tags.IgnoreAws().SnsTags(),
	}

	if len(attributes) > 0 {
		req.Attributes = attributes
	}

	output, err := snsconn.CreateTopic(req)
	if err != nil {
		return fmt.Errorf("error creating SNS Topic (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.TopicArn))

	// update mutable attributes
	if d.HasChange("application_failure_feedback_role_arn") {
		_, v := d.GetChange("application_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ApplicationFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("application_success_feedback_role_arn") {
		_, v := d.GetChange("application_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ApplicationSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("arn") {
		_, v := d.GetChange("arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "TopicArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("delivery_policy") {
		_, v := d.GetChange("delivery_policy")
		if err := updateAwsSnsTopicAttribute(d.Id(), "DeliveryPolicy", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("display_name") {
		_, v := d.GetChange("display_name")
		if err := updateAwsSnsTopicAttribute(d.Id(), "DisplayName", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("http_failure_feedback_role_arn") {
		_, v := d.GetChange("http_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "HTTPFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("http_success_feedback_role_arn") {
		_, v := d.GetChange("http_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "HTTPSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("kms_master_key_id") {
		_, v := d.GetChange("kms_master_key_id")
		if err := updateAwsSnsTopicAttribute(d.Id(), "KmsMasterKeyId", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("content_based_deduplication") {
		_, v := d.GetChange("content_based_deduplication")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ContentBasedDeduplication", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_failure_feedback_role_arn") {
		_, v := d.GetChange("lambda_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "LambdaFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_success_feedback_role_arn") {
		_, v := d.GetChange("lambda_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "LambdaSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("policy") {
		_, v := d.GetChange("policy")
		if err := updateAwsSnsTopicAttribute(d.Id(), "Policy", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_failure_feedback_role_arn") {
		_, v := d.GetChange("sqs_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "SQSFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_success_feedback_role_arn") {
		_, v := d.GetChange("sqs_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "SQSSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("application_success_feedback_sample_rate") {
		_, v := d.GetChange("application_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ApplicationSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("http_success_feedback_sample_rate") {
		_, v := d.GetChange("http_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "HTTPSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_success_feedback_sample_rate") {
		_, v := d.GetChange("lambda_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "LambdaSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_success_feedback_sample_rate") {
		_, v := d.GetChange("sqs_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "SQSSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}

	return resourceAwsSnsTopicRead(d, meta)
}

func resourceAwsSnsTopicUpdate(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn

	// update mutable attributes
	if d.HasChange("application_failure_feedback_role_arn") {
		_, v := d.GetChange("application_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ApplicationFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("application_success_feedback_role_arn") {
		_, v := d.GetChange("application_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ApplicationSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("arn") {
		_, v := d.GetChange("arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "TopicArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("delivery_policy") {
		_, v := d.GetChange("delivery_policy")
		if err := updateAwsSnsTopicAttribute(d.Id(), "DeliveryPolicy", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("display_name") {
		_, v := d.GetChange("display_name")
		if err := updateAwsSnsTopicAttribute(d.Id(), "DisplayName", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("http_failure_feedback_role_arn") {
		_, v := d.GetChange("http_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "HTTPFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("http_success_feedback_role_arn") {
		_, v := d.GetChange("http_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "HTTPSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("kms_master_key_id") {
		_, v := d.GetChange("kms_master_key_id")
		if err := updateAwsSnsTopicAttribute(d.Id(), "KmsMasterKeyId", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("content_based_deduplication") {
		_, v := d.GetChange("content_based_deduplication")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ContentBasedDeduplication", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_failure_feedback_role_arn") {
		_, v := d.GetChange("lambda_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "LambdaFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_success_feedback_role_arn") {
		_, v := d.GetChange("lambda_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "LambdaSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("policy") {
		_, v := d.GetChange("policy")
		if err := updateAwsSnsTopicAttribute(d.Id(), "Policy", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_failure_feedback_role_arn") {
		_, v := d.GetChange("sqs_failure_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "SQSFailureFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_success_feedback_role_arn") {
		_, v := d.GetChange("sqs_success_feedback_role_arn")
		if err := updateAwsSnsTopicAttribute(d.Id(), "SQSSuccessFeedbackRoleArn", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("application_success_feedback_sample_rate") {
		_, v := d.GetChange("application_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "ApplicationSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("http_success_feedback_sample_rate") {
		_, v := d.GetChange("http_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "HTTPSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("lambda_success_feedback_sample_rate") {
		_, v := d.GetChange("lambda_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "LambdaSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}
	if d.HasChange("sqs_success_feedback_sample_rate") {
		_, v := d.GetChange("sqs_success_feedback_sample_rate")
		if err := updateAwsSnsTopicAttribute(d.Id(), "SQSSuccessFeedbackSampleRate", v, snsconn); err != nil {
			return err
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.SnsUpdateTags(snsconn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceAwsSnsTopicRead(d, meta)
}

func resourceAwsSnsTopicRead(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading SNS Topic Attributes for %s", d.Id())
	attributeOutput, err := snsconn.GetTopicAttributes(&sns.GetTopicAttributesInput{
		TopicArn: aws.String(d.Id()),
	})

	if isAWSErr(err, sns.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] SNS Topic (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SNS Topic (%s): %w", d.Id(), err)
	}

	fifoTopic := false

	// set the mutable attributes
	if attributeOutput.Attributes != nil && len(attributeOutput.Attributes) > 0 {
		// set the string values
		d.Set("application_failure_feedback_role_arn", attributeOutput.Attributes["ApplicationFailureFeedbackRoleArn"])
		d.Set("application_success_feedback_role_arn", attributeOutput.Attributes["ApplicationSuccessFeedbackRoleArn"])
		d.Set("arn", attributeOutput.Attributes["TopicArn"])
		d.Set("delivery_policy", attributeOutput.Attributes["DeliveryPolicy"])
		d.Set("display_name", attributeOutput.Attributes["DisplayName"])
		d.Set("http_failure_feedback_role_arn", attributeOutput.Attributes["HTTPFailureFeedbackRoleArn"])
		d.Set("http_success_feedback_role_arn", attributeOutput.Attributes["HTTPSuccessFeedbackRoleArn"])
		d.Set("kms_master_key_id", attributeOutput.Attributes["KmsMasterKeyId"])
		d.Set("lambda_failure_feedback_role_arn", attributeOutput.Attributes["LambdaFailureFeedbackRoleArn"])
		d.Set("lambda_success_feedback_role_arn", attributeOutput.Attributes["LambdaSuccessFeedbackRoleArn"])
		d.Set("policy", attributeOutput.Attributes["Policy"])
		d.Set("sqs_failure_feedback_role_arn", attributeOutput.Attributes["SQSFailureFeedbackRoleArn"])
		d.Set("sqs_success_feedback_role_arn", attributeOutput.Attributes["SQSSuccessFeedbackRoleArn"])

		// set the boolean values
		if v, ok := attributeOutput.Attributes["FifoTopic"]; ok && aws.StringValue(v) == "true" {
			fifoTopic = true
		}
		d.Set("content_based_deduplication", false)
		if v, ok := attributeOutput.Attributes["ContentBasedDeduplication"]; ok && aws.StringValue(v) == "true" {
			d.Set("content_based_deduplication", true)
		}

		// set the number values
		var vStr string
		var v int64
		var err error

		vStr = aws.StringValue(attributeOutput.Attributes["ApplicationSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'ApplicationSuccessFeedbackSampleRate': %w", err)
			}
			d.Set("application_success_feedback_sample_rate", v)
		}

		vStr = aws.StringValue(attributeOutput.Attributes["HTTPSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'HTTPSuccessFeedbackSampleRate': %w", err)
			}
			d.Set("http_success_feedback_sample_rate", v)
		}

		vStr = aws.StringValue(attributeOutput.Attributes["LambdaSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'LambdaSuccessFeedbackSampleRate': %w", err)
			}
			d.Set("lambda_success_feedback_sample_rate", v)
		}

		vStr = aws.StringValue(attributeOutput.Attributes["SQSSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'SQSSuccessFeedbackSampleRate': %w", err)
			}
			d.Set("sqs_success_feedback_sample_rate", v)
		}
	}

	d.Set("fifo_topic", fifoTopic)

	arn, err := arn.Parse(d.Id())

	if err != nil {
		return fmt.Errorf("error parsing ARN (%s): %w", d.Id(), err)
	}

	name := arn.Resource
	d.Set("name", name)
	if fifoTopic {
		d.Set("name_prefix", naming.NamePrefixFromNameWithSuffix(name, tfsns.FifoTopicNameSuffix))
	} else {
		d.Set("name_prefix", naming.NamePrefixFromName(name))
	}

	tags, err := keyvaluetags.SnsListTags(snsconn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for SNS Topic (%s): %w", d.Id(), err)
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

func resourceAwsSnsTopicDelete(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn

	log.Printf("[DEBUG] SNS Delete Topic: %s", d.Id())
	_, err := snsconn.DeleteTopic(&sns.DeleteTopicInput{
		TopicArn: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting SNS Topic (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAwsSnsTopicCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	fifoTopic := diff.Get("fifo_topic").(bool)
	contentBasedDeduplication := diff.Get("content_based_deduplication").(bool)

	if diff.Id() == "" {
		// Create.

		var name string

		if fifoTopic {
			name = naming.GenerateWithSuffix(diff.Get("name").(string), diff.Get("name_prefix").(string), tfsns.FifoTopicNameSuffix)
		} else {
			name = naming.Generate(diff.Get("name").(string), diff.Get("name_prefix").(string))
		}

		var re *regexp.Regexp

		if fifoTopic {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,251}\.fifo$`)
		} else {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,256}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid topic name: %s", name)
		}

	}

	if !fifoTopic && contentBasedDeduplication {
		return fmt.Errorf("content-based deduplication can only be set for FIFO topics")
	}

	return nil
}

func updateAwsSnsTopicAttribute(topicArn, name string, value interface{}, conn *sns.SNS) error {
	// Ignore an empty policy
	if name == "Policy" && value == "" {
		return nil
	}
	log.Printf("[DEBUG] Updating SNS Topic Attribute: %s", name)

	// Make API call to update attributes
	req := sns.SetTopicAttributesInput{
		TopicArn:       aws.String(topicArn),
		AttributeName:  aws.String(name),
		AttributeValue: aws.String(fmt.Sprintf("%v", value)),
	}

	// Retry the update in the event of an eventually consistent style of
	// error, where say an IAM resource is successfully created but not
	// actually available. See https://github.com/hashicorp/terraform/issues/3660
	_, err := retryOnAwsCode(sns.ErrCodeInvalidParameterException, func() (interface{}, error) {
		return conn.SetTopicAttributes(&req)
	})

	if err != nil {
		return fmt.Errorf("error setting SNS Topic (%s) attributes: %w", topicArn, err)
	}

	return nil
}
