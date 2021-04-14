package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsGuardDutyInviteAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyInviteAccepterCreate,
		Read:   resourceAwsGuardDutyInviteAccepterRead,
		Delete: resourceAwsGuardDutyInviteAccepterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"master_account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},
	}
}

func resourceAwsGuardDutyInviteAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorID := d.Get("detector_id").(string)
	invitationID := ""
	masterAccountID := d.Get("master_account_id").(string)

	listInvitationsInput := &guardduty.ListInvitationsInput{}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		log.Printf("[DEBUG] Listing GuardDuty Invitations: %s", listInvitationsInput)
		err := conn.ListInvitationsPages(listInvitationsInput, func(page *guardduty.ListInvitationsOutput, lastPage bool) bool {
			for _, invitation := range page.Invitations {
				if aws.StringValue(invitation.AccountId) == masterAccountID {
					invitationID = aws.StringValue(invitation.InvitationId)
					return false
				}
			}
			return !lastPage
		})

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if invitationID == "" {
			return resource.RetryableError(fmt.Errorf("unable to find pending GuardDuty Invitation for detector ID (%s) from master account ID (%s)", detectorID, masterAccountID))
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		err = conn.ListInvitationsPages(listInvitationsInput, func(page *guardduty.ListInvitationsOutput, lastPage bool) bool {
			for _, invitation := range page.Invitations {
				if aws.StringValue(invitation.AccountId) == masterAccountID {
					invitationID = aws.StringValue(invitation.InvitationId)
					return false
				}
			}
			return !lastPage
		})
	}

	if err != nil {
		return fmt.Errorf("error listing GuardDuty Invitations: %s", err)
	}

	acceptInvitationInput := &guardduty.AcceptInvitationInput{
		DetectorId:   aws.String(detectorID),
		InvitationId: aws.String(invitationID),
		MasterId:     aws.String(masterAccountID),
	}

	log.Printf("[DEBUG] Accepting GuardDuty Invitation: %s", acceptInvitationInput)
	_, err = conn.AcceptInvitation(acceptInvitationInput)

	if err != nil {
		return fmt.Errorf("error accepting GuardDuty Invitation (%s): %s", invitationID, err)
	}

	d.SetId(detectorID)

	return resourceAwsGuardDutyInviteAccepterRead(d, meta)
}

func resourceAwsGuardDutyInviteAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.GetMasterAccountInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading GuardDuty Master Account: %s", input)
	output, err := conn.GetMasterAccount(input)

	if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		log.Printf("[WARN] GuardDuty Detector %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading GuardDuty Detector (%s) GuardDuty Master Account: %s", d.Id(), err)
	}

	if output == nil || output.Master == nil {
		return fmt.Errorf("error reading GuardDuty Detector (%s) GuardDuty Master Account: empty response", d.Id())
	}

	d.Set("detector_id", d.Id())
	d.Set("master_account_id", output.Master.AccountId)

	return nil
}

func resourceAwsGuardDutyInviteAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := &guardduty.DisassociateFromMasterAccountInput{
		DetectorId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Disassociating GuardDuty Detector (%s) from GuardDuty Master Account: %s", d.Id(), input)
	_, err := conn.DisassociateFromMasterAccount(input)

	if isAWSErr(err, guardduty.ErrCodeBadRequestException, "The request is rejected because the input detectorId is not owned by the current account.") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating GuardDuty Member Detector (%s) from GuardDuty Master Account: %s", d.Id(), err)
	}

	return nil
}
