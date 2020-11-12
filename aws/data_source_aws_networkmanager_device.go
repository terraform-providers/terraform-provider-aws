package aws

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsNetworkManagerDevice() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsNetworkManagerDeviceRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"location": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vendor": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsNetworkManagerDeviceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &networkmanager.GetDevicesInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	}

	if v, ok := d.GetOk("id"); ok {
		input.DeviceIds = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("site_id"); ok {
		input.SiteId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Reading Network Manager Device: %s", input)
	output, err := conn.GetDevices(input)

	if err != nil {
		return fmt.Errorf("error reading Network Manager Device: %s", err)
	}

	// do filtering here
	var filteredDevices []*networkmanager.Device
	if tags, ok := d.GetOk("tags"); ok {
		keyValueTags := keyvaluetags.New(tags.(map[string]interface{})).IgnoreAws()
		for _, device := range output.Devices {
			tagsMatch := true
			if len(keyValueTags) > 0 {
				listTags := keyvaluetags.NetworkmanagerKeyValueTags(device.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)
				tagsMatch = listTags.ContainsAll(keyValueTags)
			}
			if tagsMatch {
				filteredDevices = append(filteredDevices, device)
			}
		}
	} else {
		filteredDevices = output.Devices
	}

	if output == nil || len(filteredDevices) == 0 {
		return errors.New("error reading Network Manager Device: no results found")
	}

	if len(filteredDevices) > 1 {
		return errors.New("error reading Network Manager Device: more than one result found. Please try a more specific search criteria.")
	}

	device := filteredDevices[0]

	if device == nil {
		return errors.New("error reading Network Manager Device: empty result")
	}

	d.Set("arn", device.DeviceArn)
	d.Set("description", device.Description)
	d.Set("model", device.Model)
	d.Set("serial_number", device.SerialNumber)
	d.Set("site_id", device.SiteId)
	d.Set("type", device.Type)
	d.Set("vendor", device.Vendor)

	if err := d.Set("location", flattenNetworkManagerLocation(device.Location)); err != nil {
		return fmt.Errorf("error setting location: %s", err)
	}
	if err := d.Set("tags", keyvaluetags.NetworkmanagerKeyValueTags(device.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.SetId(aws.StringValue(device.DeviceId))

	return nil
}
