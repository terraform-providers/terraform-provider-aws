package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSfnActivity() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSfnActivityCreate,
		Read:   resourceAwsSfnActivityRead,
		Update: resourceAwsSfnActivityUpdate,
		Delete: resourceAwsSfnActivityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 80),
			},

			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSfnActivityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	log.Print("[DEBUG] Creating Step Function Activity")

	params := &sfn.CreateActivityInput{
		Name: aws.String(d.Get("name").(string)),
		Tags: keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().SfnTags(),
	}

	activity, err := conn.CreateActivity(params)
	if err != nil {
		return fmt.Errorf("Error creating Step Function Activity: %s", err)
	}

	d.SetId(*activity.ActivityArn)

	return resourceAwsSfnActivityRead(d, meta)
}

func resourceAwsSfnActivityUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.SfnUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsSfnActivityRead(d, meta)
}

func resourceAwsSfnActivityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading Step Function Activity: %s", d.Id())

	sm, err := conn.DescribeActivity(&sfn.DescribeActivityInput{
		ActivityArn: aws.String(d.Id()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ActivityDoesNotExist" {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", sm.Name)

	if err := d.Set("creation_date", sm.CreationDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting creation_date: %s", err)
	}

	tags, err := keyvaluetags.SfnListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for SFN Activity (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsSfnActivityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sfnconn
	log.Printf("[DEBUG] Deleting Step Functions Activity: %s", d.Id())

	input := &sfn.DeleteActivityInput{
		ActivityArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteActivity(input)

	if err != nil {
		return fmt.Errorf("Error deleting SFN Activity: %s", err)
	}

	return nil
}
