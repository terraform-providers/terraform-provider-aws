package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsCloudWatchEventBus() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchEventBusCreate,
		Read:   resourceAwsCloudWatchEventBusRead,
		Update: resourceAwsCloudWatchEventBusUpdate,
		Delete: resourceAwsCloudWatchEventBusDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventCustomEventBusName,
			},
			"event_source_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventCustomEventBusEventSourceName,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsCloudWatchEventBusCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	eventBusName := d.Get("name").(string)
	input := &events.CreateEventBusInput{
		Name: aws.String(eventBusName),
	}

	if v, ok := d.GetOk("event_source_name"); ok {
		input.EventSourceName = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().CloudwatcheventsTags()
	}

	log.Printf("[DEBUG] Creating CloudWatch Events event bus: %v", input)

	_, err := conn.CreateEventBus(input)
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Events event bus (%s) failed: %w", eventBusName, err)
	}

	d.SetId(eventBusName)

	log.Printf("[INFO] CloudWatch Events event bus (%s) created", d.Id())

	return resourceAwsCloudWatchEventBusRead(d, meta)
}

func resourceAwsCloudWatchEventBusRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &events.DescribeEventBusInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading CloudWatch Events event bus (%s)", d.Id())
	output, err := conn.DescribeEventBus(input)
	if isAWSErr(err, events.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] CloudWatch Events event bus (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading CloudWatch Events event bus: %w", err)
	}

	log.Printf("[DEBUG] Found CloudWatch Event bus: %#v", *output)

	d.Set("arn", output.Arn)
	d.Set("name", output.Name)

	tags, err := keyvaluetags.CloudwatcheventsListTags(conn, aws.StringValue(output.Arn))
	if err != nil {
		return fmt.Errorf("error listing tags for CloudWatch Events event bus (%s): %w", d.Id(), err)
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

func resourceAwsCloudWatchEventBusUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := keyvaluetags.CloudwatcheventsUpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating CloudwWatch Events event bus (%s) tags: %w", arn, err)
		}
	}

	return resourceAwsCloudWatchEventBusRead(d, meta)
}

func resourceAwsCloudWatchEventBusDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	log.Printf("[INFO] Deleting CloudWatch Events event bus (%s)", d.Id())
	_, err := conn.DeleteEventBus(&events.DeleteEventBusInput{
		Name: aws.String(d.Id()),
	})
	if isAWSErr(err, events.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] CloudWatch Events event bus (%s) not found", d.Id())
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting CloudWatch Events event bus (%s): %w", d.Id(), err)
	}
	log.Printf("[INFO] CloudWatch Events event bus (%s) deleted", d.Id())

	return nil
}
