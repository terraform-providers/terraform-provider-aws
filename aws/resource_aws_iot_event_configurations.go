package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsIotEventConfigurations() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotEventConfigurationsUpdate,
		Read:   resourceAwsIotEventConfigurationsRead,
		Update: resourceAwsIotEventConfigurationsUpdate,
		Delete: resourceAwsIotEventConfigurationsCleanAll,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"values": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeBool,
			},
		},
	}
}

func resourceAwsIotEventConfigurationsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	params := &iot.DescribeEventConfigurationsInput{}
	out, err := conn.DescribeEventConfigurations(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Update IoT Event Configuration: %s", out)

	eventConfigurations := out.EventConfigurations
	rawValues := make(map[string]interface{})
	for key, value := range eventConfigurations {
		rawValues[key] = aws.BoolValue(value.Enabled)
	}

	d.Set("values", rawValues)

	return nil
}

func resourceAwsIotEventConfigurationsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	rawValues := d.Get("values").(map[string]interface{})
	eventConfigurations := make(map[string]*iot.Configuration)
	for key, enabled := range rawValues {
		eventConfigurations[key] = &iot.Configuration{
			Enabled: aws.Bool(enabled.(bool)),
		}
	}

	params := &iot.UpdateEventConfigurationsInput{
		EventConfigurations: eventConfigurations,
	}

	out, err := conn.UpdateEventConfigurations(params)

	if err != nil {
		return err
	}

	d.SetId("event-configurations")
	log.Printf("[DEBUG] Update IoT Event Configuration: %s", out)
	return resourceAwsIotEventConfigurationsRead(d, meta)
}

func resourceAwsIotEventConfigurationsCleanAll(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	params := &iot.DescribeEventConfigurationsInput{}
	out, err := conn.DescribeEventConfigurations(params)

	if err != nil {
		return err
	}

	eventConfigurations := out.EventConfigurations
	cleanedEventConfigurations := make(map[string]*iot.Configuration)
	for key := range eventConfigurations {
		cleanedEventConfigurations[key] = &iot.Configuration{
			Enabled: aws.Bool(false),
		}
	}

	updateParams := &iot.UpdateEventConfigurationsInput{
		EventConfigurations: cleanedEventConfigurations,
	}

	_, err = conn.UpdateEventConfigurations(updateParams)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Clean IoT Event Configuration: %s", out)

	return nil
}
