package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGuardDutyThreatintelset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyThreatintelsetCreate,
		Read:   resourceAwsGuardDutyThreatintelsetRead,
		Update: resourceAwsGuardDutyThreatintelsetUpdate,
		Delete: resourceAwsGuardDutyThreatintelsetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"format": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					guardduty.ThreatIntelSetFormatTxt,
					guardduty.ThreatIntelSetFormatStix,
					guardduty.ThreatIntelSetFormatOtxCsv,
					guardduty.ThreatIntelSetFormatAlienVault,
					guardduty.ThreatIntelSetFormatProofPoint,
					guardduty.ThreatIntelSetFormatFireEye,
				}, false),
			},
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"activate": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsGuardDutyThreatintelsetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.CreateThreatIntelSetInput{
		DetectorId: aws.String(detectorID),
		Name:       aws.String(d.Get("name").(string)),
		Format:     aws.String(d.Get("format").(string)),
		Location:   aws.String(d.Get("location").(string)),
		Activate:   aws.Bool(d.Get("activate").(bool)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().GuarddutyTags()
	}

	resp, err := conn.CreateThreatIntelSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{guardduty.ThreatIntelSetStatusActivating, guardduty.ThreatIntelSetStatusDeactivating},
		Target:     []string{guardduty.ThreatIntelSetStatusActive, guardduty.ThreatIntelSetStatusInactive},
		Refresh:    guardDutyThreatintelsetRefreshStatusFunc(conn, *resp.ThreatIntelSetId, detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty ThreatIntelSet status to be \"%s\" or \"%s\": %s",
			guardduty.ThreatIntelSetStatusActive, guardduty.ThreatIntelSetStatusInactive, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, *resp.ThreatIntelSetId))
	return resourceAwsGuardDutyThreatintelsetRead(d, meta)
}

func resourceAwsGuardDutyThreatintelsetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	threatIntelSetId, detectorId, err := decodeGuardDutyThreatintelsetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.GetThreatIntelSetInput{
		DetectorId:       aws.String(detectorId),
		ThreatIntelSetId: aws.String(threatIntelSetId),
	}

	resp, err := conn.GetThreatIntelSet(input)
	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty ThreatIntelSet %q not found, removing from state", threatIntelSetId)
			d.SetId("")
			return nil
		}
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "guardduty",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("detector/%s/threatintelset/%s", detectorId, threatIntelSetId),
	}.String()
	d.Set("arn", arn)

	d.Set("detector_id", detectorId)
	d.Set("format", resp.Format)
	d.Set("location", resp.Location)
	d.Set("name", resp.Name)
	d.Set("activate", *resp.Status == guardduty.ThreatIntelSetStatusActive)

	if err := d.Set("tags", keyvaluetags.GuarddutyKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGuardDutyThreatintelsetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	threatIntelSetID, detectorId, err := decodeGuardDutyThreatintelsetID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChanges("activate", "location", "name") {
		input := &guardduty.UpdateThreatIntelSetInput{
			DetectorId:       aws.String(detectorId),
			ThreatIntelSetId: aws.String(threatIntelSetID),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}
		if d.HasChange("location") {
			input.Location = aws.String(d.Get("location").(string))
		}
		if d.HasChange("activate") {
			input.Activate = aws.Bool(d.Get("activate").(bool))
		}

		_, err = conn.UpdateThreatIntelSet(input)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GuarddutyUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating GuardDuty Threat Intel Set (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsGuardDutyThreatintelsetRead(d, meta)
}

func resourceAwsGuardDutyThreatintelsetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	threatIntelSetID, detectorId, err := decodeGuardDutyThreatintelsetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.DeleteThreatIntelSetInput{
		DetectorId:       aws.String(detectorId),
		ThreatIntelSetId: aws.String(threatIntelSetID),
	}

	_, err = conn.DeleteThreatIntelSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			guardduty.ThreatIntelSetStatusActive,
			guardduty.ThreatIntelSetStatusActivating,
			guardduty.ThreatIntelSetStatusInactive,
			guardduty.ThreatIntelSetStatusDeactivating,
			guardduty.ThreatIntelSetStatusDeletePending,
		},
		Target:     []string{guardduty.ThreatIntelSetStatusDeleted},
		Refresh:    guardDutyThreatintelsetRefreshStatusFunc(conn, threatIntelSetID, detectorId),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty ThreatIntelSet status to be \"%s\": %s", guardduty.ThreatIntelSetStatusDeleted, err)
	}

	return nil
}

func guardDutyThreatintelsetRefreshStatusFunc(conn *guardduty.GuardDuty, threatIntelSetID, detectorID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.GetThreatIntelSetInput{
			DetectorId:       aws.String(detectorID),
			ThreatIntelSetId: aws.String(threatIntelSetID),
		}
		resp, err := conn.GetThreatIntelSet(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Status, nil
	}
}

func decodeGuardDutyThreatintelsetID(id string) (threatIntelSetID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty ThreatIntelSet ID must be of the form <Detector ID>:<ThreatIntelSet ID>, was provided: %s", id)
		return
	}
	threatIntelSetID = parts[1]
	detectorID = parts[0]
	return
}
