package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsIotThingGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotThingGroupCreate,
		Read:   resourceAwsIotThingGroupRead,
		Update: resourceAwsIotThingGroupUpdate,
		Delete: resourceAwsIotThingGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"parent_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"properties": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"attributes": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     schema.TypeString,
						},
						"merge": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func parseProperties(rawProperties map[string]interface{}) *iot.ThingGroupProperties {
	rawAttributes := rawProperties["attributes"].(map[string]interface{})
	attributes := make(map[string]*string)
	for key, value := range rawAttributes {
		attributes[key] = aws.String(value.(string))
	}
	attributePayload := &iot.AttributePayload{
		Attributes: attributes,
		Merge:      aws.Bool(rawProperties["merge"].(bool)),
	}

	properties := &iot.ThingGroupProperties{
		AttributePayload: attributePayload,
	}

	if v, ok := rawProperties["description"]; ok {
		properties.ThingGroupDescription = aws.String(v.(string))
	}

	return properties
}

func resourceAwsIotThingGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	name := d.Get("name").(string)
	params := &iot.CreateThingGroupInput{
		ThingGroupName: aws.String(name),
	}

	if tags := d.Get("tags").(map[string]interface{}); len(tags) > 0 {
		params.Tags = keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IotTags()
	}

	if v, ok := d.GetOk("parent_group_name"); ok {
		params.ParentGroupName = aws.String(v.(string))
	}

	if v := d.Get("properties").([]interface{}); len(v) >= 1 {
		params.ThingGroupProperties = parseProperties(v[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating IoT Thing Group: %s", params)
	_, err := conn.CreateThingGroup(params)

	if err != nil {
		return err
	}

	d.SetId(name)

	return resourceAwsIotThingGroupRead(d, meta)
}

func flattenProperties(properties *iot.ThingGroupProperties) map[string]interface{} {
	groupProperties := make(map[string]interface{})

	if properties.AttributePayload != nil {
		rawAttributes := make(map[string]interface{})
		attributes := properties.AttributePayload.Attributes
		for key, value := range attributes {
			rawAttributes[key] = aws.StringValue(value)
		}
		groupProperties["attributes"] = rawAttributes
		groupProperties["merge"] = aws.BoolValue(properties.AttributePayload.Merge)
	}

	if properties.ThingGroupDescription != nil {
		groupProperties["description"] = aws.StringValue(properties.ThingGroupDescription)
	}

	return groupProperties
}

func resourceAwsIotThingGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	params := &iot.DescribeThingGroupInput{
		ThingGroupName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Thing: %s", params)
	out, err := conn.DescribeThingGroup(params)

	if err != nil {
		return err
	}

	d.Set("arn", out.ThingGroupArn)
	d.Set("name", out.ThingGroupName)
	d.Set("parent_group_name", out.ThingGroupMetadata.ParentGroupName)
	properties := []map[string]interface{}{flattenProperties(out.ThingGroupProperties)}
	d.Set("properties", properties)
	d.Set("version", out.Version)

	arn := *out.ThingGroupArn
	tags, err := keyvaluetags.IotListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsIotThingGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	name := d.Get("name").(string)
	params := &iot.UpdateThingGroupInput{
		ThingGroupName: aws.String(name),
	}

	if v := d.Get("properties").([]interface{}); len(v) >= 1 {
		params.ThingGroupProperties = parseProperties(v[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Updating IoT Thing Group: %s", params)
	_, err := conn.UpdateThingGroup(params)

	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.IotUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsIotThingGroupRead(d, meta)
}

func resourceAwsIotThingGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn

	params := &iot.DeleteThingGroupInput{
		ThingGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Updating IoT Thing Group: %s", params)
	_, err := conn.DeleteThingGroup(params)

	if err != nil {
		return err
	}

	return nil
}
