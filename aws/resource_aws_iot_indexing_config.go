package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsIotIndexingConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotIndexingConfigUpdate,
		Read:   resourceAwsIotIndexingConfigRead,
		Update: resourceAwsIotIndexingConfigUpdate,
		Delete: resourceAwsIotIndexingConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"thing_group_indexing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"thing_connectivity_indexing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"thing_indexing_mode": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					iot.ThingIndexingModeOff,
					iot.ThingIndexingModeRegistry,
					iot.ThingIndexingModeRegistryAndShadow,
				}, false),
				Default: iot.ThingIndexingModeOff,
			},
		},
	}
}

func resourceAwsIotIndexingConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	log.Printf("[DEBUG] Retrieving IoT indexing configuration")
	out, err := conn.GetIndexingConfiguration(&iot.GetIndexingConfigurationInput{})
	if err != nil {
		return fmt.Errorf("error retrieving IoT indexing configuration: %v", err)
	}
	log.Printf("[DEBUG] Retrieved IoT indexing configuration: %v", out)

	d.Set("thing_group_indexing_enabled",
		*out.ThingGroupIndexingConfiguration.ThingGroupIndexingMode == iot.ThingGroupIndexingModeOn)
	d.Set("thing_connectivity_indexing_enabled",
		*out.ThingIndexingConfiguration.ThingConnectivityIndexingMode == iot.ThingConnectivityIndexingModeStatus)
	d.Set("thing_indexing_mode", out.ThingIndexingConfiguration.ThingIndexingMode)

	return nil
}

func resourceAwsIotIndexingConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	input := iot.UpdateIndexingConfigurationInput{
		ThingGroupIndexingConfiguration: &iot.ThingGroupIndexingConfiguration{
			ThingGroupIndexingMode: aws.String(iot.ThingGroupIndexingModeOff),
		},
		ThingIndexingConfiguration: &iot.ThingIndexingConfiguration{
			ThingConnectivityIndexingMode: aws.String(iot.ThingConnectivityIndexingModeOff),
			ThingIndexingMode:             aws.String(d.Get("thing_indexing_mode").(string)),
		},
	}

	if d.Get("thing_group_indexing_enabled").(bool) {
		input.ThingGroupIndexingConfiguration.ThingGroupIndexingMode =
			aws.String(iot.ThingGroupIndexingModeOn)
	}
	if d.Get("thing_connectivity_indexing_enabled").(bool) {
		input.ThingIndexingConfiguration.ThingConnectivityIndexingMode =
			aws.String(iot.ThingConnectivityIndexingModeStatus)
	}

	log.Printf("[DEBUG] Updating IoT indexing configuration")
	_, err := conn.UpdateIndexingConfiguration(&input)
	if err != nil {
		return fmt.Errorf("error updating IoT indexing configuration: %v", err)
	}

	log.Printf("[DEBUG] Successfully updated IoT indexing configuration")

	d.SetId("iot-indexing-config")
	return resourceAwsIotIndexingConfigRead(d, meta)
}

func resourceAwsIotIndexingConfigDelete(_ *schema.ResourceData, _ interface{}) error {
	// There is no API for "deleting" indexing configuration or resetting it to "default" settings
	return nil
}
