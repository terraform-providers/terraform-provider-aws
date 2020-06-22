package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsGuardDutyOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyOrganizationConfigurationUpdate,
		Read:   resourceAwsGuardDutyOrganizationConfigurationRead,
		Update: resourceAwsGuardDutyOrganizationConfigurationUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"detector_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceAwsGuardDutyOrganizationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)

	input := &guardduty.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(d.Get("auto_enable").(bool)),
		DetectorId: aws.String(detectorID),
	}

	_, err := conn.UpdateOrganizationConfiguration(input)

	if err != nil {
		return fmt.Errorf("error updating GuardDuty Organization Configuration (%s): %w", detectorID, err)
	}

	d.SetId(detectorID)

	return resourceAwsGuardDutyOrganizationConfigurationRead(d, meta)
}

func resourceAwsGuardDutyOrganizationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.DescribeOrganizationConfigurationInput{
		DetectorId: aws.String(d.Id()),
	}

	output, err := conn.DescribeOrganizationConfiguration(input)

	if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		log.Printf("[WARN] GuardDuty Organization Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading GuardDuty Organization Configuration (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading GuardDuty Organization Configuration (%s): empty response", d.Id())
	}

	d.Set("detector_id", d.Id())
	d.Set("auto_enable", output.AutoEnable)

	return nil
}
