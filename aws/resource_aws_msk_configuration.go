package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kafka/waiter"
)

func resourceAwsMskConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMskConfigurationCreate,
		Read:   resourceAwsMskConfigurationRead,
		Update: resourceAwsMskConfigurationUpdate,
		Delete: resourceAwsMskConfigurationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"kafka_versions": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"server_properties": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsMskConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	input := &kafka.CreateConfigurationInput{
		KafkaVersions:    expandStringSet(d.Get("kafka_versions").(*schema.Set)),
		Name:             aws.String(d.Get("name").(string)),
		ServerProperties: []byte(d.Get("server_properties").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateConfiguration(input)

	if err != nil {
		return fmt.Errorf("error creating MSK Configuration: %s", err)
	}

	d.SetId(aws.StringValue(output.Arn))

	return resourceAwsMskConfigurationRead(d, meta)
}

func resourceAwsMskConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	configurationInput := &kafka.DescribeConfigurationInput{
		Arn: aws.String(d.Id()),
	}

	configurationOutput, err := conn.DescribeConfiguration(configurationInput)

	if isAWSErr(err, kafka.ErrCodeBadRequestException, "Configuration ARN does not exist") {
		log.Printf("[WARN] MSK Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing MSK Configuration (%s): %s", d.Id(), err)
	}

	if configurationOutput == nil {
		return fmt.Errorf("error describing MSK Configuration (%s): missing result", d.Id())
	}

	if configurationOutput.LatestRevision == nil {
		return fmt.Errorf("error describing MSK Configuration (%s): missing latest revision", d.Id())
	}

	revision := configurationOutput.LatestRevision.Revision
	revisionInput := &kafka.DescribeConfigurationRevisionInput{
		Arn:      aws.String(d.Id()),
		Revision: revision,
	}

	revisionOutput, err := conn.DescribeConfigurationRevision(revisionInput)

	if err != nil {
		return fmt.Errorf("error describing MSK Configuration (%s) Revision (%d): %s", d.Id(), aws.Int64Value(revision), err)
	}

	if revisionOutput == nil {
		return fmt.Errorf("error describing MSK Configuration (%s) Revision (%d): missing result", d.Id(), aws.Int64Value(revision))
	}

	d.Set("arn", aws.StringValue(configurationOutput.Arn))
	d.Set("description", aws.StringValue(revisionOutput.Description))

	if err := d.Set("kafka_versions", aws.StringValueSlice(configurationOutput.KafkaVersions)); err != nil {
		return fmt.Errorf("error setting kafka_versions: %s", err)
	}

	d.Set("latest_revision", aws.Int64Value(revision))
	d.Set("name", aws.StringValue(configurationOutput.Name))
	d.Set("server_properties", string(revisionOutput.ServerProperties))

	return nil
}

func resourceAwsMskConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	input := &kafka.UpdateConfigurationInput{
		Arn:              aws.String(d.Id()),
		ServerProperties: []byte(d.Get("server_properties").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.UpdateConfiguration(input)

	if err != nil {
		return fmt.Errorf("error updating MSK Configuration (%s): %w", d.Id(), err)
	}

	return resourceAwsMskConfigurationRead(d, meta)
}

func resourceAwsMskConfigurationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).kafkaconn

	input := &kafka.DeleteConfigurationInput{
		Arn: aws.String(d.Id()),
	}

	_, err := conn.DeleteConfiguration(input)

	if err != nil {
		return fmt.Errorf("error deleting MSK Configuration (%s): %w", d.Id(), err)
	}

	if _, err := waiter.ConfigurationDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for MSK Configuration (%s): %w", d.Id(), err)
	}

	return nil
}
