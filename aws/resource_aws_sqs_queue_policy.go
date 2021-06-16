package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sqs/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sqs/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

var (
	sqsQueueEmptyPolicyAttributes = map[string]string{
		sqs.QueueAttributeNamePolicy: "",
	}
)

func resourceAwsSqsQueuePolicy() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAwsSqsQueuePolicyUpsert,
		Read:   resourceAwsSqsQueuePolicyRead,
		Update: resourceAwsSqsQueuePolicyUpsert,
		Delete: resourceAwsSqsQueuePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		MigrateState:  resourceAwsSqsQueuePolicyMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"queue_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsSqsQueuePolicyUpsert(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sqsconn

	policyAttributes := map[string]string{
		sqs.QueueAttributeNamePolicy: d.Get("policy").(string),
	}
	url := d.Get("queue_url").(string)
	input := &sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(policyAttributes),
		QueueUrl:   aws.String(url),
	}

	log.Printf("[DEBUG] Setting SQS Queue Policy: %s", input)
	_, err := conn.SetQueueAttributes(input)

	if err != nil {
		return fmt.Errorf("error setting SQS Queue Policy (%s): %w", url, err)
	}

	d.SetId(url)

	err = waiter.QueueAttributesPropagated(conn, d.Id(), policyAttributes)

	if err != nil {
		return fmt.Errorf("error waiting for SQS Queue Policy (%s) to be set: %w", d.Id(), err)
	}

	return resourceAwsSqsQueuePolicyRead(d, meta)
}

func resourceAwsSqsQueuePolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sqsconn

	policy, err := finder.QueuePolicyByURL(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SQS Queue Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading SQS Queue Policy (%s): %w", d.Id(), err)
	}

	d.Set("policy", policy)
	d.Set("queue_url", d.Id())

	return nil
}

func resourceAwsSqsQueuePolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sqsconn

	log.Printf("[DEBUG] Deleting SQS Queue Policy: %s", d.Id())
	_, err := conn.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		Attributes: aws.StringMap(sqsQueueEmptyPolicyAttributes),
		QueueUrl:   aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting SQS Queue Policy (%s): %w", d.Id(), err)
	}

	err = waiter.QueueAttributesPropagated(conn, d.Id(), sqsQueueEmptyPolicyAttributes)

	if err != nil {
		return fmt.Errorf("error waiting for SQS Queue Policy (%s) to delete: %w", d.Id(), err)
	}

	return nil
}
