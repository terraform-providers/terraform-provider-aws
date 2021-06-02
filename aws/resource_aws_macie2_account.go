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
)

func resourceAwsMacie2Account() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2AccountCreate,
		ReadWithoutTimeout:   resourceMacie2AccountRead,
		UpdateWithoutTimeout: resourceMacie2AccountUpdate,
		DeleteWithoutTimeout: resourceMacie2AccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"finding_publishing_frequency": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(macie2.FindingPublishingFrequency_Values(), false),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(macie2.MacieStatus_Values(), false),
			},
			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2AccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.EnableMacieInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = aws.String(v.(string))
	}
	if v, ok := d.GetOk("status"); ok {
		input.Status = aws.String(v.(string))
	}

	var err error
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err = conn.EnableMacieWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.EnableMacieWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error enabling Macie Account: %w", err))
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceMacie2AccountRead(ctx, d, meta)
}

func resourceMacie2AccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.GetMacieSessionInput{}

	resp, err := conn.GetMacieSessionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
		log.Printf("[WARN] Macie not enabled for AWS account (%s), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Macie Account (%s): %w", d.Id(), err))
	}

	d.Set("status", resp.Status)
	d.Set("finding_publishing_frequency", resp.FindingPublishingFrequency)
	d.Set("service_role", resp.ServiceRole)
	d.Set("created_at", aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set("updated_at", aws.TimeValue(resp.UpdatedAt).Format(time.RFC3339))

	return nil
}

func resourceMacie2AccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.UpdateMacieSessionInput{}

	if d.HasChange("finding_publishing_frequency") {
		input.FindingPublishingFrequency = aws.String(d.Get("finding_publishing_frequency").(string))
	}

	if d.HasChange("status") {
		input.Status = aws.String(d.Get("status").(string))
	}

	_, err := conn.UpdateMacieSessionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Macie Account (%s): %w", d.Id(), err))
	}

	return resourceMacie2AccountRead(ctx, d, meta)
}

func resourceMacie2AccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DisableMacieInput{}

	err := resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err := conn.DisableMacieWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, macie2.ErrCodeConflictException, "Cannot disable Macie while associated with an administrator account") {
			return resource.RetryableError(err)
		}

		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
				tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DisableMacieWithContext(ctx, input)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error disabling Macie Account (%s): %w", d.Id(), err))
	}

	return nil
}
