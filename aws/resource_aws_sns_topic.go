package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
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
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSnsTopicCreate(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn
	tags := keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().SnsTags()
	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	log.Printf("[DEBUG] SNS create topic: %s", name)

	req := &sns.CreateTopicInput{
		Name: aws.String(name),
		Tags: tags,
	}

	output, err := snsconn.CreateTopic(req)
	if err != nil {
		return fmt.Errorf("Error creating SNS topic: %s", err)
	}

	d.SetId(*output.TopicArn)

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

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SnsUpdateTags(snsconn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSnsTopicRead(d, meta)
}

func resourceAwsSnsTopicRead(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading SNS Topic Attributes for %s", d.Id())
	attributeOutput, err := snsconn.GetTopicAttributes(&sns.GetTopicAttributesInput{
		TopicArn: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, sns.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] SNS Topic (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	// set the mutable attributes
	if attributeOutput.Attributes != nil && len(attributeOutput.Attributes) > 0 {
		// set the string values
		d.Set("application_failure_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["ApplicationFailureFeedbackRoleArn"]))
		d.Set("application_success_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["ApplicationSuccessFeedbackRoleArn"]))
		d.Set("arn", aws.StringValue(attributeOutput.Attributes["TopicArn"]))
		d.Set("delivery_policy", aws.StringValue(attributeOutput.Attributes["DeliveryPolicy"]))
		d.Set("display_name", aws.StringValue(attributeOutput.Attributes["DisplayName"]))
		d.Set("http_failure_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["HTTPFailureFeedbackRoleArn"]))
		d.Set("http_success_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["HTTPSuccessFeedbackRoleArn"]))
		d.Set("kms_master_key_id", aws.StringValue(attributeOutput.Attributes["KmsMasterKeyId"]))
		d.Set("lambda_failure_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["LambdaFailureFeedbackRoleArn"]))
		d.Set("lambda_success_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["LambdaSuccessFeedbackRoleArn"]))
		d.Set("policy", aws.StringValue(attributeOutput.Attributes["Policy"]))
		d.Set("sqs_failure_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["SQSFailureFeedbackRoleArn"]))
		d.Set("sqs_success_feedback_role_arn", aws.StringValue(attributeOutput.Attributes["SQSSuccessFeedbackRoleArn"]))

		// set the number values
		var vStr string
		var v int64
		var err error

		vStr = aws.StringValue(attributeOutput.Attributes["ApplicationSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'ApplicationSuccessFeedbackSampleRate': %s", err)
			}
			d.Set("application_success_feedback_sample_rate", v)
		}

		vStr = aws.StringValue(attributeOutput.Attributes["HTTPSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'HTTPSuccessFeedbackSampleRate': %s", err)
			}
			d.Set("http_success_feedback_sample_rate", v)
		}

		vStr = aws.StringValue(attributeOutput.Attributes["LambdaSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'LambdaSuccessFeedbackSampleRate': %s", err)
			}
			d.Set("lambda_success_feedback_sample_rate", v)
		}

		vStr = aws.StringValue(attributeOutput.Attributes["SQSSuccessFeedbackSampleRate"])
		if vStr != "" {
			v, err = strconv.ParseInt(vStr, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing integer attribute 'SQSSuccessFeedbackSampleRate': %s", err)
			}
			d.Set("sqs_success_feedback_sample_rate", v)
		}
	}

	// If we have no name set (import) then determine it from the ARN.
	// This is a bit of a heuristic for now since AWS provides no other
	// way to get it.
	if _, ok := d.GetOk("name"); !ok {
		arn := d.Get("arn").(string)
		idx := strings.LastIndex(arn, ":")
		if idx > -1 {
			d.Set("name", arn[idx+1:])
		}
	}

	tags, err := keyvaluetags.SnsListTags(snsconn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSnsTopicDelete(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn

	log.Printf("[DEBUG] SNS Delete Topic: %s", d.Id())
	_, err := snsconn.DeleteTopic(&sns.DeleteTopicInput{
		TopicArn: aws.String(d.Id()),
	})

	return err
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

	return err
}
