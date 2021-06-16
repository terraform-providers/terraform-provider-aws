package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesis/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesis/waiter"
)

func resourceAwsKinesisStreamConsumer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsKinesisStreamConsumerCreate,
		Read:   resourceAwsKinesisStreamConsumerRead,
		Delete: resourceAwsKinesisStreamConsumerDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"creation_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"stream_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsKinesisStreamConsumerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn

	name := d.Get("name").(string)
	streamArn := d.Get("stream_arn").(string)

	input := &kinesis.RegisterStreamConsumerInput{
		ConsumerName: aws.String(name),
		StreamARN:    aws.String(streamArn),
	}

	output, err := conn.RegisterStreamConsumer(input)
	if err != nil {
		return fmt.Errorf("error creating Kinesis Stream Consumer (%s): %w", name, err)
	}

	if output == nil || output.Consumer == nil {
		return fmt.Errorf("error creating Kinesis Stream Consumer (%s): empty output", name)
	}

	d.SetId(aws.StringValue(output.Consumer.ConsumerARN))

	if _, err := waiter.StreamConsumerCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Kinesis Stream Consumer (%s) creation: %w", d.Id(), err)
	}

	return resourceAwsKinesisStreamConsumerRead(d, meta)
}

func resourceAwsKinesisStreamConsumerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn

	consumer, err := finder.StreamConsumerByARN(conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Kinesis Stream Consumer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Kinesis Stream Consumer (%s): %w", d.Id(), err)
	}

	if consumer == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading Kinesis Stream Consumer (%s): empty output after creation", d.Id())
		}
		log.Printf("[WARN] Kinesis Stream Consumer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", consumer.ConsumerARN)
	d.Set("name", consumer.ConsumerName)
	d.Set("creation_timestamp", aws.TimeValue(consumer.ConsumerCreationTimestamp).Format(time.RFC3339))
	d.Set("stream_arn", consumer.StreamARN)

	return nil
}

func resourceAwsKinesisStreamConsumerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kinesisconn

	input := &kinesis.DeregisterStreamConsumerInput{
		ConsumerARN: aws.String(d.Id()),
	}

	_, err := conn.DeregisterStreamConsumer(input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Kinesis Stream Consumer (%s): %w", d.Id(), err)
	}

	if _, err := waiter.StreamConsumerDeleted(conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error waiting for Kinesis Stream Consumer (%s) deletion: %w", d.Id(), err)
	}

	return nil
}
