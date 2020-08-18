package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsIotEventsInput() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotInputCreate,
		Read:   resourceAwsIotInputRead,
		Update: resourceAwsIotInputUpdate,
		Delete: resourceAwsIotInputDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"definition": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"json_path": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 128),
									},
								},
							},
						},
					},
				},
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func prepareInputDefinition(d *schema.ResourceData) *iotevents.InputDefinition {
	rawInputDefinition := d.Get("definition").(*schema.Set).List()[0].(map[string]interface{})
	rawAttributes := rawInputDefinition["attribute"].(*schema.Set).List()

	// Convert raw attributes data to a list of Attributes structures
	attributes := make([]*iotevents.Attribute, len(rawAttributes))
	for i, v := range rawAttributes {
		rawAttr := v.(map[string]interface{})
		attributes[i] = &iotevents.Attribute{
			JsonPath: aws.String(rawAttr["json_path"].(string)),
		}
	}

	inputDefinition := &iotevents.InputDefinition{
		Attributes: attributes,
	}
	return inputDefinition
}

func flattenIoTEventsInputDefinition(inputDefinition *iotevents.InputDefinition) []map[string]interface{} {
	attributes := make([]map[string]interface{}, 0)

	for _, v := range inputDefinition.Attributes {
		result := make(map[string]interface{})
		result["json_path"] = aws.StringValue(v.JsonPath)
		attributes = append(attributes, result)
	}
	rawInputDefinition := make(map[string]interface{})
	rawInputDefinition["attribute"] = attributes

	return []map[string]interface{}{rawInputDefinition}
}

func resourceAwsIotInputCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	inputName := d.Get("name").(string)

	params := &iotevents.CreateInputInput{
		InputName:       aws.String(inputName),
		InputDefinition: prepareInputDefinition(d),
		Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().IoteventsTags(),
	}

	if v, ok := d.GetOk("description"); ok {
		params.InputDescription = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating IoT Input: %s", params)
	_, err := conn.CreateInput(params)

	if err != nil {
		return err
	}

	d.SetId(inputName)

	return resourceAwsIotInputRead(d, meta)
}

func resourceAwsIotInputRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	params := &iotevents.DescribeInputInput{
		InputName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Events Input: %s", params)
	out, err := conn.DescribeInput(params)

	if err != nil {
		return err
	}

	d.Set("name", out.Input.InputConfiguration.InputName)
	d.Set("description", out.Input.InputConfiguration.InputDescription)
	d.Set("definition", flattenIoTEventsInputDefinition(out.Input.InputDefinition))
	d.Set("arn", out.Input.InputConfiguration.InputArn)

	arn := *out.Input.InputConfiguration.InputArn
	tags, err := keyvaluetags.IoteventsListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsIotInputUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	inputName := d.Get("name").(string)

	params := &iotevents.UpdateInputInput{
		InputName:       aws.String(inputName),
		InputDefinition: prepareInputDefinition(d),
	}

	if v, ok := d.GetOk("description"); ok {
		params.InputDescription = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Update IoT Input: %s", params)
	_, err := conn.UpdateInput(params)

	if err != nil {
		return err
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.IoteventsUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}
	return resourceAwsIotInputRead(d, meta)
}

func resourceAwsIotInputDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ioteventsconn

	params := &iotevents.DeleteInputInput{
		InputName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Events Input: %s", params)
	_, err := conn.DeleteInput(params)

	return err
}
