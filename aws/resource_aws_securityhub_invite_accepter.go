package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSecurityHubInviteAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubInviteAccepterCreate,
		Read:   resourceAwsSecurityHubInviteAccepterRead,
		Delete: resourceAwsSecurityHubInviteAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"master_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"invitation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsSecurityHubInviteAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Accepting Security Hub invitation")

	invitationId, err := resourceAwsSecurityHubInviteAccepterGetInvitationId(conn, d.Get("master_id").(string))

	if err != nil {
		return err
	}

	_, err = conn.AcceptInvitation(&securityhub.AcceptInvitationInput{
		InvitationId: aws.String(invitationId),
		MasterId:     aws.String(d.Get("master_id").(string)),
	})

	if err != nil {
		return fmt.Errorf("error accepting Security Hub invitation: %w", err)
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceAwsSecurityHubInviteAccepterRead(d, meta)
}

func resourceAwsSecurityHubInviteAccepterGetInvitationId(conn *securityhub.SecurityHub, masterId string) (string, error) {
	log.Printf("[DEBUG] Getting InvitationId for MasterId %s", masterId)

	resp, err := conn.ListInvitations(&securityhub.ListInvitationsInput{})

	if err != nil {
		return "", fmt.Errorf("error listing Security Hub invitations: %w", err)
	}

	for _, invitation := range resp.Invitations {
		log.Printf("[DEBUG] Invitation: %s", invitation)
		if *invitation.AccountId == masterId {
			return *invitation.InvitationId, nil
		}
	}

	return "", fmt.Errorf("Cannot find InvitationId for MasterId %s", masterId)
}

func resourceAwsSecurityHubInviteAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Reading Security Hub master account")

	resp, err := conn.GetMasterAccount(&securityhub.GetMasterAccountInput{})
	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		log.Print("[WARN] Security Hub master account not found, removing from state")
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error retrieving Security Hub master account: %w", err)
	}

	master := resp.Master

	if master == nil {
		log.Print("[WARN] Security Hub master account not found, removing from state")
		d.SetId("")
		return nil
	}

	d.Set("invitation_id", master.InvitationId)
	d.Set("master_id", master.AccountId)

	return nil
}

func resourceAwsSecurityHubInviteAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Print("[DEBUG] Disassociating from Security Hub master account")

	_, err := conn.DisassociateFromMasterAccount(&securityhub.DisassociateFromMasterAccountInput{})

	if tfawserr.ErrMessageContains(err, "BadRequestException", "The request is rejected because the current account is not associated to a master account") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error disassociating from Security Hub master account: %w", err)
	}

	return nil
}
