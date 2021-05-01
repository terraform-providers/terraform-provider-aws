package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsCloudWatchEventArchive() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCloudWatchEventArchiveCreate,
		Read:   resourceAwsCloudWatchEventArchiveRead,
		Update: resourceAwsCloudWatchEventArchiveUpdate,
		Delete: resourceAwsCloudWatchEventArchiveDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateCloudWatchEventArchiveName,
			},
			"event_source_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"event_pattern": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateEventPatternValue(),
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v.(string))
					return json
				},
			},
			"retention_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsCloudWatchEventArchiveCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input, err := buildCreateArchiveInputStruct(d)

	if err != nil {
		return fmt.Errorf("Creating CloudWatch Events Archive parameters failed: %w", err)
	}

	log.Printf("[DEBUG] Creating CloudWatch Events Archive: %s", input)

	_, err = conn.CreateArchive(input)
	if err != nil {
		return fmt.Errorf("Creating CloudWatch Events Archive failed: %w", err)
	}

	d.SetId(d.Get("name").(string))

	log.Printf("[INFO] CloudWatch Events Archive (%s) created", d.Id())

	return resourceAwsCloudWatchEventArchiveRead(d, meta)
}

func resourceAwsCloudWatchEventArchiveRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn
	input := &events.DescribeArchiveInput{
		ArchiveName: aws.String(d.Id()),
	}

	out, err := conn.DescribeArchive(input)

	if isAWSErr(err, events.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] CloudWatch Events archive (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error reading CloudWatch Events archive: %w", err)
	}

	log.Printf("[DEBUG] Found Archive: #{*out}")

	d.Set("name", out.ArchiveName)
	d.Set("description", out.Description)
	d.Set("event_pattern", out.EventPattern)
	d.Set("event_source_arn", out.EventSourceArn)
	d.Set("arn", out.ArchiveArn)
	d.Set("retention_days", out.RetentionDays)

	return nil
}

func resourceAwsCloudWatchEventArchiveUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input, err := buildUpdateArchiveInputStruct(d)

	if err != nil {
		return fmt.Errorf("Creating CloudWatch Events Archive parameters failed: %w", err)
	}

	log.Printf("[DEBUG] Updating CloudWatch Events Archive: %s", input)
	_, err = conn.UpdateArchive(input)
	if err != nil {
		return fmt.Errorf("error updating CloudWatch Events Archive (%s): %w", d.Id(), err)
	}

	return resourceAwsCloudWatchEventArchiveRead(d, meta)
}

func resourceAwsCloudWatchEventArchiveDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cloudwatcheventsconn

	input := &events.DeleteArchiveInput{
		ArchiveName: aws.String(d.Get("name").(string)),
	}

	_, err := conn.DeleteArchive(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, events.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting CloudWatch Events Archive (%s): %w", d.Id(), err)
	}

	return nil
}

func buildCreateArchiveInputStruct(d *schema.ResourceData) (*events.CreateArchiveInput, error) {
	input := events.CreateArchiveInput{
		ArchiveName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		pattern, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		input.EventPattern = aws.String(pattern)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("event_source_arn"); ok {
		input.EventSourceArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retention_days"); ok {
		input.RetentionDays = aws.Int64(int64(v.(int)))
	}

	return &input, nil
}

func buildUpdateArchiveInputStruct(d *schema.ResourceData) (*events.UpdateArchiveInput, error) {
	input := events.UpdateArchiveInput{
		ArchiveName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("event_pattern"); ok {
		pattern, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("event pattern contains an invalid JSON: %w", err)
		}
		input.EventPattern = aws.String(pattern)
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("retention_days"); ok {
		input.RetentionDays = aws.Int64(int64(v.(int)))
	}

	return &input, nil
}
