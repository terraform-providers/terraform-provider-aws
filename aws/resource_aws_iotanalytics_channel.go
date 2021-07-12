package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsIoTAnalyticsChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIoTAnalyticsChannelCreate,
		Read:   resourceAwsIoTAnalyticsChannelRead,
		Update: resourceAwsIoTAnalyticsChannelUpdate,
		Delete: resourceAwsIoTAnalyticsChannelDelete,

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
			"channel_storage": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bucket": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"key_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"unlimited": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"number_of_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsIoTAnalyticsChannelCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn
	log.Print("Couldn't read first byte")
	x := d.Get("channel_storage")
	log.Print(x)
	params := &iotanalytics.CreateChannelInput{
		ChannelName: aws.String(d.Get("name").(string)),
	}

	// if v, ok := d.GetOk("thing_type_name"); ok {
	// 	params.ThingTypeName = aws.String(v.(string))
	// }
	// if v, ok := d.GetOk("attributes"); ok {
	// 	params.AttributePayload = &iotanalytics.AttributePayload{
	// 		Attributes: stringMapToPointers(v.(map[string]interface{})),
	// 	}
	// }

	log.Printf("[DEBUG] Creating IoT Analytics Channel: %s", params)
	out, err := conn.CreateChannel(params)
	if err != nil {
		return err
	}

	d.SetId(*out.ChannelName)

	return resourceAwsIoTAnalyticsChannelRead(d, meta)
}

func resourceAwsIoTAnalyticsChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DescribeChannelInput{
		ChannelName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Analytics Channel: %s", params)
	out, err := conn.DescribeChannel(params)

	if err != nil {
		if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] IoT Analytics Channel %q not found, removing from state", d.Id())
			d.SetId("")
		}
		return err
	}

	log.Printf("[DEBUG] Received IoT Analytics Channel: %s", out)

	d.Set("arn", out.Channel.Arn)
	d.Set("name", out.Channel.Name)
	// d.Set("attributes", aws.StringValueMap(out.Attributes))
	// d.Set("default_client_id", out.DefaultClientId)
	// d.Set("thing_type_name", out.ThingTypeName)
	// d.Set("version", out.Version)

	return nil
}

func resourceAwsIoTAnalyticsChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.UpdateChannelInput{
		ChannelName: aws.String(d.Get("name").(string)),
	}
	// if d.HasChange("thing_type_name") {
	// 	if v, ok := d.GetOk("thing_type_name"); ok {
	// 		params.ThingTypeName = aws.String(v.(string))
	// 	} else {
	// 		params.RemoveThingType = aws.Bool(true)
	// 	}
	// }
	// if d.HasChange("attributes") {
	// 	attributes := map[string]*string{}

	// 	if v, ok := d.GetOk("attributes"); ok {
	// 		if m, ok := v.(map[string]interface{}); ok {
	// 			attributes = stringMapToPointers(m)
	// 		}
	// 	}
	// 	params.AttributePayload = &iot.AttributePayload{
	// 		Attributes: attributes,
	// 	}
	// }

	_, err := conn.UpdateChannel(params)
	if err != nil {
		return err
	}

	return resourceAwsIoTAnalyticsChannelRead(d, meta)
}

func resourceAwsIoTAnalyticsChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeleteChannelInput{
		ChannelName: aws.String(d.Get("name").(string)),
	}
	log.Printf("[DEBUG] Deleting IoT Analytics Channel: %s", params)

	_, err := conn.DeleteChannel(params)
	if err != nil {
		if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}
