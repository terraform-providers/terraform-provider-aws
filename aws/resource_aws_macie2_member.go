package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/macie2/waiter"
)

func resourceAwsMacie2Member() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2MemberCreate,
		ReadWithoutTimeout:   resourceMacie2MemberRead,
		UpdateWithoutTimeout: resourceMacie2MemberUpdate,
		DeleteWithoutTimeout: resourceMacie2MemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"administrator_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invited_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(macie2.MacieStatus_Values(), false),
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"invitation_disable_email_notification": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"invitation_message": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Second),
			Update: schema.DefaultTimeout(60 * time.Second),
		},
	}
}

func resourceMacie2MemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	accountId := d.Get("account_id").(string)
	input := &macie2.CreateMemberInput{
		Account: &macie2.AccountDetail{
			AccountId: aws.String(accountId),
			Email:     aws.String(d.Get("email").(string)),
		},
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().Macie2Tags()
	}

	var err error
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateMemberWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateMemberWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie Member: %w", err))
	}

	d.SetId(accountId)

	if !d.Get("invite").(bool) {
		return resourceMacie2MemberRead(ctx, d, meta)
	}

	// Invitation workflow

	inputInvite := &macie2.CreateInvitationsInput{
		AccountIds: []*string{aws.String(d.Id())},
	}

	if v, ok := d.GetOk("invitation_disable_email_notification"); ok {
		inputInvite.DisableEmailNotification = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("invitation_message"); ok {
		inputInvite.Message = aws.String(v.(string))
	}

	log.Printf("[INFO] Inviting Macie2 Member: %s", inputInvite)

	var output *macie2.CreateInvitationsOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		output, err = conn.CreateInvitationsWithContext(ctx, inputInvite)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateInvitationsWithContext(ctx, inputInvite)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error inviting Macie Member: %w", err))
	}

	if len(output.UnprocessedAccounts) != 0 {
		return diag.FromErr(fmt.Errorf("error inviting Macie Member: %s: %s", aws.StringValue(output.UnprocessedAccounts[0].ErrorCode), aws.StringValue(output.UnprocessedAccounts[0].ErrorMessage)))
	}

	if _, err = waiter.MemberInvited(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for Macie Member (%s) invitation: %w", d.Id(), err))
	}

	return resourceMacie2MemberRead(ctx, d, meta)
}

func resourceMacie2MemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	input := &macie2.GetMemberInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetMemberWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeConflictException, "member accounts are associated with your account") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "account is not associated with your account") {
			log.Printf("[WARN] Macie Member (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading Macie Member (%s): %w", d.Id(), err))
	}

	d.Set("account_id", resp.AccountId)
	d.Set("email", resp.Email)
	d.Set("relationship_status", resp.RelationshipStatus)
	d.Set("administrator_account_id", resp.AdministratorAccountId)
	d.Set("master_account_id", resp.MasterAccountId)
	d.Set("invited_at", aws.TimeValue(resp.InvitedAt).Format(time.RFC3339))
	d.Set("updated_at", aws.TimeValue(resp.UpdatedAt).Format(time.RFC3339))
	d.Set("arn", resp.Arn)
	tags := keyvaluetags.Macie2KeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie Member (%s): %w", "tags", d.Id(), err))
	}

	if err = d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie Member (%s): %w", "tags_all", d.Id(), err))
	}

	status := aws.StringValue(resp.RelationshipStatus)
	log.Printf("[DEBUG] print resp.RelationshipStatus: %v", aws.StringValue(resp.RelationshipStatus))
	if status == macie2.RelationshipStatusEnabled ||
		status == macie2.RelationshipStatusInvited || status == macie2.RelationshipStatusEmailVerificationInProgress ||
		status == macie2.RelationshipStatusPaused {
		d.Set("invite", true)
	}

	if status == macie2.RelationshipStatusRemoved {
		d.Set("invite", false)
	}

	// To fake a result for status in order to avoid an error related to difference for ImportVerifyState
	// It sets to MacieStatusPaused because it can only be changed to PAUSED, normally when it's accepted its status is ENABLED
	status = macie2.MacieStatusEnabled
	if aws.StringValue(resp.RelationshipStatus) == macie2.RelationshipStatusPaused {
		status = macie2.MacieStatusPaused
	}
	d.Set("status", status)

	return nil
}

func resourceMacie2MemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	// Invitation workflow

	if d.HasChange("invite") {
		if d.Get("invite").(bool) {
			inputInvite := &macie2.CreateInvitationsInput{
				AccountIds: []*string{aws.String(d.Id())},
			}

			if v, ok := d.GetOk("invitation_disable_email_notification"); ok {
				inputInvite.DisableEmailNotification = aws.Bool(v.(bool))
			}
			if v, ok := d.GetOk("invitation_message"); ok {
				inputInvite.Message = aws.String(v.(string))
			}

			log.Printf("[INFO] Inviting Macie2 Member: %s", inputInvite)
			var output *macie2.CreateInvitationsOutput
			var err error
			err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
				output, err = conn.CreateInvitationsWithContext(ctx, inputInvite)

				if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
					return resource.RetryableError(err)
				}

				if err != nil {
					return resource.NonRetryableError(err)
				}

				return nil
			})

			if isResourceTimeoutError(err) {
				output, err = conn.CreateInvitationsWithContext(ctx, inputInvite)
			}

			if err != nil {
				return diag.FromErr(fmt.Errorf("error inviting Macie Member: %w", err))
			}

			if len(output.UnprocessedAccounts) != 0 {
				return diag.FromErr(fmt.Errorf("error inviting Macie Member: %s: %s", aws.StringValue(output.UnprocessedAccounts[0].ErrorCode), aws.StringValue(output.UnprocessedAccounts[0].ErrorMessage)))
			}

			if _, err = waiter.MemberInvited(ctx, conn, d.Id()); err != nil {
				return diag.FromErr(fmt.Errorf("error waiting for Macie Member (%s) invitation: %w", d.Id(), err))
			}
		} else {
			input := &macie2.DisassociateMemberInput{
				Id: aws.String(d.Id()),
			}

			_, err := conn.DisassociateMemberWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
					tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
					return nil
				}
				return diag.FromErr(fmt.Errorf("error disassociating Macie Member invite (%s): %w", d.Id(), err))
			}
		}
	}

	// End Invitation workflow

	if d.HasChange("status") {
		input := &macie2.UpdateMemberSessionInput{
			Id:     aws.String(d.Id()),
			Status: aws.String(d.Get("status").(string)),
		}

		_, err := conn.UpdateMemberSessionWithContext(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating Macie Member (%s): %w", d.Id(), err))
		}

	}

	return resourceMacie2MemberRead(ctx, d, meta)
}

func resourceMacie2MemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteMemberInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteMemberWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeConflictException, "member accounts are associated with your account") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "account is not associated with your account") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie Member (%s): %w", d.Id(), err))
	}
	return nil
}
