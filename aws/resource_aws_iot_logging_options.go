package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
)

func resourceAwsIotLoggingOptions() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotLoggingOptionsCreate,
		Read:   resourceAwsIotLoggingOptionsRead,
		Update: resourceAwsIotLoggingOptionsUpdate,
		Delete: resourceAwsIotLoggingOptionsDelete,

		Schema: map[string]*schema.Schema{
			"default_log_level": {
				Type:         schema.TypeString,
				ValidateFunc: validateLogLevel(),
				Required:     true,
			},
			"disable_all_logs": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsIotLoggingOptionsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	input := &iot.SetV2LoggingOptionsInput{}
	if v, ok := d.GetOk("default_log_level"); ok {
		input.DefaultLogLevel = aws.String(v.(string))
	}
	if v, ok := d.GetOk("disable_all_logs"); ok {
		input.DisableAllLogs = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("role_arn"); ok {
		input.RoleArn = aws.String(v.(string))
	}
	out, err := conn.SetV2LoggingOptions(input)
	if err != nil {
		return fmt.Errorf("Set IoT Logging Options failed: %s", err)
	}
	log.Printf("[INFO] resurce create succes: %+v", out)
	return resourceAwsIotLoggingOptionsRead(d, meta)
}

func resourceAwsIotLoggingOptionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	out, err := conn.GetV2LoggingOptions(nil)
	if err != nil {
		return fmt.Errorf("Get IoT Loggin Options failed: %s", err)
	}
	log.Printf("[INFO] resource read success: %+v", out)

	d.Set("default_log_level", out.DefaultLogLevel)
	d.Set("disable_all_logs", out.DisableAllLogs)
	d.Set("role_arn", out.RoleArn)

	return nil
}

func resourceAwsIotLoggingOptionsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	input := &iot.SetV2LoggingOptionsInput{}

	if d.HasChange("default_log_level") {
		input.DefaultLogLevel = aws.String(d.Get("default_log_level").(string))
	}
	if d.HasChange("disable_all_logs") {
		input.DisableAllLogs = aws.Bool(d.Get("disable_all_logs").(bool))
	}
	if d.HasChange("role_arn") {
		input.RoleArn = aws.String(d.Get("role_arn").(string))
	}

	out, err := conn.SetV2LoggingOptions(input);
	if err != nil {
		return fmt.Errorf("Update IoT Logging Options failed: %s", err)
	}

	log.Printf("[INFO] resurce create succes: %+v", out)
	return resourceAwsIotLoggingOptionsRead(d, meta)
}

func resourceAwsIotLoggingOptionsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotconn
	_, err := conn.DeleteV2LoggingLevel(&iot.DeleteV2LoggingLevelInput{
		TargetName: aws.String(d.Id()),
		TargetType: aws.String(iot.LogTargetTypeThingGroup),
	})
	// TODO: check aws error.
	if err != nil {
		fmt.Errorf("Error deleting IoT Loggin Lavel: %s", err)
	}
	return nil
}

func validateLogLevel() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"DEBUG",
		"INFO",
		"ERROR",
		"WARN",
		"DISABLED",
	}, false)
}
