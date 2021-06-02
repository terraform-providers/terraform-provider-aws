package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/waiter"
)

func resourceAwsRamResourceShareAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamResourceShareAccepterCreate,
		Read:   resourceAwsRamResourceShareAccepterRead,
		Delete: resourceAwsRamResourceShareAccepterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"invitation_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"receiver_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sender_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceAwsRamResourceShareAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	shareARN := d.Get("share_arn").(string)

	invitation, err := finder.ResourceShareInvitationByResourceShareArnAndStatus(conn, shareARN, ram.ResourceShareInvitationStatusPending)

	if err != nil {
		return err
	}

	if invitation == nil || aws.StringValue(invitation.ResourceShareInvitationArn) == "" {
		return fmt.Errorf(
			"No RAM Resource Share (%s) invitation found\n\n"+
				"NOTE: If both AWS accounts are in the same AWS Organization and RAM Sharing with AWS Organizations is enabled, this resource is not necessary",
			shareARN)
	}

	input := &ram.AcceptResourceShareInvitationInput{
		ClientToken:                aws.String(resource.UniqueId()),
		ResourceShareInvitationArn: invitation.ResourceShareInvitationArn,
	}

	log.Printf("[DEBUG] Accept RAM resource share invitation request: %s", input)
	output, err := conn.AcceptResourceShareInvitation(input)

	if err != nil {
		return fmt.Errorf("Error accepting RAM resource share invitation: %s", err)
	}

	d.SetId(shareARN)

	_, err = waiter.ResourceShareInvitationAccepted(
		conn,
		aws.StringValue(output.ResourceShareInvitation.ResourceShareInvitationArn),
		d.Timeout(schema.TimeoutCreate),
	)

	if err != nil {
		return fmt.Errorf("Error waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return resourceAwsRamResourceShareAccepterRead(d, meta)
}

func resourceAwsRamResourceShareAccepterRead(d *schema.ResourceData, meta interface{}) error {
	accountID := meta.(*AWSClient).accountid
	conn := meta.(*AWSClient).ramconn

	invitation, err := finder.ResourceShareInvitationByResourceShareArnAndStatus(conn, d.Id(), ram.ResourceShareInvitationStatusAccepted)

	if err != nil && !tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceShareInvitationArnNotFoundException) {
		return fmt.Errorf("error retrieving invitation for resource share %s: %w", d.Id(), err)
	}

	if invitation != nil {
		d.Set("invitation_arn", invitation.ResourceShareInvitationArn)
		d.Set("receiver_account_id", invitation.ReceiverAccountId)
	} else {
		d.Set("receiver_account_id", accountID)
	}

	resourceShare, err := finder.ResourceShareOwnerOtherAccountsByArn(conn, d.Id())

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceArnNotFoundException) || tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException)) {
		log.Printf("[WARN] No RAM resource share with ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving resource share: %w", err)
	}

	if resourceShare == nil {
		return fmt.Errorf("error getting resource share (%s): empty result", d.Id())
	}

	d.Set("status", resourceShare.Status)
	d.Set("sender_account_id", resourceShare.OwningAccountId)
	d.Set("share_arn", resourceShare.ResourceShareArn)
	d.Set("share_id", resourceAwsRamResourceShareGetIDFromARN(d.Id()))
	d.Set("share_name", resourceShare.Name)

	listInput := &ram.ListResourcesInput{
		MaxResults:        aws.Int64(int64(500)),
		ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		ResourceShareArns: aws.StringSlice([]string{d.Id()}),
	}

	var resourceARNs []*string
	err = conn.ListResourcesPages(listInput, func(page *ram.ListResourcesOutput, lastPage bool) bool {
		for _, resource := range page.Resources {
			resourceARNs = append(resourceARNs, resource.Arn)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("Error reading RAM resource share resources %s: %s", d.Id(), err)
	}

	if err := d.Set("resources", flattenStringList(resourceARNs)); err != nil {
		return fmt.Errorf("unable to set resources: %s", err)
	}

	return nil
}

func resourceAwsRamResourceShareAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	receiverAccountID := d.Get("receiver_account_id").(string)

	if receiverAccountID == "" {
		return fmt.Errorf("The receiver account ID is required to leave a resource share")
	}

	input := &ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceShareArn: aws.String(d.Id()),
		Principals:       []*string{aws.String(receiverAccountID)},
	}
	log.Printf("[DEBUG] Leave RAM resource share request: %s", input)

	_, err := conn.DisassociateResourceShare(input)

	if err != nil {
		return fmt.Errorf("Error leaving RAM resource share: %s", err)
	}

	_, err = waiter.ResourceShareOwnedBySelfDisassociated(conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return fmt.Errorf("Error waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRamResourceShareGetIDFromARN(arn string) string {
	return strings.Replace(arn[strings.LastIndex(arn, ":")+1:], "resource-share/", "rs-", -1)
}
