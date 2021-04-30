package aws

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
	tfsqs "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sqs"
)

var sqsQueueAttributeMap = map[string]string{
	"delay_seconds":                     sqs.QueueAttributeNameDelaySeconds,
	"max_message_size":                  sqs.QueueAttributeNameMaximumMessageSize,
	"message_retention_seconds":         sqs.QueueAttributeNameMessageRetentionPeriod,
	"receive_wait_time_seconds":         sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds,
	"visibility_timeout_seconds":        sqs.QueueAttributeNameVisibilityTimeout,
	"policy":                            sqs.QueueAttributeNamePolicy,
	"redrive_policy":                    sqs.QueueAttributeNameRedrivePolicy,
	"arn":                               sqs.QueueAttributeNameQueueArn,
	"fifo_queue":                        sqs.QueueAttributeNameFifoQueue,
	"content_based_deduplication":       sqs.QueueAttributeNameContentBasedDeduplication,
	"kms_master_key_id":                 sqs.QueueAttributeNameKmsMasterKeyId,
	"kms_data_key_reuse_period_seconds": sqs.QueueAttributeNameKmsDataKeyReusePeriodSeconds,
}

// A number of these are marked as computed because if you don't
// provide a value, SQS will provide you with defaults (which are the
// default values specified below)
func resourceAwsSqsQueue() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSqsQueueCreate,
		Read:   resourceAwsSqsQueueRead,
		Update: resourceAwsSqsQueueUpdate,
		Delete: resourceAwsSqsQueueDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			resourceAwsSqsQueueCustomizeDiff,
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
			"delay_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"max_message_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  262144,
			},
			"message_retention_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  345600,
			},
			"receive_wait_time_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"visibility_timeout_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
			},
			"policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},
			"redrive_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"fifo_queue": {
				Type:     schema.TypeBool,
				Default:  false,
				ForceNew: true,
				Optional: true,
			},
			"content_based_deduplication": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"kms_master_key_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kms_data_key_reuse_period_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsSqsQueueCreate(d *schema.ResourceData, meta interface{}) error {
	sqsconn := meta.(*AWSClient).sqsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	var name string
	fifoQueue := d.Get("fifo_queue").(bool)

	if fifoQueue {
		name = naming.GenerateWithSuffix(d.Get("name").(string), d.Get("name_prefix").(string), tfsqs.FifoQueueNameSuffix)
	} else {
		name = naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))
	}

	log.Printf("[DEBUG] SQS queue create: %s", name)

	req := &sqs.CreateQueueInput{
		QueueName: aws.String(name),
	}

	// Tag-on-create is currently only supported in AWS Commercial
	if len(tags) > 0 && meta.(*AWSClient).partition == endpoints.AwsPartitionID {
		req.Tags = tags.IgnoreAws().SqsTags()
	}

	attributes := make(map[string]*string)

	queueResource := *resourceAwsSqsQueue()

	for k, s := range queueResource.Schema {
		if attrKey, ok := sqsQueueAttributeMap[k]; ok {
			if value, ok := d.GetOk(k); ok {
				switch s.Type {
				case schema.TypeInt:
					attributes[attrKey] = aws.String(strconv.Itoa(value.(int)))
				case schema.TypeBool:
					attributes[attrKey] = aws.String(strconv.FormatBool(value.(bool)))
				default:
					attributes[attrKey] = aws.String(value.(string))
				}
			}

		}
	}

	if len(attributes) > 0 {
		req.Attributes = attributes
	}

	var output *sqs.CreateQueueOutput
	err := resource.Retry(70*time.Second, func() *resource.RetryError {
		var err error
		output, err = sqsconn.CreateQueue(req)
		if err != nil {
			if isAWSErr(err, sqs.ErrCodeQueueDeletedRecently, "You must wait 60 seconds after deleting a queue before you can create another with the same name.") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		output, err = sqsconn.CreateQueue(req)
	}
	if err != nil {
		return fmt.Errorf("Error creating SQS queue: %s", err)
	}

	d.SetId(aws.StringValue(output.QueueUrl))

	// Tag-on-create is currently only supported in AWS Commercial
	if meta.(*AWSClient).partition == endpoints.AwsPartitionID {
		return resourceAwsSqsQueueRead(d, meta)
	} else {
		return resourceAwsSqsQueueUpdate(d, meta)
	}
}

func resourceAwsSqsQueueUpdate(d *schema.ResourceData, meta interface{}) error {
	sqsconn := meta.(*AWSClient).sqsconn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.SqsUpdateTags(sqsconn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating SQS Queue (%s) tags: %s", d.Id(), err)
		}
	}

	attributes := make(map[string]*string)

	resource := *resourceAwsSqsQueue()

	for k, s := range resource.Schema {
		if attrKey, ok := sqsQueueAttributeMap[k]; ok {
			if d.HasChange(k) {
				log.Printf("[DEBUG] Updating %s", attrKey)
				_, n := d.GetChange(k)
				switch s.Type {
				case schema.TypeInt:
					attributes[attrKey] = aws.String(strconv.Itoa(n.(int)))
				case schema.TypeBool:
					attributes[attrKey] = aws.String(strconv.FormatBool(n.(bool)))
				default:
					attributes[attrKey] = aws.String(n.(string))
				}
			}
		}
	}

	if len(attributes) > 0 {
		req := &sqs.SetQueueAttributesInput{
			QueueUrl:   aws.String(d.Id()),
			Attributes: attributes,
		}
		if _, err := sqsconn.SetQueueAttributes(req); err != nil {
			return fmt.Errorf("Error updating SQS attributes: %s", err)
		}
	}

	return resourceAwsSqsQueueRead(d, meta)
}

func resourceAwsSqsQueueRead(d *schema.ResourceData, meta interface{}) error {
	sqsconn := meta.(*AWSClient).sqsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	attributeOutput, err := sqsconn.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(d.Id()),
		AttributeNames: []*string{aws.String("All")},
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			log.Printf("ERROR Found %s", awsErr.Code())
			if awsErr.Code() == sqs.ErrCodeQueueDoesNotExist {
				d.SetId("")
				log.Printf("[DEBUG] SQS Queue (%s) not found", d.Get("name").(string))
				return nil
			}
		}
		return err
	}

	name, err := extractNameFromSqsQueueUrl(d.Id())
	if err != nil {
		return err
	}

	fifoQueue := false

	// Always set attribute defaults
	d.Set("arn", "")
	d.Set("content_based_deduplication", false)
	d.Set("delay_seconds", 0)
	d.Set("kms_data_key_reuse_period_seconds", 300)
	d.Set("kms_master_key_id", "")
	d.Set("max_message_size", 262144)
	d.Set("message_retention_seconds", 345600)
	d.Set("policy", "")
	d.Set("receive_wait_time_seconds", 0)
	d.Set("redrive_policy", "")
	d.Set("visibility_timeout_seconds", 30)

	if attributeOutput != nil {
		queueAttributes := aws.StringValueMap(attributeOutput.Attributes)

		if v, ok := queueAttributes[sqs.QueueAttributeNameQueueArn]; ok {
			d.Set("arn", v)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameContentBasedDeduplication]; ok && v != "" {
			vBool, err := strconv.ParseBool(v)

			if err != nil {
				return fmt.Errorf("error parsing content_based_deduplication value (%s) into boolean: %s", v, err)
			}

			d.Set("content_based_deduplication", vBool)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameDelaySeconds]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("error parsing delay_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("delay_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameFifoQueue]; ok && v != "" {
			vBool, err := strconv.ParseBool(v)

			if err != nil {
				return fmt.Errorf("error parsing fifo_queue value (%s) into boolean: %s", v, err)
			}

			fifoQueue = vBool
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameKmsDataKeyReusePeriodSeconds]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("error parsing kms_data_key_reuse_period_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("kms_data_key_reuse_period_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameKmsMasterKeyId]; ok {
			d.Set("kms_master_key_id", v)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameMaximumMessageSize]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("error parsing max_message_size value (%s) into integer: %s", v, err)
			}

			d.Set("max_message_size", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameMessageRetentionPeriod]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("error parsing message_retention_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("message_retention_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNamePolicy]; ok {
			d.Set("policy", v)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("error parsing receive_wait_time_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("receive_wait_time_seconds", vInt)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameRedrivePolicy]; ok {
			d.Set("redrive_policy", v)
		}

		if v, ok := queueAttributes[sqs.QueueAttributeNameVisibilityTimeout]; ok && v != "" {
			vInt, err := strconv.Atoi(v)

			if err != nil {
				return fmt.Errorf("error parsing visibility_timeout_seconds value (%s) into integer: %s", v, err)
			}

			d.Set("visibility_timeout_seconds", vInt)
		}
	}

	d.Set("fifo_queue", fifoQueue)
	d.Set("name", name)
	if fifoQueue {
		d.Set("name_prefix", naming.NamePrefixFromNameWithSuffix(name, tfsqs.FifoQueueNameSuffix))
	} else {
		d.Set("name_prefix", naming.NamePrefixFromName(name))
	}

	tags, err := keyvaluetags.SqsListTags(sqsconn, d.Id())

	if err != nil {
		// Non-standard partitions (e.g. US Gov) and some local development
		// solutions do not yet support this API call. Depending on the
		// implementation it may return InvalidAction or AWS.SimpleQueueService.UnsupportedOperation
		if !isAWSErr(err, "InvalidAction", "") && !isAWSErr(err, sqs.ErrCodeUnsupportedOperation, "") {
			return fmt.Errorf("error listing tags for SQS Queue (%s): %s", d.Id(), err)
		}
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

func resourceAwsSqsQueueDelete(d *schema.ResourceData, meta interface{}) error {
	sqsconn := meta.(*AWSClient).sqsconn

	log.Printf("[DEBUG] SQS Delete Queue: %s", d.Id())
	_, err := sqsconn.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: aws.String(d.Id()),
	})
	return err
}

func resourceAwsSqsQueueCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	fifoQueue := diff.Get("fifo_queue").(bool)
	contentBasedDeduplication := diff.Get("content_based_deduplication").(bool)

	if diff.Id() == "" {
		// Create.

		var name string

		if fifoQueue {
			name = naming.GenerateWithSuffix(diff.Get("name").(string), diff.Get("name_prefix").(string), tfsqs.FifoQueueNameSuffix)
		} else {
			name = naming.Generate(diff.Get("name").(string), diff.Get("name_prefix").(string))
		}

		var re *regexp.Regexp

		if fifoQueue {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,75}\.fifo$`)
		} else {
			re = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)
		}

		if !re.MatchString(name) {
			return fmt.Errorf("invalid queue name: %s", name)
		}

	}

	if !fifoQueue && contentBasedDeduplication {
		return fmt.Errorf("content-based deduplication can only be set for FIFO queue")
	}

	return nil
}

func extractNameFromSqsQueueUrl(queue string) (string, error) {
	//http://sqs.us-west-2.amazonaws.com/123456789012/queueName
	u, err := url.Parse(queue)
	if err != nil {
		return "", err
	}
	segments := strings.Split(u.Path, "/")
	if len(segments) != 3 {
		return "", fmt.Errorf("SQS Url not parsed correctly")
	}

	return segments[2], nil

}
