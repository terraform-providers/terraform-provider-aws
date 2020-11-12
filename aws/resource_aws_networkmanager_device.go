package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsNetworkManagerDevice() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNetworkManagerDeviceCreate,
		Read:   resourceAwsNetworkManagerDeviceRead,
		Update: resourceAwsNetworkManagerDeviceUpdate,
		Delete: resourceAwsNetworkManagerDeviceDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("arn", d.Id())

				idErr := fmt.Errorf("Expected ID in format of arn:aws:networkmanager::ACCOUNTID:device/GLOBALNETWORKID/DEVICEID and provided: %s", d.Id())

				resARN, err := arn.Parse(d.Id())
				if err != nil {
					return nil, idErr
				}

				identifiers := strings.TrimPrefix(resARN.Resource, "device/")
				identifierParts := strings.Split(identifiers, "/")
				if len(identifierParts) != 2 {
					return nil, idErr
				}
				d.SetId(identifierParts[1])
				d.Set("global_network_id", identifierParts[0])

				return []*schema.ResourceData{d}, nil
			},
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
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"location": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
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
				Optional: true,
			},
			"serial_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vendor": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsNetworkManagerDeviceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.CreateDeviceInput{
		Description:     aws.String(d.Get("description").(string)),
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
		Location:        expandNetworkManagerLocation(d.Get("location").([]interface{})),
		Model:           aws.String(d.Get("model").(string)),
		SerialNumber:    aws.String(d.Get("serial_number").(string)),
		SiteId:          aws.String(d.Get("site_id").(string)),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().NetworkmanagerTags(),
		Type:            aws.String(d.Get("type").(string)),
		Vendor:          aws.String(d.Get("vendor").(string)),
	}

	log.Printf("[DEBUG] Creating Network Manager Device: %s", input)

	output, err := conn.CreateDevice(input)
	if err != nil {
		return fmt.Errorf("error creating Network Manager Device: %s", err)
	}

	d.SetId(aws.StringValue(output.Device.DeviceId))

	stateConf := &resource.StateChangeConf{
		Pending: []string{networkmanager.DeviceStatePending},
		Target:  []string{networkmanager.DeviceStateAvailable},
		Refresh: networkmanagerDeviceRefreshFunc(conn, aws.StringValue(output.Device.GlobalNetworkId), aws.StringValue(output.Device.DeviceId)),
		Timeout: 10 * time.Minute,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Network Manager Device (%s) availability: %s", d.Id(), err)
	}

	return resourceAwsNetworkManagerDeviceRead(d, meta)
}

func resourceAwsNetworkManagerDeviceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	device, err := networkmanagerDescribeDevice(conn, d.Get("global_network_id").(string), d.Id())

	if isAWSErr(err, "InvalidDeviceID.NotFound", "") {
		log.Printf("[WARN] Network Manager Device (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Network Manager Device: %s", err)
	}

	if device == nil {
		log.Printf("[WARN] Network Manager Device (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(device.State) == networkmanager.DeviceStateDeleting {
		log.Printf("[WARN] Network Manager Device (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(device.State))
		d.SetId("")
		return nil
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

	return nil
}

func resourceAwsNetworkManagerDeviceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	if d.HasChanges("description", "location", "model", "serial_number", "site_id", "type", "vendor") {
		request := &networkmanager.UpdateDeviceInput{
			Description:     aws.String(d.Get("description").(string)),
			DeviceId:        aws.String(d.Id()),
			GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
			Location:        expandNetworkManagerLocation(d.Get("location").([]interface{})),
			Model:           aws.String(d.Get("model").(string)),
			SerialNumber:    aws.String(d.Get("serial_number").(string)),
			SiteId:          aws.String(d.Get("site_id").(string)),
			Type:            aws.String(d.Get("type").(string)),
			Vendor:          aws.String(d.Get("vendor").(string)),
		}

		_, err := conn.UpdateDevice(request)
		if err != nil {
			return fmt.Errorf("Failure updating Network Manager Device (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.NetworkmanagerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Network Manager Device (%s) tags: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsNetworkManagerDeviceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).networkmanagerconn

	input := &networkmanager.DeleteDeviceInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
		DeviceId:        aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Network Manager Device (%s): %s", d.Id(), input)
	err := resource.Retry(2*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDevice(input)

		if isAWSErr(err, "IncorrectState", "has non-deleted Device Associations") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "IncorrectState", "has non-deleted Device") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteDevice(input)
	}

	if isAWSErr(err, "InvalidDeviceID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Network Manager Device: %s", err)
	}

	if err := waitForNetworkManagerDeviceDeletion(conn, d.Get("global_network_id").(string), d.Id()); err != nil {
		return fmt.Errorf("error waiting for Network Manager Device (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func networkmanagerDeviceRefreshFunc(conn *networkmanager.NetworkManager, globalNetworkID, deviceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		device, err := networkmanagerDescribeDevice(conn, globalNetworkID, deviceID)

		if isAWSErr(err, "InvalidDeviceID.NotFound", "") {
			return nil, "DELETED", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading Network Manager Device (%s): %s", deviceID, err)
		}

		if device == nil {
			return nil, "DELETED", nil
		}

		return device, aws.StringValue(device.State), nil
	}
}

func networkmanagerDescribeDevice(conn *networkmanager.NetworkManager, globalNetworkID, deviceID string) (*networkmanager.Device, error) {
	input := &networkmanager.GetDevicesInput{
		GlobalNetworkId: aws.String(globalNetworkID),
		DeviceIds:       []*string{aws.String(deviceID)},
	}

	log.Printf("[DEBUG] Reading Network Manager Device (%s): %s", deviceID, input)
	for {
		output, err := conn.GetDevices(input)

		if err != nil {
			return nil, err
		}

		if output == nil || len(output.Devices) == 0 {
			return nil, nil
		}

		for _, device := range output.Devices {
			if device == nil {
				continue
			}

			if aws.StringValue(device.DeviceId) == deviceID {
				return device, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil, nil
}

func waitForNetworkManagerDeviceDeletion(conn *networkmanager.NetworkManager, globalNetworkID, deviceID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			networkmanager.DeviceStateAvailable,
			networkmanager.DeviceStateDeleting,
		},
		Target:         []string{""},
		Refresh:        networkmanagerDeviceRefreshFunc(conn, globalNetworkID, deviceID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for Network Manager Device (%s) deletion", deviceID)
	_, err := stateConf.WaitForState()

	if isResourceNotFoundError(err) {
		return nil
	}

	return err
}
