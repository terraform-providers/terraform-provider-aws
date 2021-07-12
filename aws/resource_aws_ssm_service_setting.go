package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsSsmServiceSetting() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmServiceSettingUpdate,
		Read:   resourceAwsSsmServiceSettingRead,
		Update: resourceAwsSsmServiceSettingUpdate,
		Delete: resourceAwsSsmServiceSettingReset,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"setting_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"setting_value": {
				Type:     schema.TypeString,
				Required: true,
			},
			"last_modified_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_user": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSsmServiceSettingUpdate(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] SSM service setting create: %s", d.Id())

	updateServiceSettingInput := &ssm.UpdateServiceSettingInput{
		SettingId:    aws.String(d.Get("setting_id").(string)),
		SettingValue: aws.String(d.Get("setting_value").(string)),
	}

	if _, err := ssmconn.UpdateServiceSetting(updateServiceSettingInput); err != nil {
		return fmt.Errorf("Error updating SSM service setting: %s", err)
	}

	d.SetId(d.Get("setting_id").(string))

	return resourceAwsSsmServiceSettingRead(d, meta)
}

func resourceAwsSsmServiceSettingRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Reading SSM Activation: %s", d.Id())

	params := &ssm.GetServiceSettingInput{
		SettingId: aws.String(d.Id()),
	}

	resp, err := ssmconn.GetServiceSetting(params)

	if err != nil {
		return fmt.Errorf("Error reading SSM service setting: %s", err)
	}

	serviceSetting := resp.ServiceSetting
	// AWS SSM service setting API requires the entire ARN as input,
	// but setting_id in the output is only a part of ARN.
	d.Set("setting_id", serviceSetting.ARN)
	d.Set("setting_value", serviceSetting.SettingValue)
	d.Set("arn", serviceSetting.ARN)
	d.Set("last_modified_date", aws.TimeValue(serviceSetting.LastModifiedDate).Format(time.RFC3339))
	d.Set("last_modified_user", serviceSetting.LastModifiedUser)
	d.Set("status", serviceSetting.Status)

	return nil
}

func resourceAwsSsmServiceSettingReset(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Deleting SSM Service Setting: %s", d.Id())

	resetServiceSettingInput := &ssm.ResetServiceSettingInput{
		SettingId: aws.String(d.Get("setting_id").(string)),
	}

	_, err := ssmconn.ResetServiceSetting(resetServiceSettingInput)

	if err != nil {
		return fmt.Errorf("Error deleting SSM Service Setting: %s", err)
	}

	return nil
}
