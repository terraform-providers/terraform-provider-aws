package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGreengrassDeviceDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassDeviceDefinitionCreate,
		Read:   resourceAwsGreengrassDeviceDefinitionRead,
		Update: resourceAwsGreengrassDeviceDefinitionUpdate,
		Delete: resourceAwsGreengrassDeviceDefinitionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"amzn_client_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"tags": tagsSchema(),
			"latest_definition_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"device_definition_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"sync_shadow": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"thing_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createDeviceDefinitionVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var rawData map[string]interface{}
	if v := d.Get("device_definition_version").(*schema.Set).List(); len(v) == 0 {
		return nil
	} else {
		rawData = v[0].(map[string]interface{})
	}

	params := &greengrass.CreateDeviceDefinitionVersionInput{
		DeviceDefinitionId: aws.String(d.Id()),
	}

	if v := d.Get("amzn_client_token").(string); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	devices := make([]*greengrass.Device, 0)
	for _, deviceToCast := range rawData["device"].(*schema.Set).List() {
		rawDevice := deviceToCast.(map[string]interface{})
		device := &greengrass.Device{
			CertificateArn: aws.String(rawDevice["certificate_arn"].(string)),
			Id:             aws.String(rawDevice["id"].(string)),
			SyncShadow:     aws.Bool(rawDevice["sync_shadow"].(bool)),
			ThingArn:       aws.String(rawDevice["thing_arn"].(string)),
		}
		devices = append(devices, device)
	}
	params.Devices = devices

	log.Printf("[DEBUG] Creating Greengrass Device Definition Version: %s", params)
	_, err := conn.CreateDeviceDefinitionVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassDeviceDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateDeviceDefinitionInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().GreengrassTags()
	}

	log.Printf("[DEBUG] Creating Greengrass Device Definition: %s", params)
	out, err := conn.CreateDeviceDefinition(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createDeviceDefinitionVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassDeviceDefinitionRead(d, meta)
}

func setDeviceDefinitionVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetDeviceDefinitionVersionInput{
		DeviceDefinitionId:        aws.String(d.Id()),
		DeviceDefinitionVersionId: aws.String(latestVersion),
	}

	out, err := conn.GetDeviceDefinitionVersion(params)

	if err != nil {
		return err
	}

	rawVersion := make(map[string]interface{})
	d.Set("latest_definition_version_arn", *out.Arn)

	rawDeviceList := make([]map[string]interface{}, 0)
	for _, device := range out.Definition.Devices {
		rawDevice := make(map[string]interface{})
		rawDevice["certificate_arn"] = *device.CertificateArn
		rawDevice["sync_shadow"] = *device.SyncShadow
		rawDevice["thing_arn"] = *device.ThingArn
		rawDevice["id"] = *device.Id
		rawDeviceList = append(rawDeviceList, rawDevice)
	}

	rawVersion["device"] = rawDeviceList

	d.Set("device_definition_version", []map[string]interface{}{rawVersion})

	return nil
}

func resourceAwsGreengrassDeviceDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetDeviceDefinitionInput{
		DeviceDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Device Definition: %s", params)
	out, err := conn.GetDeviceDefinition(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Device Definition: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	arn := *out.Arn
	tags, err := keyvaluetags.GreengrassListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}
	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	if out.LatestVersion != nil {
		err = setDeviceDefinitionVersion(*out.LatestVersion, d, conn)

		if err != nil {
			return err
		}
	}

	return nil
}

func resourceAwsGreengrassDeviceDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateDeviceDefinitionInput{
		Name:               aws.String(d.Get("name").(string)),
		DeviceDefinitionId: aws.String(d.Id()),
	}

	_, err := conn.UpdateDeviceDefinition(params)
	if err != nil {
		return err
	}

	if d.HasChange("device_definition_version") {
		err = createDeviceDefinitionVersion(d, conn)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.GreengrassUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}
	return resourceAwsGreengrassDeviceDefinitionRead(d, meta)
}

func resourceAwsGreengrassDeviceDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteDeviceDefinitionInput{
		DeviceDefinitionId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting Greengrass Device Definition: %s", params)

	_, err := conn.DeleteDeviceDefinition(params)

	if err != nil {
		return err
	}

	return nil
}
