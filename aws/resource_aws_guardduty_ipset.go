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

func resourceAwsGuardDutyIpset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyIpsetCreate,
		Read:   resourceAwsGuardDutyIpsetRead,
		Update: resourceAwsGuardDutyIpsetUpdate,
		Delete: resourceAwsGuardDutyIpsetDelete,

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
					guardduty.IpSetFormatTxt,
					guardduty.IpSetFormatStix,
					guardduty.IpSetFormatOtxCsv,
					guardduty.IpSetFormatAlienVault,
					guardduty.IpSetFormatProofPoint,
					guardduty.IpSetFormatFireEye,
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

func resourceAwsGuardDutyIpsetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)
	input := &guardduty.CreateIPSetInput{
		DetectorId: aws.String(detectorID),
		Name:       aws.String(d.Get("name").(string)),
		Format:     aws.String(d.Get("format").(string)),
		Location:   aws.String(d.Get("location").(string)),
		Activate:   aws.Bool(d.Get("activate").(bool)),
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().GuarddutyTags()
	}

	resp, err := conn.CreateIPSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{guardduty.IpSetStatusActivating, guardduty.IpSetStatusDeactivating},
		Target:     []string{guardduty.IpSetStatusActive, guardduty.IpSetStatusInactive},
		Refresh:    guardDutyIpsetRefreshStatusFunc(conn, *resp.IpSetId, detectorID),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty IpSet status to be \"%s\" or \"%s\": %s", guardduty.IpSetStatusActive, guardduty.IpSetStatusInactive, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", detectorID, *resp.IpSetId))
	return resourceAwsGuardDutyIpsetRead(d, meta)
}

func resourceAwsGuardDutyIpsetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	ipSetId, detectorId, err := decodeGuardDutyIpsetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.GetIPSetInput{
		DetectorId: aws.String(detectorId),
		IpSetId:    aws.String(ipSetId),
	}

	resp, err := conn.GetIPSet(input)
	if err != nil {
		if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
			log.Printf("[WARN] GuardDuty IpSet %q not found, removing from state", ipSetId)
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
		Resource:  fmt.Sprintf("detector/%s/ipset/%s", detectorId, ipSetId),
	}.String()
	d.Set("arn", arn)

	d.Set("detector_id", detectorId)
	d.Set("format", resp.Format)
	d.Set("location", resp.Location)
	d.Set("name", resp.Name)
	d.Set("activate", *resp.Status == guardduty.IpSetStatusActive)

	if err := d.Set("tags", keyvaluetags.GuarddutyKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsGuardDutyIpsetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	ipSetId, detectorId, err := decodeGuardDutyIpsetID(d.Id())
	if err != nil {
		return err
	}

	if d.HasChanges("activate", "location", "name") {
		input := &guardduty.UpdateIPSetInput{
			DetectorId: aws.String(detectorId),
			IpSetId:    aws.String(ipSetId),
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

		_, err = conn.UpdateIPSet(input)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GuarddutyUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating GuardDuty IP Set (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsGuardDutyIpsetRead(d, meta)
}

func resourceAwsGuardDutyIpsetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	ipSetId, detectorId, err := decodeGuardDutyIpsetID(d.Id())
	if err != nil {
		return err
	}
	input := &guardduty.DeleteIPSetInput{
		DetectorId: aws.String(detectorId),
		IpSetId:    aws.String(ipSetId),
	}

	_, err = conn.DeleteIPSet(input)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			guardduty.IpSetStatusActive,
			guardduty.IpSetStatusActivating,
			guardduty.IpSetStatusInactive,
			guardduty.IpSetStatusDeactivating,
			guardduty.IpSetStatusDeletePending,
		},
		Target:     []string{guardduty.IpSetStatusDeleted},
		Refresh:    guardDutyIpsetRefreshStatusFunc(conn, ipSetId, detectorId),
		Timeout:    5 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for GuardDuty IpSet status to be \"%s\": %s", guardduty.IpSetStatusDeleted, err)
	}

	return nil
}

func guardDutyIpsetRefreshStatusFunc(conn *guardduty.GuardDuty, ipSetID, detectorID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.GetIPSetInput{
			DetectorId: aws.String(detectorID),
			IpSetId:    aws.String(ipSetID),
		}
		resp, err := conn.GetIPSet(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, *resp.Status, nil
	}
}

func decodeGuardDutyIpsetID(id string) (ipsetID, detectorID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		err = fmt.Errorf("GuardDuty IPSet ID must be of the form <Detector ID>:<IPSet ID>, was provided: %s", id)
		return
	}
	ipsetID = parts[1]
	detectorID = parts[0]
	return
}
